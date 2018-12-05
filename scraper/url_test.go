package scraper

import (
	"net/url"
	"testing"
)

func Test_urlRelativeToOther(t *testing.T) {

	type filePathFixture struct {
		SrcURL          url.URL
		BaseURL         url.URL
		ExpectedSrcPath string
	}

	var fixtures = []filePathFixture{
		{url.URL{Path: "/earth/brasil/rio/cat.jpg"}, url.URL{Path: "/earth/brasil/rio"}, "cat.jpg"},
		{url.URL{Path: "/earth/brasil/rio/cat.jpg"}, url.URL{Path: "/earth/"}, "brasil/rio/cat.jpg"},
		{url.URL{Path: "/earth/cat.jpg"}, url.URL{Path: "/earth/brasil/rio/"}, "../../cat.jpg"},
		{url.URL{Path: "/earth/argentina/cat.jpg"}, url.URL{Path: "/earth/brasil/rio/"}, "../../argentina/cat.jpg"},
		{url.URL{Path: "/earth/brasil/rio/cat.jpg"}, url.URL{Path: "/mars/dogtown/"}, "../../earth/brasil/rio/cat.jpg"},
		{url.URL{Path: "///earth//////cat.jpg"}, url.URL{Path: "///earth/brasil//rio////////"}, "../../cat.jpg"},
	}

	for _, fix := range fixtures {
		relativeURL := urlRelativeToOther(&fix.SrcURL, &fix.BaseURL)
		if relativeURL != fix.ExpectedSrcPath {
			t.Errorf("URL %s should have become %s but was %s", fix.SrcURL.Path, fix.BaseURL.Path, relativeURL)
		}
	}
}

func Test_urlRelativeToRoot(t *testing.T) {
	s, err := New("http://localhost")
	if err != nil {
		t.Errorf("Scraper New failed: %v", err)
	}
	type urlFixture struct {
		SrcURL   url.URL
		Expected string
	}

	var fixtures = []urlFixture{
		{url.URL{Path: "/earth/brasil/rio/cat.jpg"}, "../../../"},
		{url.URL{Path: "cat.jpg"}, ""},
		{url.URL{Path: "/earth/argentina"}, "../"},
		{url.URL{Path: "///earth//////cat.jpg"}, "../"},
	}

	for _, fix := range fixtures {
		relativeURL := s.urlRelativeToRoot(&fix.SrcURL)
		if relativeURL != fix.Expected {
			t.Errorf("URL %s should have gotten relative root path %s but was %s", fix.SrcURL.Path, fix.Expected, relativeURL)
		}
	}
}
