package scraper

import (
	"net/http"
	"strings"
)

func Headers(headers []string) http.Header {
	h := http.Header{}
	for _, header := range headers {
		sl := strings.SplitN(header, ":", 2)
		if len(sl) == 2 {
			h.Set(sl[0], sl[1])
		}
	}
	return h
}
