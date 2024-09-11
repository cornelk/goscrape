package css

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/cornelk/gotokit/log"
	"github.com/gorilla/css/scanner"
)

var cssURLRe = regexp.MustCompile(`^url\(['"]?(.*?)['"]?\)$`)

type Token = scanner.Token

type urlProcessor func(token *Token, data string, url *url.URL)

// Process the CSS data and call a processor for every found URL.
func Process(logger *log.Logger, url *url.URL, data string, processor urlProcessor) {
	css := scanner.New(data)

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
			logger.Error("Parsing URL failed",
				log.String("url", src),
				log.Err(err))
			continue
		}
		u = url.ResolveReference(u)
		processor(token, src, u)
	}
}
