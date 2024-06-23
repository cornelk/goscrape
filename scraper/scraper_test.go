package scraper

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/cornelk/gotokit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestScraper(t *testing.T, startURL string, urls map[string][]byte) *Scraper {
	t.Helper()

	logger := log.NewTestLogger(t)
	cfg := Config{
		URL: startURL,
	}
	scraper, err := New(logger, cfg)
	require.NoError(t, err)
	require.NotNil(t, scraper)

	scraper.dirCreator = func(_ string) error {
		return nil
	}
	scraper.fileWriter = func(_ string, _ *bytes.Buffer) error {
		return nil
	}
	scraper.fileExistenceCheck = func(_ string) bool {
		return false
	}
	scraper.httpDownloader = func(_ context.Context, url *url.URL) (*bytes.Buffer, *url.URL, error) {
		ur := url.String()
		b, ok := urls[ur]
		if ok {
			return bytes.NewBuffer(b), url, nil
		}
		return nil, nil, fmt.Errorf("url '%s' not found in test data", ur)
	}

	return scraper
}

func TestScraper(t *testing.T) {
	indexPage := []byte(`
<html>
<head>
<link href=' https://example.org/style.css#fragment' rel='stylesheet' type='text/css'>
</head>
<body>
<a href="https://example.org/page2">Example</a>
</body>
</html>
`)

	page2 := []byte(`
<html>
<body>
<!--link to index with fragment-->
<a href="/#fragment">a</a>
<!--link to page with fragment-->
<a href="/sub/#fragment">a</a>
</body>
</html>
`)

	css := []byte(``)

	startURL := "https://example.org/#fragment" // start page with fragment
	urls := map[string][]byte{
		"https://example.org/":          indexPage,
		"https://example.org/page2":     page2,
		"https://example.org/sub/":      indexPage,
		"https://example.org/style.css": css,
	}

	scraper := newTestScraper(t, startURL, urls)
	require.NotNil(t, scraper)

	ctx := context.Background()
	err := scraper.Start(ctx)
	require.NoError(t, err)
	assert.Contains(t, scraper.processed, "/")
	assert.Contains(t, scraper.processed, "/page2")
	assert.Contains(t, scraper.processed, "/sub/")
}
