package scraper

import (
	"testing"

	"go.uber.org/zap/zaptest"
)

func TestRemoveAnchor(t *testing.T) {
	logger := zaptest.NewLogger(t)
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
