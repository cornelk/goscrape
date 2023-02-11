package scraper

import (
	"testing"

	"github.com/cornelk/gotokit/log"
)

func TestRemoveAnchor(t *testing.T) {
	logger := log.NewTestLogger(t)
	s, err := New(logger, Config{})
	if err != nil {
		t.Errorf("Scraper New failed: %v", err)
	}

	var fixtures = map[string]string{
		"github.com":                 "github.com",
		"https://github.com/":        "https://github.com/",
		"https://github.com/#anchor": "https://github.com/",
	}

	for input, expected := range fixtures {
		output := s.RemoveAnchor(input)
		if output != expected {
			t.Errorf("URL %s should have been %s but was %s", input, expected, output)
		}
	}
}
