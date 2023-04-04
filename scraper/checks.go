package scraper

import (
	"net/url"

	"github.com/cornelk/gotokit/log"
)

// shouldPageBeDownloaded checks whether a page should be downloaded.
func (s *Scraper) shouldPageBeDownloaded(url *url.URL, currentDepth uint) bool {
	if url.Scheme != "http" && url.Scheme != "https" {
		return false
	}
	if url.Host != s.URL.Host {
		s.logger.Debug("Skipping external host page", log.Stringer("url", url))
		return false
	}

	p := url.Path
	if p == "" {
		p = "/"
	}

	if _, ok := s.processed[p]; ok { // was already downloaded or checked?
		if url.Fragment != "" {
			return false
		}
		s.logger.Debug("Skipping already checked page", log.Stringer("url", url))
		return false
	}

	s.processed[p] = struct{}{}
	if s.config.MaxDepth != 0 && currentDepth == s.config.MaxDepth {
		s.logger.Debug("Skipping too deep level page", log.Stringer("url", url))
		return false
	}

	if s.includes != nil && !s.isURLIncluded(url) {
		return false
	}
	if s.excludes != nil && s.isURLExcluded(url) {
		return false
	}

	s.logger.Debug("New page to download", log.Stringer("url", url))
	return true
}

func (s *Scraper) isURLIncluded(url *url.URL) bool {
	if url.Scheme == "data" {
		return true
	}

	for _, re := range s.includes {
		if re.MatchString(url.Path) {
			s.logger.Info("Including URL",
				log.Stringer("url", url),
				log.Stringer("included_expression", re))
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
			s.logger.Info("Skipping URL",
				log.Stringer("url", url),
				log.Stringer("excluded_expression", re))
			return true
		}
	}
	return false
}
