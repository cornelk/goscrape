package scraper

import (
	"io"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
)

func (s *Scraper) fixFileReferences(URL *url.URL, buf io.Reader) (string, error) {
	g, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		return "", err
	}

	relativeToRoot := s.urlRelativeToRoot(URL)

	g.Find("a").Each(func(_ int, selection *goquery.Selection) {
		s.fixQuerySelection(URL, "href", selection, true, relativeToRoot)
	})

	g.Find("link").Each(func(_ int, selection *goquery.Selection) {
		s.fixQuerySelection(URL, "href", selection, false, relativeToRoot)
	})

	g.Find("img").Each(func(_ int, selection *goquery.Selection) {
		s.fixQuerySelection(URL, "src", selection, false, relativeToRoot)
	})

	g.Find("script").Each(func(_ int, selection *goquery.Selection) {
		s.fixQuerySelection(URL, "src", selection, false, relativeToRoot)
	})

	return g.Html()
}

func (s *Scraper) fixQuerySelection(URL *url.URL, attribute string, selection *goquery.Selection, linkIsAPage bool, relativeToRoot string) {
	src, ok := selection.Attr(attribute)
	if !ok {
		return
	}

	if strings.HasPrefix(src, "data:") {
		return
	}

	resolved := s.resolveURL(URL, src, linkIsAPage, relativeToRoot)
	if src == resolved { // nothing changed
		return
	}

	s.log.Debug("HTML Element relinked", zap.String("URL", src), zap.String("Fixed", resolved))
	selection.SetAttr(attribute, resolved)
}
