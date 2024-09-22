package scraper

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/cornelk/gotokit/log"
	"github.com/gorilla/css/scanner"
)

var cssURLRe = regexp.MustCompile(`^url\(['"]?(.*?)['"]?\)$`)

func (s *Scraper) checkCSSForUrls(url *url.URL, data []byte) []byte {
	urls := make(map[string]string)
	str := string(data)
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
			s.logger.Error("Parsing URL failed",
				log.String("url", src),
				log.Err(err))
			continue
		}
		u = url.ResolveReference(u)

		s.imagesQueue = append(s.imagesQueue, u)

		cssPath := *url
		cssPath.Path = path.Dir(cssPath.Path) + "/"
		resolved := resolveURL(&cssPath, src, s.URL.Host, false, "")
		urls[token.Value] = resolved
	}

	if len(urls) == 0 {
		return data
	}

	for ori, filePath := range urls {
		fixed := fmt.Sprintf("url(%s)", filePath)
		str = strings.ReplaceAll(str, ori, fixed)
		s.logger.Debug("CSS Element relinked",
			log.String("url", ori),
			log.String("fixed_url", fixed))
	}

	return []byte(str)
}
