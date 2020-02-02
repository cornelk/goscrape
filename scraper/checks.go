package scraper

import (
	"net/url"

	"go.uber.org/zap"
)

// checkPageURL checks if a page should be downloaded
func (s *Scraper) checkPageURL(url *url.URL, currentDepth uint) bool {
	if url.Scheme != "http" && url.Scheme != "https" {
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

	if _, ok := s.processed[p]; ok { // was already downloaded or checked
		if url.Fragment != "" {
			return false
		}
		s.log.Debug("Skipping already checked page", zap.Stringer("URL", url))
		return false
	}

	s.processed[p] = struct{}{}
	if s.config.MaxDepth != 0 && currentDepth == s.config.MaxDepth {
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
