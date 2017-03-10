package scraper

import (
	"net/url"
	"testing"
)

func TestGetFilePath(t *testing.T) {
	s, err := New("https://google.com/")
	if err != nil {
		t.Errorf("Scraper New failed: %v", err)
	}

	var fixtures = map[string]string{
		"https://github.com/":          "google.com/_github.com/index.html",
		"https://github.com/#anchor":   "google.com/_github.com/index.html",
		"https://github.com/test":      "google.com/_github.com/test.html",
		"https://github.com/test.aspx": "google.com/_github.com/test.html",
	}

	for input, result := range fixtures {
		URL, err := url.Parse(input)
		if err != nil {
			t.Errorf("URL parse failed: %v", err)
		}

		output := s.GetFilePath(URL, true)
		if output != result {
			t.Errorf("URL %s should have become file %s but was %s", input, result, output)
		}
	}
}
