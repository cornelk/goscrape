package scraper

import (
	"bytes"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/cornelk/gotokit/log"
	"github.com/gorilla/css/scanner"
	"github.com/headzoo/surf/browser"
)

func (s *Scraper) checkCSSForUrls(url *url.URL, buf *bytes.Buffer) *bytes.Buffer {
	urls := make(map[string]string)
	str := buf.String()
	css := scanner.New(str)

	for {
		token := css.Next()
		if token.Type == scanner.TokenEOF || token.Type == scanner.TokenError {
			break
		}
		if token.Type != scanner.TokenURI {
			continue
		}

		match := s.cssURLRe.FindStringSubmatch(token.Value)
		if match == nil {
			continue
		}

		src := match[1]
		if strings.HasPrefix(strings.ToLower(src), "data:") {
			continue // skip embedded data
		}

		u, err := url.Parse(src)
		if err != nil {
			return buf
		}
		u = url.ResolveReference(u)

		img := browser.NewImageAsset(u, "", "", "")
		s.imagesQueue = append(s.imagesQueue, &img.DownloadableAsset)

		cssPath := *url
		cssPath.Path = path.Dir(cssPath.Path) + "/"
		resolved := s.resolveURL(&cssPath, src, false, "")
		urls[token.Value] = resolved
	}

	if len(urls) == 0 {
		return buf
	}

	for ori, filePath := range urls {
		fixed := fmt.Sprintf("url(%s)", filePath)
		str = strings.ReplaceAll(str, ori, fixed)
		s.logger.Debug("CSS Element relinked",
			log.String("url", ori),
			log.String("fixed_url", fixed))
	}

	return bytes.NewBufferString(str)
}
