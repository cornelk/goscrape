package scraper

import (
	"bytes"
	"net/url"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestCheckCSSForURLs(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cfg := Config{
		URL: "http://localhost",
	}
	s, err := New(logger, cfg)
	if err != nil {
		t.Errorf("Scraper New failed: %v", err)
	}

	var fixtures = map[string]string{
		"url('http://localhost/uri/between/single/quote')": "http://localhost/uri/between/single/quote",
		`url("http://localhost/uri/between/double/quote")`: "http://localhost/uri/between/double/quote",
		"url(http://localhost/uri)":                        "http://localhost/uri",
		"url(data:image/gif;base64,R0lGODl)":               "",
		`div#gopher {
			background: url(/doc/gopher/frontpage.png) no-repeat;
			height: 155px;
			}`: "http://localhost/doc/gopher/frontpage.png",
	}

	u, _ := url.Parse("http://localhost")
	for input, expected := range fixtures {
		s.imagesQueue = nil
		buf := bytes.NewBufferString(input)
		s.checkCSSForUrls(u, buf)

		if expected == "" {
			if len(s.imagesQueue) != 0 {
				t.Errorf("CSS %s should not result in an image in queue with URL %s", input, s.imagesQueue[0].URL.String())
			}
			continue
		}

		if len(s.imagesQueue) == 0 {
			t.Errorf("CSS %s did not result in an image in queue", input)
		}

		res := s.imagesQueue[0].URL.String()
		if res != expected {
			t.Errorf("URL %s should have been %s but was %s", input, expected, res)
		}
	}
}
