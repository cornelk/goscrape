// Package htmlindex provides an index for all HTML tags of relevance for scraping.
package htmlindex

import (
	"fmt"
	"net/url"
	"slices"
	"sort"
	"strings"

	"github.com/cornelk/goscrape/css"
	"github.com/cornelk/gotokit/log"
	"golang.org/x/net/html"
)

// Index provides an index for all HTML tags of relevance for scraping.
type Index struct {
	logger *log.Logger

	// key is HTML tag, value is a map of all its urls and the HTML nodes for it
	data map[string]map[string][]*html.Node
}

// New returns a new index.
func New(logger *log.Logger) *Index {
	return &Index{
		logger: logger,
		data:   make(map[string]map[string][]*html.Node),
	}
}

// Index the given HTML document.
func (idx *Index) Index(baseURL *url.URL, node *html.Node) {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		switch child.Type {
		case html.ElementNode:
			idx.indexElementNode(baseURL, node, child)
		default:
		}
	}
}

func (idx *Index) indexElementNode(baseURL *url.URL, node, child *html.Node) {
	var references []string

	info, ok := Nodes[child.Data]
	if ok {
		references = idx.nodeAttributeURLs(baseURL, child, info.parser, info.Attributes...)
	}

	m, ok := idx.data[child.Data]
	if !ok {
		m = map[string][]*html.Node{}
		idx.data[child.Data] = m
	}

	for _, reference := range references {
		m[reference] = append(m[reference], child)
	}

	if node.FirstChild != nil && !info.noChildParsing {
		idx.Index(baseURL, child)
	}
}

// URLs returns all URLs of the references found for a specific tag.
func (idx *Index) URLs(tag string) ([]*url.URL, error) {
	m, ok := idx.data[tag]
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
func (idx *Index) Nodes(tag string) map[string][]*html.Node {
	m, ok := idx.data[tag]
	if ok {
		return m
	}
	return map[string][]*html.Node{}
}

// nodeAttributeURLs returns resolved URLs based on the base URL and the HTML node attribute values.
func (idx *Index) nodeAttributeURLs(baseURL *url.URL, node *html.Node,
	parser nodeAttributeParser, attributeNames ...string) []string {

	var results []string

	processReferences := func(references []string) {
		for _, reference := range references {
			ur, err := url.Parse(reference)
			if err != nil {
				continue
			}

			ur = baseURL.ResolveReference(ur)
			results = append(results, ur.String())
		}
	}

	for _, attr := range node.Attr {
		if !slices.Contains(attributeNames, attr.Key) {
			continue
		}

		var references []string
		var parserHandled bool

		if parser != nil {
			data := nodeAttributeParserData{
				logger:    idx.logger,
				url:       baseURL,
				node:      node,
				attribute: attr.Key,
				value:     strings.TrimSpace(attr.Val),
			}
			references, parserHandled = parser(data)
		}
		if parser == nil || !parserHandled {
			references = append(references, strings.TrimSpace(attr.Val))
		}

		processReferences(references)
	}

	// special case to support style tag
	if len(attributeNames) == 0 && parser != nil {
		data := nodeAttributeParserData{
			logger: idx.logger,
			url:    baseURL,
			node:   node,
		}
		references, _ := parser(data)
		processReferences(references)
	}

	return results
}

// srcSetValueSplitter returns the URL values of the srcset attribute of img nodes.
func srcSetValueSplitter(data nodeAttributeParserData) ([]string, bool) {
	if !SrcSetAttributes.Contains(data.attribute) {
		return nil, false
	}

	// split the set of responsive images
	values := strings.Split(data.value, ",")

	for i, value := range values {
		value = strings.TrimSpace(value)
		// remove the width in pixels after the url
		values[i], _, _ = strings.Cut(value, " ")
	}

	return values, true
}

// styleParser returns the URL values of a CSS style tag.
func styleParser(data nodeAttributeParserData) ([]string, bool) {
	if data.node.FirstChild == nil {
		return nil, false
	}

	var urls []string
	processor := func(_ *css.Token, _ string, url *url.URL) {
		urls = append(urls, url.String())
	}

	cssData := data.node.FirstChild.Data
	css.Process(data.logger, data.url, cssData, processor)

	return urls, true
}

// styleAttributeParser returns the URL values from CSS content in a style attribute.
func styleAttributeParser(data nodeAttributeParserData) ([]string, bool) {
	if data.attribute != "style" {
		return nil, false
	}

	if data.value == "" {
		return nil, true
	}

	var urls []string
	processor := func(_ *css.Token, _ string, url *url.URL) {
		urls = append(urls, url.String())
	}

	css.Process(data.logger, data.url, data.value, processor)

	return urls, true
}

// combineAttributeParsers combines multiple attribute parsers into a single parser.
func combineAttributeParsers(parsers ...nodeAttributeParser) nodeAttributeParser {
	return func(data nodeAttributeParserData) ([]string, bool) {
		var allUrls []string
		var handled bool

		for _, parser := range parsers {
			if parser == nil {
				continue
			}

			urls, parserHandled := parser(data)
			if parserHandled {
				allUrls = append(allUrls, urls...)
				handled = true
			}
		}

		return allUrls, handled
	}
}
