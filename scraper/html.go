package scraper

import (
	"io"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/uber-go/zap"
)

func (s *Scraper) fixFileReferences(URL *url.URL, buf io.Reader) (string, error) {
	var relativeToRoot string
	splits := strings.Split(URL.Path, "/")
	for i := range splits {
		if len(splits[i]) > 0 {
			relativeToRoot += "../"
		}
	}

	g, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		return "", err
	}

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

	ur, err := url.Parse(src)
	if err != nil {
		return
	}

	var refRes *url.URL
	if ur.Host != "" && ur.Host != s.URL.Host {
		refRes = URL.ResolveReference(ur)
		refRes.Path = filepath.Join("_"+ur.Host, refRes.Path)
	} else {
		refRes = URL.ResolveReference(ur)
	}

	refRes.Host = ""   // remove host
	refRes.Scheme = "" // remove http/https
	refStr := refRes.String()

	if refStr == "" {
		refStr = "/" // website root
	} else {
		if refStr[0] == '/' && len(relativeToRoot) > 0 {
			refStr = relativeToRoot + refStr[1:]
		} else {
			refStr = relativeToRoot + refStr
		}
	}

	if linkIsAPage && refStr[len(refStr)-1] == '/' {
		refStr += "index.html"
	}
	refStr = strings.TrimPrefix(refStr, "/")

	if src == refStr { // nothing changed
		return
	}

	s.log.Debug("HTML Element relinked", zap.String("URL", src), zap.String("Fixed", refStr))
	selection.SetAttr(attribute, refStr)
}
