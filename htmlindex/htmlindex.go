package htmlindex

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"golang.org/x/net/html"
)

// Index provides an index for all HTML tags of relevance for scraping.
type Index struct {
	// key is HTML tag, value is a map of all its urls and the HTML nodes for it
	data map[string]map[string][]*html.Node
}

// New returns a new index.
func New() *Index {
	return &Index{
		data: make(map[string]map[string][]*html.Node),
	}
}

// Index the given HTML document.
func (h *Index) Index(baseURL *url.URL, node *html.Node) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type != html.ElementNode {
			continue
		}

		var references []string

		info, ok := Nodes[child.Data]
		if ok {
			references = nodeAttributeURLs(baseURL, child, info.parser, info.Attributes...)
		}

		m, ok := h.data[child.Data]
		if !ok {
			m = map[string][]*html.Node{}
			h.data[child.Data] = m
		}

		for _, reference := range references {
			m[reference] = append(m[reference], child)
		}

		if node.FirstChild != nil {
			h.Index(baseURL, child)
		}
	}
}

// URLs returns all URLs of the references found for a specific tag.
func (h *Index) URLs(tag string) ([]*url.URL, error) {
	m, ok := h.data[tag]
	if !ok {
		return nil, nil
	}

	data := make([]string, 0, len(m))
	for key := range m {
		data = append(data, key)
	}
	sort.Strings(data)

	urls := make([]*url.URL, 0, len(m))
	for _, fullURL := range data {
		u, err := url.Parse(fullURL)
		if err != nil {
			return nil, fmt.Errorf("parsing URL '%s': %w", fullURL, err)
		}
		urls = append(urls, u)
	}

	return urls, nil
}

// Nodes returns a map of all URLs and their HTML nodes.
func (h *Index) Nodes(tag string) map[string][]*html.Node {
	m, ok := h.data[tag]
	if ok {
		return m
	}
	return map[string][]*html.Node{}
}

// nodeAttributeURLs returns resolved URLs based on the base URL and the HTML node attribute values.
func nodeAttributeURLs(baseURL *url.URL, node *html.Node,
	parser nodeAttributeParser, attributeName ...string) []string {

	var results []string

	for _, attr := range node.Attr {
		var process bool
		for _, name := range attributeName {
			if attr.Key == name {
				process = true
				break
			}
		}
		if !process {
			continue
		}

		var references []string
		var parserHandled bool

		if parser != nil {
			references, parserHandled = parser(attr.Key, strings.TrimSpace(attr.Val))
		}
		if parser == nil || !parserHandled {
			references = append(references, strings.TrimSpace(attr.Val))
		}

		for _, reference := range references {
			ur, err := url.Parse(reference)
			if err != nil {
				continue
			}

			ur = baseURL.ResolveReference(ur)
			results = append(results, ur.String())
		}
	}

	return results
}

// srcSetValueSplitter returns the URL values of the srcset attribute of img nodes.
func srcSetValueSplitter(attribute, attributeValue string) ([]string, bool) {
	if _, isSrcSet := SrcSetAttributes[attribute]; !isSrcSet {
		return nil, false
	}

	// split the set of responsive images
	values := strings.Split(attributeValue, ",")

	for i, value := range values {
		value = strings.TrimSpace(value)
		// remove the width in pixels after the url
		values[i], _, _ = strings.Cut(value, " ")
	}

	return values, true
}
