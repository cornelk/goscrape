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
	"github.com/uber-go/zap"
)

type (
	// Scraper contains all scraping data
	Scraper struct {
		ImageQuality uint
		MaxDepth     uint
		URL          *url.URL

		browser  *browser.Browser
		excludes []*regexp.Regexp
		log      zap.Logger

		assets         map[string]bool
		assetsExternal map[string]bool
		pages          map[string]bool
	}
)

// New creates a new Scraper instance
func New(URL string) (*Scraper, error) {
	u, err := url.Parse(URL)
	if err != nil {
		return nil, err
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

// SetExcludes sets and checks the exclusions regular expressions
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
	return s.scrapeURL(s.URL, 0)
}

func (s *Scraper) scrapeURL(URL *url.URL, currentDepth uint) error {
	s.log.Info("Downloading", zap.Stringer("URL", URL))
	err := s.browser.Open(URL.String())
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	_, err = s.browser.Download(buf)
	if err != nil {
		return err
	}

	html, err := s.fixFileReferences(URL, buf)
	if err != nil {
		return err
	}

	buf = bytes.NewBufferString(html)
	filePath := s.getFilePath(URL)
	err = s.writeFile(filePath, buf) // always update html files, content might have changed
	if err != nil {
		return err
	}

	for _, stylesheet := range s.browser.Stylesheets() {
		err = s.downloadAssetURL(&stylesheet.DownloadableAsset)
		if err != nil {
			return nil
		}
	}
	for _, script := range s.browser.Scripts() {
		err = s.downloadAssetURL(&script.DownloadableAsset)
		if err != nil {
			return nil
		}
	}
	for _, image := range s.browser.Images() {
		err = s.downloadAssetURL(&image.DownloadableAsset)
		if err != nil {
			return nil
		}
	}

	for _, link := range s.browser.Links() {
		err = s.checkPageURL(link.URL, currentDepth)
		if err != nil {
			return nil
		}
	}

	return nil
}

func (s *Scraper) checkPageURL(URL *url.URL, currentDepth uint) error {
	if URL.Host != s.URL.Host {
		return nil
	}
	if URL.Path == "" || URL.Path == "/" {
		return nil
	}

	_, ok := s.pages[URL.Path]
	if ok { // was already downloaded or checked
		return nil
	}

	s.pages[URL.Path] = false
	if s.MaxDepth != 0 && currentDepth == s.MaxDepth {
		return nil
	}

	if s.isURLExcluded(URL) {
		return nil
	}

	return s.scrapeURL(URL, currentDepth+1)
}

// downloadAssetURL downloads an asset if it does not exist on disk yet.
func (s *Scraper) downloadAssetURL(asset *browser.DownloadableAsset) error {
	URL := asset.URL

	if URL.Host == s.URL.Host {
		_, ok := s.assets[URL.Path]
		if ok { // was already downloaded or checked
			return nil
		}

		s.assets[URL.Path] = false
	} else {
		if s.isExternalFileChecked(URL) {
			return nil
		}
	}

	if s.isURLExcluded(URL) {
		return nil
	}

	filePath := s.getFilePath(URL)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return nil
	}

	s.log.Info("Downloading", zap.Stringer("URL", URL))

	buf := &bytes.Buffer{}
	_, err := asset.Download(buf)
	if err != nil {
		return err
	}

	buf = s.checkFileTypeForRecode(filePath, buf)

	return s.writeFile(filePath, buf)
}

func (s *Scraper) isURLExcluded(URL *url.URL) bool {
	for _, re := range s.excludes {
		if re.MatchString(URL.Path) {
			s.log.Info("Skipping URL", zap.Stringer("URL", URL), zap.Stringer("Excluder", re))
			return true
		}
	}
	return false
}

func (s *Scraper) isExternalFileChecked(URL *url.URL) bool {
	if URL.Host == s.URL.Host {
		return false
	}

	fullURL := URL.String()
	_, ok := s.assetsExternal[fullURL]
	if ok { // was already downloaded or checked
		return true
	}

	s.assetsExternal[fullURL] = true
	s.log.Info("External URL", zap.Stringer("URL", URL))

	return false
}
