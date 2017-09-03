package scraper

import (
	"bytes"
	"net/url"
	"regexp"
	"strings"

	"fmt"
	"github.com/gorilla/css/scanner"
	"github.com/headzoo/surf/browser"
	"go.uber.org/zap"
)

var cssURLRe = regexp.MustCompile(`^url\(['"]?(.*?)['"]?\)$`)

func (s *Scraper) checkCSSForURLs(URL *url.URL, buf *bytes.Buffer) *bytes.Buffer {
	replacings := make(map[string]string)
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

		match := cssURLRe.FindStringSubmatch(token.Value)
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
		u = URL.ResolveReference(u)

		img := browser.NewImageAsset(u, "", "", "")
		s.imagesQueue = append(s.imagesQueue, &img.DownloadableAsset)

		resolved := s.resolveURL(u, src, false, "")
		replacings[token.Value] = resolved
	}

	if len(replacings) == 0 {
		return buf
	}

	for ori, fpath := range replacings {
		fixed := fmt.Sprintf("url(%s)", fpath)
		str = strings.Replace(str, ori, fixed, -1)
		s.log.Debug("CSS Element relinked", zap.String("URL", ori), zap.String("Fixed", fixed))
	}

	return bytes.NewBufferString(str)
}
