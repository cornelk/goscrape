package scraper

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"time"

	"github.com/cornelk/goscrape/htmlindex"
	"github.com/cornelk/gotokit/log"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
	"golang.org/x/net/html"
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

	Cookies   []Cookie
	Header    http.Header
	Proxy     string
	UserAgent string
}

type (
	httpDownloader     func(ctx context.Context, u *url.URL) ([]byte, *url.URL, error)
	dirCreator         func(path string) error
	fileExistenceCheck func(filePath string) bool
	fileWriter         func(filePath string, data []byte) error
)

// Scraper contains all scraping data.
type Scraper struct {
	config  Config
	cookies *cookiejar.Jar
	logger  *log.Logger
	URL     *url.URL // contains the main URL to parse, will be modified in case of a redirect

	auth   string
	client *http.Client

	includes []*regexp.Regexp
	excludes []*regexp.Regexp

	// key is the URL of page or asset
	processed map[string]struct{}

	imagesQueue       []*url.URL
	webPageQueue      []*url.URL
	webPageQueueDepth map[string]uint

	dirCreator         dirCreator
	fileExistenceCheck fileExistenceCheck
	fileWriter         fileWriter
	httpDownloader     httpDownloader
}

// New creates a new Scraper instance.
// nolint: funlen
func New(logger *log.Logger, cfg Config) (*Scraper, error) {
	var errs []error

	u, err := url.Parse(cfg.URL)
	if err != nil {
		errs = append(errs, err)
	}
	u.Fragment = ""

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

	cookies, err := createCookieJar(u, cfg.Cookies)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar:     cookies,
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}

	if cfg.Proxy != "" {
		dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("creating proxy from URL: %w", err)
		}

		dialerCtx, ok := dialer.(proxy.ContextDialer)
		if !ok {
			return nil, errors.New("proxy dialer is not a context dialer")
		}

		client.Transport = &http.Transport{
			DialContext: dialerCtx.DialContext,
		}
	}

	s := &Scraper{
		config:  cfg,
		cookies: cookies,
		logger:  logger,
		URL:     u,

		client: client,

		includes: includes,
		excludes: excludes,

		processed: map[string]struct{}{},

		webPageQueueDepth: map[string]uint{},
	}

	s.dirCreator = s.createDownloadPath
	s.fileExistenceCheck = s.fileExists
	s.fileWriter = s.writeFile
	s.httpDownloader = s.downloadURL

	if s.config.Username != "" {
		s.auth = "Basic " + base64.StdEncoding.EncodeToString([]byte(s.config.Username+":"+s.config.Password))
	}

	return s, nil
}

// Start starts the scraping.
func (s *Scraper) Start(ctx context.Context) error {
	if err := s.dirCreator(s.config.OutputDirectory); err != nil {
		return err
	}

	if !s.shouldURLBeDownloaded(s.URL, 0, false) {
		return errors.New("start page is excluded from downloading")
	}

	if err := s.processURL(ctx, s.URL, 0); err != nil {
		return err
	}

	for len(s.webPageQueue) > 0 {
		ur := s.webPageQueue[0]
		s.webPageQueue = s.webPageQueue[1:]
		currentDepth := s.webPageQueueDepth[ur.String()]
		if err := s.processURL(ctx, ur, currentDepth+1); err != nil && errors.Is(err, context.Canceled) {
			return err
		}
	}

	return nil
}

func (s *Scraper) processURL(ctx context.Context, u *url.URL, currentDepth uint) error {
	s.logger.Info("Downloading webpage", log.String("url", u.String()))
	data, respURL, err := s.httpDownloader(ctx, u)
	if err != nil {
		s.logger.Error("Processing HTTP Request failed",
			log.String("url", u.String()),
			log.Err(err))
		return err
	}

	fileExtension := ""
	kind, err := filetype.Match(data)
	if err == nil && kind != types.Unknown {
		fileExtension = kind.Extension
	}

	if currentDepth == 0 {
		u = respURL
		// use the URL that the website returned as new base url for the
		// scrape, in case of a redirect it changed
		s.URL = u
	}

	buf := bytes.NewBuffer(data)
	doc, err := html.Parse(buf)
	if err != nil {
		s.logger.Error("Parsing HTML failed",
			log.String("url", u.String()),
			log.Err(err))
		return fmt.Errorf("parsing HTML: %w", err)
	}

	index := htmlindex.New()
	index.Index(u, doc)

	s.storeDownload(u, data, doc, index, fileExtension)

	if err := s.downloadReferences(ctx, index); err != nil {
		return err
	}

	// check first and download afterward to not hit max depth limit for
	// start page links because of recursive linking
	// a hrefs
	references, err := index.URLs(htmlindex.ATag)
	if err != nil {
		s.logger.Error("Parsing URL failed", log.Err(err))
	}

	for _, ur := range references {
		ur.Fragment = ""

		if s.shouldURLBeDownloaded(ur, currentDepth, false) {
			s.webPageQueue = append(s.webPageQueue, ur)
			s.webPageQueueDepth[ur.String()] = currentDepth
		}
	}

	return nil
}

// storeDownload writes the download to a file, if a known binary file is detected,
// processing of the file as page to look for links is skipped.
func (s *Scraper) storeDownload(u *url.URL, data []byte, doc *html.Node,
	index *htmlindex.Index, fileExtension string) {

	isAPage := false
	if fileExtension == "" {
		fixed, hasChanges, err := s.fixURLReferences(u, doc, index)
		if err != nil {
			s.logger.Error("Fixing file references failed",
				log.String("url", u.String()),
				log.Err(err))
			return
		}

		if hasChanges {
			data = fixed
		}
		isAPage = true
	}

	filePath := s.getFilePath(u, isAPage)
	// always update html files, content might have changed
	if err := s.fileWriter(filePath, data); err != nil {
		s.logger.Error("Writing to file failed",
			log.String("URL", u.String()),
			log.String("file", filePath),
			log.Err(err))
	}
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
