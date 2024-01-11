package scraper

import (
	"net/url"

	"github.com/cornelk/gotokit/log"
)

// shouldURLBeDownloaded checks whether a page should be downloaded.
// nolint: cyclop
func (s *Scraper) shouldURLBeDownloaded(url *url.URL, currentDepth uint, isAsset bool) bool {
	if url.Scheme != "http" && url.Scheme != "https" {
		return false
	}

	p := url.String()
	if url.Host == s.URL.Host {
		p = url.Path
	}
	if p == "" {
		p = "/"
	}

	if _, ok := s.processed[p]; ok { // was already downloaded or checked?
		if url.Fragment != "" {
			return false
		}
		return false
	}

	s.processed[p] = struct{}{}

	if !isAsset {
		if url.Host != s.URL.Host {
			s.logger.Debug("Skipping external host page", log.String("url", url.String()))
			return false
		}

		if s.config.MaxDepth != 0 && currentDepth == s.config.MaxDepth {
			s.logger.Debug("Skipping too deep level page", log.String("url", url.String()))
			return false
		}
	}

	if s.includes != nil && !s.isURLIncluded(url) {
		return false
	}
	if s.excludes != nil && s.isURLExcluded(url) {
		return false
	}

	s.logger.Debug("New URL to download", log.String("url", url.String()))
	return true
}

func (s *Scraper) isURLIncluded(url *url.URL) bool {
	for _, re := range s.includes {
		if re.MatchString(url.Path) {
			s.logger.Info("Including URL",
				log.String("url", url.String()),
				log.Stringer("included_expression", re))
			return true
		}
	}
	return false
}

func (s *Scraper) isURLExcluded(url *url.URL) bool {
	for _, re := range s.excludes {
		if re.MatchString(url.Path) {
			s.logger.Info("Skipping URL",
				log.String("url", url.String()),
				log.Stringer("excluded_expression", re))
			return true
		}
	}
	return false
}
