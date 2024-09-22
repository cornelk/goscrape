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
	"#",       // fragment
	"/#",      // fragment
	"data:",   // embedded data
	"mailto:", // mail address
}

// fixURLReferences fixes URL references to point to relative file names.
// It returns a bool that indicates that no reference needed to be fixed,
// in this case the returned HTML string will be empty.
func (s *Scraper) fixURLReferences(url *url.URL, doc *html.Node,
	index *htmlindex.Index) ([]byte, bool, error) {

	relativeToRoot := urlRelativeToRoot(url)
	if !s.fixHTMLNodeURLs(url, relativeToRoot, index) {
		return nil, false, nil
	}

	var rendered bytes.Buffer
	if err := html.Render(&rendered, doc); err != nil {
		return nil, false, fmt.Errorf("rendering html: %w", err)
	}
	return rendered.Bytes(), true, nil
}

// fixHTMLNodeURLs processes all HTML nodes that contain URLs that need to be fixed
// to link to downloaded files. It returns whether any URLS have been fixed.
func (s *Scraper) fixHTMLNodeURLs(baseURL *url.URL, relativeToRoot string, index *htmlindex.Index) bool {
	var changed bool

	for tag, nodeInfo := range htmlindex.Nodes {
		isHyperlink := tag == htmlindex.ATag

		urls := index.Nodes(tag)
		for _, nodes := range urls {
			for _, node := range nodes {
				if s.fixNodeURL(baseURL, nodeInfo.Attributes, node, isHyperlink, relativeToRoot) {
					changed = true
				}
			}
		}
	}

	return changed
}

// fixNodeURL fixes the URL references of a HTML node to point to a relative file name.
// It returns whether any attribute value bas been adjusted.
func (s *Scraper) fixNodeURL(baseURL *url.URL, attributes []string, node *html.Node,
	isHyperlink bool, relativeToRoot string) bool {

	var changed bool

	for i, attr := range node.Attr {
		var process bool
		for _, name := range attributes {
			if attr.Key == name {
				process = true
				break
			}
		}
		if !process {
			continue
		}

		value := strings.TrimSpace(attr.Val)
		if value == "" {
			continue
		}

		for _, prefix := range ignoredURLPrefixes {
			if strings.HasPrefix(value, prefix) {
				return false
			}
		}

		var adjusted string

		if _, isSrcSet := htmlindex.SrcSetAttributes[attr.Key]; isSrcSet {
			adjusted = resolveSrcSetURLs(baseURL, value, s.URL.Host, isHyperlink, relativeToRoot)
		} else {
			adjusted = resolveURL(baseURL, value, s.URL.Host, isHyperlink, relativeToRoot)
		}

		if adjusted == value { // check for no change
			continue
		}

		s.logger.Debug("HTML node relinked",
			log.String("value", value),
			log.String("fixed_value", adjusted))

		attribute := &node.Attr[i]
		attribute.Val = adjusted
		changed = true
	}

	return changed
}

func resolveSrcSetURLs(base *url.URL, srcSetValue, mainPageHost string, isHyperlink bool, relativeToRoot string) string {
	// split the set of responsive images
	values := strings.Split(srcSetValue, ",")

	for i, value := range values {
		value = strings.TrimSpace(value)
		parts := strings.Split(value, " ")
		parts[0] = resolveURL(base, parts[0], mainPageHost, isHyperlink, relativeToRoot)
		values[i] = strings.Join(parts, " ")
	}

	return strings.Join(values, ", ")
}
