package scraper

import (
	"bytes"
	"net/url"
	"testing"

	"github.com/cornelk/gotokit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckCSSForURLs(t *testing.T) {
	logger := log.NewTestLogger(t)
	cfg := Config{
		URL: "http://localhost",
	}
	s, err := New(logger, cfg)
	require.NoError(t, err)

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
			assert.Empty(t, s.imagesQueue)
			continue
		}

		assert.Positive(t, len(s.imagesQueue))

		res := s.imagesQueue[0].String()
		assert.Equal(t, expected, res)
	}
}
