package scraper

import "testing"

func TestRemoveAnchor(t *testing.T) {
	s, err := New("")
	if err != nil {
		t.Errorf("Scraper New failed: %v", err)
	}

	var fixtures = map[string]string{
		"github.com":                 "github.com",
		"https://github.com/":        "https://github.com/",
		"https://github.com/#anchor": "https://github.com/",
	}

	for input, result := range fixtures {
		output := s.RemoveAnchor(input)
		if output != result {
			t.Errorf("URL %s should have been %s but was %s", input, result, output)
		}
	}
}
