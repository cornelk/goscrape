package scraper

import (
	"net/url"
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestGetFilePath(t *testing.T) {
	type filePathFixture struct {
		BaseURL          string
		DownloadURL      string
		ExpectedFilePath string
	}

	var fixtures = []filePathFixture{
		{"https://google.com/", "https://github.com/", "google.com/_github.com/index.html"},
		{"https://google.com/", "https://github.com/#anchor", "google.com/_github.com/index.html"},
		{"https://google.com/", "https://github.com/test", "google.com/_github.com/test.html"},
		{"https://google.com/", "https://github.com/test/", "google.com/_github.com/test/index.html"},
		{"https://google.com/", "https://github.com/test.aspx", "google.com/_github.com/test.html"},
		{"https://google.com/", "https://google.com/settings", "google.com/settings.html"},
	}

	var cfg Config
	logger := zaptest.NewLogger(t)
	for _, fix := range fixtures {
		cfg.URL = fix.BaseURL
		s, err := New(logger, cfg)
		if err != nil {
			t.Errorf("Scraper New failed: %v", err)
		}

		URL, err := url.Parse(fix.DownloadURL)
		if err != nil {
			t.Errorf("URL parse failed: %v", err)
		}

		output := s.GetFilePath(URL, true)
		if output != fix.ExpectedFilePath {
			t.Errorf("URL %s should have become file %s but was %s", fix.DownloadURL, fix.ExpectedFilePath, output)
		}
	}
}
