package scraper

import (
	"io"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/uber-go/zap"
)

func (s *Scraper) fixFileReferences(buf io.Reader) (string, error) {
	g, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		return "", err
	}

	g.Find("a").Each(func(_ int, selection *goquery.Selection) {
		s.fixQuerySelection("href", selection)
	})

	g.Find("link").Each(func(_ int, selection *goquery.Selection) {
		s.fixQuerySelection("href", selection)
	})

	g.Find("img").Each(func(_ int, selection *goquery.Selection) {
		s.fixQuerySelection("src", selection)
	})

	g.Find("script").Each(func(_ int, selection *goquery.Selection) {
		s.fixQuerySelection("src", selection)
	})

	return g.Html()
}

func (s *Scraper) fixQuerySelection(attribute string, selection *goquery.Selection) {
	src, ok := selection.Attr(attribute)
	if !ok {
		return
	}

	ur, err := url.Parse(src)
	if err != nil {
		return
	}
	if ur.Host != s.URL.Host {
		return
	}

	refRes := s.URL.ResolveReference(ur)
	refRes.Scheme = "" // remove http/https
	refRes.Host = ""   // remove host
	refStr := refRes.String()

	if refStr == "" {
		refStr = "/" // website root
	} else if len(refStr) > 1 && refStr[0] == '/' {
		refStr = refStr[1:]
	}
	if refStr[len(refStr)-1] == '/' {
		refStr += "index.html"
	}

	s.log.Debug("HTML Element fixed", zap.Stringer("URL", refRes), zap.String("Fixed", refStr))
	selection.SetAttr(attribute, refStr)
}
