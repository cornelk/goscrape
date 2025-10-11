package scraper

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/cornelk/gotokit/log"
	"github.com/cornelk/gotokit/set"
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
	scraper.fileWriter = func(_ string, _ []byte) error {
		return nil
	}
	scraper.fileExistenceCheck = func(_ string) bool {
		return false
	}
	scraper.httpDownloader = func(_ context.Context, url *url.URL) ([]byte, *url.URL, error) {
		ur := url.String()
		b, ok := urls[ur]
		if ok {
			return b, url, nil
		}
		return nil, nil, fmt.Errorf("url '%s' not found in test data", ur)
	}

	return scraper
}

func TestScraperLinks(t *testing.T) {
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

	expectedProcessed := set.NewFromSlice([]string{
		"/",
		"/page2",
		"/sub",
		"/style.css",
	})
	assert.Equal(t, expectedProcessed, scraper.processed)
}

func TestScraperAttributes(t *testing.T) {
	indexPage := []byte(`
<html>
<head>
</head>

<body background="bg.gif">

<!--embedded image-->
<img src='data:image/gif;base64,R0lGODlhAQABAAD/ACwAAAAAAQABAAACADs%3D=' />

</body>
</html>
`)
	empty := []byte(``)

	startURL := "https://example.org/"
	urls := map[string][]byte{
		"https://example.org/":       indexPage,
		"https://example.org/bg.gif": empty,
	}

	scraper := newTestScraper(t, startURL, urls)
	require.NotNil(t, scraper)

	ctx := context.Background()
	err := scraper.Start(ctx)
	require.NoError(t, err)

	expectedProcessed := set.NewFromSlice([]string{
		"/",
		"/bg.gif",
	})
	assert.Equal(t, expectedProcessed, scraper.processed)
}

func TestScraperInternalCss(t *testing.T) {
	indexPage := []byte(`
<html>
<head>
<style>
h1 {
  background-image: url('https://example.org/background.jpg');
}
h2 {
  background-image: url('/img/bg.jpg');
}
h3 {
  background-image: url(bg3.jpg);
}
</style>
</head>
<body>
</body>
</html>
`)
	empty := []byte(``)

	domain := "example.org"
	file1Reference := "background.jpg"
	file2Reference := "img/bg.jpg"
	file3Reference := "bg3.jpg"
	fullURL := "https://" + domain

	urls := map[string][]byte{
		fullURL + "/":                  indexPage,
		fullURL + "/" + file1Reference: empty,
		fullURL + "/" + file2Reference: empty,
		fullURL + "/" + file3Reference: empty,
	}

	scraper := newTestScraper(t, fullURL+"/", urls)
	require.NotNil(t, scraper)

	files := map[string][]byte{}
	scraper.fileWriter = func(filePath string, data []byte) error {
		files[filePath] = data
		return nil
	}

	ctx := context.Background()
	err := scraper.Start(ctx)
	require.NoError(t, err)

	expectedProcessed := set.NewFromSlice([]string{
		"/",
		"/" + file1Reference,
		"/" + file2Reference,
		"/" + file3Reference,
	})
	require.Equal(t, expectedProcessed, scraper.processed)

	ref := domain + "/index.html"
	content := string(files[ref])
	assert.Contains(t, content, "url('"+file1Reference+"')")
	assert.Contains(t, content, "url('"+file2Reference+"')")
	assert.Contains(t, content, "url("+file3Reference+")")
}

func TestScraperStyleAttributes(t *testing.T) {
	indexPage := []byte(`
<html>
<head>
</head>
<body style="background-image: url('https://example.org/body-bg.jpg');">
<img src="regular.jpg" style="border: 1px solid url('/img-border.png');" />
<a href="page2.html" style="background: url(link-bg.gif) no-repeat;">Link</a>
<body>
</html>
`)
	empty := []byte(``)

	domain := "example.org"
	bodyBgFile := "body-bg.jpg"
	imgBorderFile := "img-border.png"
	linkBgFile := "link-bg.gif"
	regularImgFile := "regular.jpg"
	linkFile := "page2.html"
	fullURL := "https://" + domain

	urls := map[string][]byte{
		fullURL + "/":                  indexPage,
		fullURL + "/" + bodyBgFile:     empty,
		fullURL + "/" + imgBorderFile:  empty,
		fullURL + "/" + linkBgFile:     empty,
		fullURL + "/" + regularImgFile: empty,
		fullURL + "/" + linkFile:       indexPage, // reuse indexPage for simplicity
	}

	scraper := newTestScraper(t, fullURL+"/", urls)
	require.NotNil(t, scraper)

	files := map[string][]byte{}
	scraper.fileWriter = func(filePath string, data []byte) error {
		files[filePath] = data
		return nil
	}

	ctx := context.Background()
	err := scraper.Start(ctx)
	require.NoError(t, err)

	expectedProcessed := set.NewFromSlice([]string{
		"/",
		"/" + bodyBgFile,
		"/" + imgBorderFile,
		"/" + linkBgFile,
		"/" + regularImgFile,
		"/" + linkFile,
	})
	require.Equal(t, expectedProcessed, scraper.processed)

	// Check that the HTML content has been properly relinked
	ref := domain + "/index.html"
	content := string(files[ref])

	// Verify the style attribute URLs have been relinked to relative paths
	// Note: HTML encoding will turn single quotes into &#39;
	assert.Contains(t, content, "url(&#39;"+bodyBgFile+"&#39;)")
	assert.Contains(t, content, "url(&#39;"+imgBorderFile+"&#39;)")
	assert.Contains(t, content, "url("+linkBgFile+")")
}
