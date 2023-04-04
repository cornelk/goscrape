package scraper

import (
	"bytes"
	"encoding/base64"
	"errors"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/cornelk/gotokit/log"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
	"github.com/headzoo/surf"
	"github.com/headzoo/surf/agent"
	"github.com/headzoo/surf/browser"
	"golang.org/x/net/proxy"
)

// Config contains the scraper configuration.
type Config struct {
	URL      string
	Includes []string
	Excludes []string

	ImageQuality uint // image quality from 0 to 100%, 0 to disable reencoding
	MaxDepth     uint // download depth, 0 for unlimited
	Timeout      uint // time limit in seconds to process each http request

	OutputDirectory string
	Username        string
	Password        string

	UserAgent string
	Proxy     string
}

// Scraper contains all scraping data.
type Scraper struct {
	config  Config
	logger  *log.Logger
	URL     *url.URL
	browser *browser.Browser

	cssURLRe *regexp.Regexp
	includes []*regexp.Regexp
	excludes []*regexp.Regexp

	// key is the URL of page or asset
	processed map[string]struct{}

	imagesQueue []*browser.DownloadableAsset
}

// New creates a new Scraper instance.
func New(logger *log.Logger, cfg Config) (*Scraper, error) {
	var errs []error

	u, err := url.Parse(cfg.URL)
	if err != nil {
		errs = append(errs, err)
	}

	includes, err := compileRegexps(cfg.Includes)
	if err != nil {
		errs = append(errs, err)
	}

	excludes, err := compileRegexps(cfg.Excludes)
	if err != nil {
		errs = append(errs, err)
	}

	proxyURL, err := url.Parse(cfg.Proxy)
	if err != nil {
		errs = append(errs, err)
	}

	if errs != nil {
		return nil, errors.Join(errs...)
	}

	if u.Scheme == "" {
		u.Scheme = "http" // if no URL scheme was given default to http
	}

	if cfg.UserAgent == "" {
		cfg.UserAgent = agent.GoogleBot()
	}

	b := surf.NewBrowser()
	b.SetUserAgent(cfg.UserAgent)
	b.SetTimeout(time.Duration(cfg.Timeout) * time.Second)

	if cfg.Proxy != "" {
		dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
		if err != nil {
			return nil, err
		}
		b.SetTransport(&http.Transport{
			Dial: dialer.Dial,
		})
	}

	s := &Scraper{
		config: cfg,

		browser:   b,
		logger:    logger,
		processed: make(map[string]struct{}),
		URL:       u,
		cssURLRe:  regexp.MustCompile(`^url\(['"]?(.*?)['"]?\)$`),
		includes:  includes,
		excludes:  excludes,
	}
	return s, nil
}

// compileRegexps compiles the given regex strings to regular expressions
// to be used in the include and exclude filters.
func compileRegexps(regexps []string) ([]*regexp.Regexp, error) {
	var errs []error
	var compiled []*regexp.Regexp

	for _, exp := range regexps {
		re, err := regexp.Compile(exp)
		if err == nil {
			compiled = append(compiled, re)
		} else {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return compiled, nil
}

// Start starts the scraping.
func (s *Scraper) Start() error {
	if s.config.OutputDirectory != "" {
		if err := os.MkdirAll(s.config.OutputDirectory, os.ModePerm); err != nil {
			return err
		}
	}

	p := s.URL.Path
	if p == "" {
		p = "/"
	}
	s.processed[p] = struct{}{}

	if s.config.Username != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(s.config.Username + ":" + s.config.Password))
		s.browser.AddRequestHeader("Authorization", "Basic "+auth)
	}

	s.downloadURL(s.URL, 0)
	return nil
}

func (s *Scraper) downloadURL(u *url.URL, currentDepth uint) {
	s.logger.Info("Downloading", log.Stringer("URL", u))

	if err := s.browser.Open(u.String()); err != nil {
		s.logger.Error("Request failed",
			log.Stringer("url", u),
			log.Err(err))
		return
	}

	if c := s.browser.StatusCode(); c != http.StatusOK {
		s.logger.Error("Request failed",
			log.Stringer("url", u),
			log.Int("http_status_code", c))
		return
	}

	buf := &bytes.Buffer{}
	if _, err := s.browser.Download(buf); err != nil {
		s.logger.Error("Downloading content failed",
			log.Stringer("url", u),
			log.Err(err))
		return
	}

	fileExtension := ""
	kind, err := filetype.Match(buf.Bytes())
	if err == nil && kind != types.Unknown {
		fileExtension = kind.Extension
	}

	if currentDepth == 0 {
		u = s.browser.Url()
		// use the URL that the website returned as new base url for the
		// scrape, in case of a redirect it changed
		s.URL = u
	}

	s.storeDownload(u, buf, fileExtension)

	s.downloadReferences()

	var toScrape []*url.URL
	// check first and download afterwards to not hit max depth limit for
	// start page links because of recursive linking
	for _, link := range s.browser.Links() {
		if s.shouldPageBeDownloaded(link.URL, currentDepth) {
			toScrape = append(toScrape, link.URL)
		}
	}

	for _, URL := range toScrape {
		s.downloadURL(URL, currentDepth+1)
	}
}

// storeDownload writes the download to a file, if a known binary file is detected, processing of the file as
// page to look for links is skipped.
func (s *Scraper) storeDownload(u *url.URL, buf *bytes.Buffer, fileExtension string) {
	isAPage := false
	if fileExtension == "" {
		html, fixed, err := s.fixURLReferences(u, buf)
		if err != nil {
			s.logger.Error("Fixing file references failed",
				log.Stringer("url", u),
				log.Err(err))
			return
		}

		if fixed {
			buf = bytes.NewBufferString(html)
		}
		isAPage = true
	}

	filePath := s.GetFilePath(u, isAPage)
	// always update html files, content might have changed
	if err := s.writeFile(filePath, buf); err != nil {
		s.logger.Error("Writing to file failed",
			log.Stringer("URL", u),
			log.String("file", filePath),
			log.Err(err))
	}
}
