package scraper

import (
	"net/url"
	"os"
	"testing"

	"github.com/cornelk/gotokit/log"
)

func TestGetFilePath(t *testing.T) {
	type filePathFixture struct {
		BaseURL          string
		DownloadURL      string
		ExpectedFilePath string
	}

	pathSeparator := string(os.PathSeparator)
	expectedBasePath := "google.com" + pathSeparator
	var fixtures = []filePathFixture{
		{"https://google.com/", "https://github.com/", expectedBasePath + "_github.com" + pathSeparator + "index.html"},
		{"https://google.com/", "https://github.com/#fragment", expectedBasePath + "_github.com" + pathSeparator + "index.html"},
		{"https://google.com/", "https://github.com/test", expectedBasePath + "_github.com" + pathSeparator + "test.html"},
		{"https://google.com/", "https://github.com/test/", expectedBasePath + "_github.com" + pathSeparator + "test" + pathSeparator + "index.html"},
		{"https://google.com/", "https://github.com/test.aspx", expectedBasePath + "_github.com" + pathSeparator + "test.aspx"},
		{"https://google.com/", "https://google.com/settings", expectedBasePath + "settings.html"},
	}

	var cfg Config
	logger := log.NewTestLogger(t)
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

		output := s.getFilePath(URL, true)
		if output != fix.ExpectedFilePath {
			t.Errorf("URL %s should have become file %s but was %s", fix.DownloadURL, fix.ExpectedFilePath, output)
		}
	}
}
