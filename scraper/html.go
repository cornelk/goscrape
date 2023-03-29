package scraper

import (
	"io"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/cornelk/gotokit/log"
)

func (s *Scraper) fixFileReferences(url *url.URL, buf io.Reader) (string, error) {
	g, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		return "", err
	}

	relativeToRoot := s.urlRelativeToRoot(url)

	g.Find("a").Each(func(_ int, selection *goquery.Selection) {
		s.fixQuerySelection(url, "href", selection, true, relativeToRoot)
	})

	g.Find("link").Each(func(_ int, selection *goquery.Selection) {
		s.fixQuerySelection(url, "href", selection, false, relativeToRoot)
	})

	g.Find("img").Each(func(_ int, selection *goquery.Selection) {
		s.fixQuerySelection(url, "src", selection, false, relativeToRoot)
	})

	g.Find("script").Each(func(_ int, selection *goquery.Selection) {
		s.fixQuerySelection(url, "src", selection, false, relativeToRoot)
	})

	return g.Html()
}

func (s *Scraper) fixQuerySelection(url *url.URL, attribute string, selection *goquery.Selection,
	isHyperlink bool, relativeToRoot string) {

	src, ok := selection.Attr(attribute)
	if !ok {
		return
	}
	if strings.HasPrefix(src, "#") || strings.HasPrefix(src, "/#") { // anchor
		return
	}

	if strings.HasPrefix(src, "data:") {
		return
	}
	if strings.HasPrefix(src, "mailto:") {
		return
	}

	resolved := s.resolveURL(url, src, isHyperlink, relativeToRoot)
	if src == resolved { // nothing changed
		return
	}

	s.log.Debug("HTML Element relinked",
		log.String("url", src),
		log.String("fixed_url", resolved))
	selection.SetAttr(attribute, resolved)
}
