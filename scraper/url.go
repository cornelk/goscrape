package scraper

import (
	"net/url"
	"path/filepath"
	"strings"
)

// RemoveAnchor removes anchors from URLS
func (s *Scraper) RemoveAnchor(path string) string {
	sl := strings.LastIndexByte(path, '/')
	if sl == -1 {
		return path
	}
	an := strings.LastIndexByte(path[sl+1:], '#')
	if an == -1 {
		return path
	}
	return path[:sl+an+1]
}

func (s *Scraper) resolveURL(base *url.URL, reference string, linkIsAPage bool, relativeToRoot string) string {
	ur, err := url.Parse(reference)
	if err != nil {
		return ""
	}

	var resolvedurl *url.URL
	if ur.Host != "" && ur.Host != s.URL.Host {
		if linkIsAPage { // do not change links to external websites
			return ""
		}

		resolvedurl = base.ResolveReference(ur)
		resolvedurl.Path = filepath.Join("_"+ur.Host, resolvedurl.Path)
	} else {
		if linkIsAPage {
			resolvedurl = base.ResolveReference(GetPageURL(ur))
		} else {
			resolvedurl = base.ResolveReference(ur)
		}
	}

	if resolvedurl.Host == s.URL.Host {
		if strings.Contains(resolvedurl.Path, base.Path) {
			resolvedurl.Path = strings.Replace(resolvedurl.Path, base.Path, "", 1)
			relativeToRoot = ""
		}
	}

	resolvedurl.Host = ""   // remove host
	resolvedurl.Scheme = "" // remove http/https
	resolved := resolvedurl.String()

	if resolved == "" {
		resolved = "/" // website root
	} else {
		if resolved[0] == '/' && len(relativeToRoot) > 0 {
			resolved = relativeToRoot + resolved[1:]
		} else {
			resolved = relativeToRoot + resolved
		}
	}

	if linkIsAPage {
		if resolved[len(resolved)-1] == '/' {
			resolved += PageDirIndex // link dir index to index.html
		} else {
			l := strings.LastIndexByte(resolved, '/')
			if l != -1 && l < len(resolved) && resolved[l+1] == '#' {
				resolved = resolved[:l+1] + PageDirIndex + resolved[l+1:] // link anchor correct
			}
		}
	}

	resolved = strings.TrimPrefix(resolved, "/")
	return resolved
}

func (s *Scraper) urlRelativeToRoot(URL *url.URL) string {
	var rel string
	splits := strings.Split(URL.Path, "/")
	for i := range splits {
		if (len(splits[i]) > 0) && (i < len(splits)-1) {
			rel += "../"
		}
	}
	return rel
}
