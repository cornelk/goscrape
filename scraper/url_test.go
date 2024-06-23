package scraper

import (
	"net/url"
	"testing"
)

func TestResolveURL(t *testing.T) {
	type filePathFixture struct {
		BaseURL        url.URL
		Reference      string
		IsHyperlink    bool
		RelativeToRoot string
		Resolved       string
	}

	pathlessURL := url.URL{
		Scheme: "https",
		Host:   "petpic.xyz",
		Path:   "",
	}

	URL := url.URL{
		Scheme: "https",
		Host:   "petpic.xyz",
		Path:   "/earth/",
	}

	var fixtures = []filePathFixture{
		{pathlessURL, "", true, "", "index.html"},
		{pathlessURL, "#contents", true, "", "#contents"},
		{URL, "brasil/index.html", true, "", "brasil/index.html"},
		{URL, "brasil/rio/index.html", true, "", "brasil/rio/index.html"},
		{URL, "../argentina/cat.jpg", false, "", "../argentina/cat.jpg"},
	}

	for _, fix := range fixtures {
		resolved := resolveURL(&fix.BaseURL, fix.Reference, URL.Host, fix.IsHyperlink, fix.RelativeToRoot)

		if resolved != fix.Resolved {
			t.Errorf("Reference %s should be resolved to %s but was %s", fix.Reference, fix.Resolved, resolved)
		}
	}
}

func Test_urlRelativeToOther(t *testing.T) {
	type filePathFixture struct {
		SrcURL          url.URL
		BaseURL         url.URL
		ExpectedSrcPath string
	}

	var fixtures = []filePathFixture{
		{url.URL{Path: "/earth/brasil/rio/cat.jpg"}, url.URL{Path: "/earth/brasil/rio/"}, "cat.jpg"},
		{url.URL{Path: "/earth/brasil/rio/cat.jpg"}, url.URL{Path: "/earth/"}, "brasil/rio/cat.jpg"},
		{url.URL{Path: "/earth/cat.jpg"}, url.URL{Path: "/earth/brasil/rio/"}, "../../cat.jpg"},
		{url.URL{Path: "/earth/argentina/cat.jpg"}, url.URL{Path: "/earth/brasil/rio/"}, "../../argentina/cat.jpg"},
		{url.URL{Path: "/earth/brasil/rio/cat.jpg"}, url.URL{Path: "/mars/dogtown/"}, "../../earth/brasil/rio/cat.jpg"},
		{url.URL{Path: "///earth//////cat.jpg"}, url.URL{Path: "///earth/brasil//rio////////"}, "../../cat.jpg"},
	}

	for _, fix := range fixtures {
		relativeURL := urlRelativeToOther(&fix.SrcURL, &fix.BaseURL)
		if relativeURL != fix.ExpectedSrcPath {
			t.Errorf("URL %s should have become %s but was %s", fix.SrcURL.Path, fix.ExpectedSrcPath, relativeURL)
		}
	}
}

func Test_urlRelativeToRoot(t *testing.T) {
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
		relativeURL := urlRelativeToRoot(&fix.SrcURL)
		if relativeURL != fix.Expected {
			t.Errorf("URL %s should have gotten relative root path %s but was %s", fix.SrcURL.Path, fix.Expected, relativeURL)
		}
	}
}
