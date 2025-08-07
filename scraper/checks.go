// Package scraper provides a web scraper that can download a website and its assets.
package scraper

import (
	"net/url"
	"strings"

	"github.com/cornelk/gotokit/log"
)

// normalizeURLPath removes trailing slashes from URL paths for duplicate detection.
// This treats URLs with and without trailing slashes as the same resource.
func normalizeURLPath(path string) string {
	if path == "" {
		return "/"
	}
	// Keep root path as is, but remove trailing slashes from other paths
	if path != "/" && strings.HasSuffix(path, "/") {
		return strings.TrimSuffix(path, "/")
	}
	return path
}

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

	// Normalize the path for duplicate detection to handle trailing slashes
	normalizedPath := normalizeURLPath(p)

	if s.processed.Contains(normalizedPath) { // was already downloaded or checked?
		if url.Fragment != "" {
			return false
		}
		return false
	}

	s.processed.Add(normalizedPath)

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
