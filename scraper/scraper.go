package scraper

import (
	"bytes"
	"net/url"
	"os"
	"regexp"

	"github.com/cornelk/goscrape/appcontext"
	"github.com/headzoo/surf"
	"github.com/headzoo/surf/agent"
	"github.com/headzoo/surf/browser"
	"go.uber.org/zap"
)

type (
	// Scraper contains all scraping data
	Scraper struct {
		// Configuration
		ImageQuality    uint
		MaxDepth        uint
		OutputDirectory string
		URL             *url.URL

		browser  *browser.Browser
		includes []*regexp.Regexp
		excludes []*regexp.Regexp
		log      *zap.Logger

		assets         map[string]bool
		imagesQueue    []*browser.DownloadableAsset
		assetsExternal map[string]bool
		pages          map[string]bool
	}

	// assetProcessor is a processor of a downloaded asset.
	assetProcessor func(URL *url.URL, buf *bytes.Buffer) *bytes.Buffer
)

// New creates a new Scraper instance
func New(startURL string) (*Scraper, error) {
	u, err := url.Parse(startURL)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" {
		u.Scheme = "http" // if no URL scheme was given default to http
	}

	b := surf.NewBrowser()
	b.SetUserAgent(agent.GoogleBot())

	s := &Scraper{
		browser:        b,
		log:            appcontext.Logger,
		assets:         make(map[string]bool),
		assetsExternal: make(map[string]bool),
		pages:          make(map[string]bool),
		URL:            u,
	}
	return s, nil
}

// SetIncludes sets and checks the inclusion regular expressions.
func (s *Scraper) SetIncludes(includes []string) error {
	for _, e := range includes {
		re, err := regexp.Compile(e)
		if err != nil {
			return err
		}

		s.includes = append(s.includes, re)
		s.log.Debug("Including", zap.Stringer("RE", re))
	}

	return nil
}

// SetExcludes sets and checks the exclusions regular expressions.
func (s *Scraper) SetExcludes(excludes []string) error {
	for _, e := range excludes {
		re, err := regexp.Compile(e)
		if err != nil {
			return err
		}

		s.excludes = append(s.excludes, re)
		s.log.Debug("Excluding", zap.Stringer("RE", re))
	}

	return nil
}

// Start starts the scraping
func (s *Scraper) Start() error {
	if s.OutputDirectory != "" {
		if err := os.MkdirAll(s.OutputDirectory, os.ModePerm); err != nil {
			return err
		}
	}

	p := s.URL.Path
	if p == "" {
		p = "/"
	}
	s.pages[p] = false
	return s.scrapeURL(s.URL, 0)
}

func (s *Scraper) scrapeURL(u *url.URL, currentDepth uint) error {
	s.log.Info("Downloading", zap.Stringer("URL", u))
	if err := s.browser.Open(u.String()); err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if _, err := s.browser.Download(buf); err != nil {
		return err
	}

	if currentDepth == 0 {
		u = s.browser.Url() // use the URL that the website returned as new base url for the scrape, in case of a redirect
		s.URL = u
	}

	html, err := s.fixFileReferences(u, buf)
	if err != nil {
		return err
	}

	buf = bytes.NewBufferString(html)
	filePath := s.GetFilePath(u, true)
	// always update html files, content might have changed
	if err = s.writeFile(filePath, buf); err != nil {
		return err
	}

	if err = s.downloadReferences(); err != nil {
		return err
	}

	var toScrape []*url.URL
	// check first and download afterwards to not hit max depth limit for start page links because of recursive linking
	for _, link := range s.browser.Links() {
		if s.checkPageURL(link.URL, currentDepth) {
			toScrape = append(toScrape, link.URL)
		}
	}

	for _, URL := range toScrape {
		if err = s.scrapeURL(URL, currentDepth+1); err != nil {
			return err
		}
	}
	return nil
}

func (s *Scraper) downloadReferences() error {
	for _, image := range s.browser.Images() {
		s.imagesQueue = append(s.imagesQueue, &image.DownloadableAsset)
	}
	for _, stylesheet := range s.browser.Stylesheets() {
		if err := s.downloadAssetURL(&stylesheet.DownloadableAsset, s.checkCSSForUrls); err != nil {
			return err
		}
	}
	for _, script := range s.browser.Scripts() {
		if err := s.downloadAssetURL(&script.DownloadableAsset, nil); err != nil {
			return err
		}
	}
	for _, image := range s.imagesQueue {
		if err := s.downloadAssetURL(image, s.checkImageForRecode); err != nil {
			return err
		}
	}
	s.imagesQueue = nil
	return nil
}

// checkPageURL checks if a page should be downloaded
func (s *Scraper) checkPageURL(url *url.URL, currentDepth uint) bool {
	if url.Scheme == "mailto" {
		return false
	}
	if url.Host != s.URL.Host {
		s.log.Debug("Skipping external host page", zap.Stringer("URL", url))
		return false
	}

	p := url.Path
	if p == "" {
		p = "/"
	}

	if _, ok := s.pages[p]; ok { // was already downloaded or checked
		if url.Fragment != "" {
			return false
		}
		s.log.Debug("Skipping already checked page", zap.Stringer("URL", url))
		return false
	}

	s.pages[p] = false
	if s.MaxDepth != 0 && currentDepth == s.MaxDepth {
		s.log.Debug("Skipping too deep level page", zap.Stringer("URL", url))
		return false
	}

	if s.includes != nil && !s.isURLIncluded(url) {
		return false
	}
	if s.excludes != nil && s.isURLExcluded(url) {
		return false
	}

	s.log.Debug("New page to queue", zap.Stringer("URL", url))
	return true
}

// downloadAssetURL downloads an asset if it does not exist on disk yet.
func (s *Scraper) downloadAssetURL(asset *browser.DownloadableAsset, processor assetProcessor) error {
	URL := asset.URL

	if URL.Host == s.URL.Host {
		if _, ok := s.assets[URL.Path]; ok { // was already downloaded or checked
			return nil
		}

		s.assets[URL.Path] = false
	} else if s.isExternalFileChecked(URL) {
		return nil
	}

	if s.includes != nil && !s.isURLIncluded(URL) {
		return nil
	}
	if s.excludes != nil && s.isURLExcluded(URL) {
		return nil
	}

	filePath := s.GetFilePath(URL, false)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return nil
	}

	s.log.Info("Downloading", zap.Stringer("URL", URL))

	buf := &bytes.Buffer{}
	_, err := asset.Download(buf)
	if err != nil {
		return err
	}

	if processor != nil {
		buf = processor(URL, buf)
	}

	return s.writeFile(filePath, buf)
}

func (s *Scraper) isURLIncluded(url *url.URL) bool {
	if url.Scheme == "data" {
		return true
	}

	for _, re := range s.includes {
		if re.MatchString(url.Path) {
			s.log.Info("Including URL",
				zap.Stringer("URL", url),
				zap.Stringer("Included", re))
			return true
		}
	}
	return false
}

func (s *Scraper) isURLExcluded(url *url.URL) bool {
	if url.Scheme == "data" {
		return true
	}

	for _, re := range s.excludes {
		if re.MatchString(url.Path) {
			s.log.Info("Skipping URL",
				zap.Stringer("URL", url),
				zap.Stringer("Excluded", re))
			return true
		}
	}
	return false
}

func (s *Scraper) isExternalFileChecked(url *url.URL) bool {
	if url.Host == s.URL.Host {
		return false
	}

	fullURL := url.String()
	if _, ok := s.assetsExternal[fullURL]; ok { // was already downloaded or checked
		return true
	}

	s.assetsExternal[fullURL] = true
	s.log.Info("External URL", zap.Stringer("URL", url))

	return false
}
