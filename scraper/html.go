package scraper

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/cornelk/gotokit/log"
	"golang.org/x/net/html"
)

// fixURLReferences fixes URL references to point to relative file names.
// It returns a bool that indicates that no reference needed to be fixed, in this case the returned HTML string
// will be empty.
func (s *Scraper) fixURLReferences(url *url.URL, buf *bytes.Buffer) (string, bool, error) {
	relativeToRoot := s.urlRelativeToRoot(url)
	doc, err := html.Parse(buf)
	if err != nil {
		return "", false, fmt.Errorf("parsing html: %w", err)
	}

	if !s.parseHTMLNodeChildren(url, relativeToRoot, doc) {
		return "", false, nil
	}

	var rendered bytes.Buffer
	if err = html.Render(&rendered, doc); err != nil {
		return "", false, fmt.Errorf("rendering html: %w", err)
	}
	return rendered.String(), true, nil
}

// parseHTMLNodeChildren parses all HTML children of a HTML node recursively for nodes of interest that contain
// URLS that need to be fixed to link to downloaded files. It returns whether any URLS have been fixed.
func (s *Scraper) parseHTMLNodeChildren(baseURL *url.URL, relativeToRoot string, node *html.Node) bool {
	changed := false

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode {
			continue
		}

		switch child.Data {
		case "a":
			if s.fixNodeURL(baseURL, "href", child, true, relativeToRoot) {
				changed = true
			}
		case "link":
			if s.fixNodeURL(baseURL, "href", child, false, relativeToRoot) {
				changed = true
			}
		case "img", "script":
			if s.fixNodeURL(baseURL, "src", child, false, relativeToRoot) {
				changed = true
			}

		default:
			if node.FirstChild != nil {
				if s.parseHTMLNodeChildren(baseURL, relativeToRoot, child) {
					changed = true
				}
			}
		}
	}
	return changed
}

// ignoredURLPrefixes contains a list of URL prefixes that do not need to bo adjusted.
var ignoredURLPrefixes = []string{
	"#",       // anchor
	"/#",      // anchor
	"data:",   // embedded data
	"mailto:", // mail address
}

// fixURLReferences fixe the URL references of a HTML node to point to a relative file name.
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

	resolved := s.resolveURL(baseURL, nodeURL, isHyperlink, relativeToRoot)
	if nodeURL == resolved { // no change
		return false
	}

	s.logger.Debug("HTML Element relinked",
		log.String("url", nodeURL),
		log.String("fixed_url", resolved))
	attribute.Val = resolved
	return true
}
