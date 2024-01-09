package scraper

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/cornelk/goscrape/htmlindex"
	"github.com/cornelk/gotokit/log"
	"golang.org/x/net/html"
)

// ignoredURLPrefixes contains a list of URL prefixes that do not need to bo adjusted.
var ignoredURLPrefixes = []string{
	"#",       // anchor
	"/#",      // anchor
	"data:",   // embedded data
	"mailto:", // mail address
}

// fixURLReferences fixes URL references to point to relative file names.
// It returns a bool that indicates that no reference needed to be fixed,
// in this case the returned HTML string will be empty.
func (s *Scraper) fixURLReferences(url *url.URL, doc *html.Node,
	index *htmlindex.Index) (string, bool, error) {

	relativeToRoot := urlRelativeToRoot(url)
	if !s.fixHTMLNodeURLs(url, relativeToRoot, index) {
		return "", false, nil
	}

	var rendered bytes.Buffer
	if err := html.Render(&rendered, doc); err != nil {
		return "", false, fmt.Errorf("rendering html: %w", err)
	}
	return rendered.String(), true, nil
}

// fixHTMLNodeURLs processes all HTML nodes that contain URLs that need to be fixed
// to link to downloaded files. It returns whether any URLS have been fixed.
func (s *Scraper) fixHTMLNodeURLs(baseURL *url.URL, relativeToRoot string, index *htmlindex.Index) bool {
	changed := false

	urls := index.Nodes("a")
	for _, nodes := range urls {
		for _, node := range nodes {
			if s.fixNodeURL(baseURL, "href", node, true, relativeToRoot) {
				changed = true
			}
		}
	}

	urls = index.Nodes("link")
	for _, nodes := range urls {
		for _, node := range nodes {
			if s.fixNodeURL(baseURL, "href", node, false, relativeToRoot) {
				changed = true
			}
		}
	}

	urls = index.Nodes("img")
	for _, nodes := range urls {
		for _, node := range nodes {
			if s.fixNodeURL(baseURL, "src", node, false, relativeToRoot) {
				changed = true
			}
		}
	}

	urls = index.Nodes("script")
	for _, nodes := range urls {
		for _, node := range nodes {
			if s.fixNodeURL(baseURL, "src", node, false, relativeToRoot) {
				changed = true
			}
		}
	}

	return changed
}

// fixURLReferences fixes the URL references of a HTML node to point to a relative file name.
// It returns whether the URL bas been adjusted.
func (s *Scraper) fixNodeURL(baseURL *url.URL, attributeName string, node *html.Node,
	isHyperlink bool, relativeToRoot string) bool {

	var nodeURL string
	var attribute *html.Attribute
	for i, attr := range node.Attr {
		if attr.Key != attributeName {
			continue
		}

		attribute = &node.Attr[i]
		nodeURL = strings.TrimSpace(attr.Val)
		if nodeURL == "" {
			return false
		}
		break
	}
	if attribute == nil {
		return false
	}

	for _, prefix := range ignoredURLPrefixes {
		if strings.HasPrefix(nodeURL, prefix) {
			return false
		}
	}

	resolved := resolveURL(baseURL, nodeURL, s.URL.Host, isHyperlink, relativeToRoot)
	if nodeURL == resolved { // no change
		return false
	}

	s.logger.Debug("HTML Element relinked",
		log.String("url", nodeURL),
		log.String("fixed_url", resolved))
	attribute.Val = resolved
	return true
}
