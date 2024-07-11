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

const (
	BackgroundAttribute = "background"
	HrefAttribute       = "href"
	SrcAttribute        = "src"
	SrcSetAttribute     = "srcset"
)

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

		switch child.Data {
		case "a", "link":
			references = nodeAttributeURLs(baseURL, child, HrefAttribute, nil)

		case "script":
			references = nodeAttributeURLs(baseURL, child, SrcAttribute, nil)

		case "img":
			references = nodeAttributeURLs(baseURL, child, SrcAttribute, nil)
			references = append(references, nodeAttributeURLs(baseURL, child, SrcSetAttribute, srcSetValueSplitter)...)

		default:
			// handle body references and all childs
			if child.Data == "body" {
				references = nodeAttributeURLs(baseURL, child, BackgroundAttribute, nil) // TODO: handle srcset attribute)
			}

			if node.FirstChild != nil {
				h.Index(baseURL, child)
			}
		}

		m, ok := h.data[child.Data]
		if !ok {
			m = map[string][]*html.Node{}
			h.data[child.Data] = m
		}

		for _, reference := range references {
			m[reference] = append(m[reference], child)
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

type nodeAttributeParser func(value string) []string

// nodeAttributeURLs returns resolved URLs based on the base URL and the HTML node attribute values.
func nodeAttributeURLs(baseURL *url.URL, node *html.Node,
	attributeName string, parser nodeAttributeParser) []string {

	for _, attr := range node.Attr {
		if attr.Key != attributeName {
			continue
		}

		var references []string
		if parser == nil {
			references = append(references, strings.TrimSpace(attr.Val))
		} else {
			references = parser(strings.TrimSpace(attr.Val))
		}

		var results []string
		for _, reference := range references {
			ur, err := url.Parse(reference)
			if err != nil {
				continue
			}

			ur = baseURL.ResolveReference(ur)
			results = append(results, ur.String())
		}
		return results
	}

	return nil
}

// srcSetValueSplitter returns the URL values of the srcset attribute of img nodes.
func srcSetValueSplitter(attributeValue string) []string {
	// split the set of responsive images
	values := strings.Split(attributeValue, ",")

	for i, value := range values {
		value = strings.TrimSpace(value)
		// remove the width in pixels after the url
		values[i], _, _ = strings.Cut(value, " ")
	}

	return values
}
