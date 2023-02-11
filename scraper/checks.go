package scraper

import (
	"net/url"

	"github.com/cornelk/gotokit/log"
)

// checkPageURL checks whether a page should be downloaded.
func (s *Scraper) checkPageURL(url *url.URL, currentDepth uint) bool {
	if url.Scheme != "http" && url.Scheme != "https" {
		return false
	}
	if url.Host != s.URL.Host {
		s.log.Debug("Skipping external host page", log.Stringer("URL", url))
		return false
	}

	p := url.Path
	if p == "" {
		p = "/"
	}

	if _, ok := s.processed[p]; ok { // was already downloaded or checked
		if url.Fragment != "" {
			return false
		}
		s.log.Debug("Skipping already checked page", log.Stringer("URL", url))
		return false
	}

	s.processed[p] = struct{}{}
	if s.config.MaxDepth != 0 && currentDepth == s.config.MaxDepth {
		s.log.Debug("Skipping too deep level page", log.Stringer("URL", url))
		return false
	}

	if s.includes != nil && !s.isURLIncluded(url) {
		return false
	}
	if s.excludes != nil && s.isURLExcluded(url) {
		return false
	}

	s.log.Debug("New page to queue", log.Stringer("URL", url))
	return true
}

func (s *Scraper) isURLIncluded(url *url.URL) bool {
	if url.Scheme == "data" {
		return true
	}

	for _, re := range s.includes {
		if re.MatchString(url.Path) {
			s.log.Info("Including URL",
				log.Stringer("URL", url),
				log.Stringer("Included", re))
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
				log.Stringer("URL", url),
				log.Stringer("Excluded", re))
			return true
		}
	}
	return false
}
