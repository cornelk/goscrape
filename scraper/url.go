package scraper

import (
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

func resolveURL(base *url.URL, reference, mainPageHost string, isHyperlink bool, relativeToRoot string) string {
	ur, err := url.Parse(reference)
	if err != nil {
		return ""
	}

	var resolvedURL *url.URL
	if ur.Host != "" && ur.Host != mainPageHost {
		if isHyperlink { // do not change links to external websites
			return reference
		}

		resolvedURL = base.ResolveReference(ur)
		resolvedURL.Path = filepath.Join("_"+ur.Host, resolvedURL.Path)
	} else {
		if isHyperlink {
			ur.Path = getPageFilePath(ur)
			resolvedURL = base.ResolveReference(ur)
		} else {
			resolvedURL = base.ResolveReference(ur)
		}
	}

	if resolvedURL.Host == mainPageHost {
		resolvedURL.Path = urlRelativeToOther(resolvedURL, base)
		relativeToRoot = ""
	}

	resolvedURL.Host = ""   // remove host
	resolvedURL.Scheme = "" // remove http/https
	resolved := resolvedURL.String()

	if resolved == "" {
		resolved = "/" // website root
	} else {
		if resolved[0] == '/' && len(relativeToRoot) > 0 {
			resolved = relativeToRoot + resolved[1:]
		} else {
			resolved = relativeToRoot + resolved
		}
	}

	if isHyperlink {
		if resolved[len(resolved)-1] == '/' {
			resolved += PageDirIndex // link dir index to index.html
		} else {
			l := strings.LastIndexByte(resolved, '/')
			if l != -1 && l < len(resolved) && resolved[l+1] == '#' {
				resolved = resolved[:l+1] + PageDirIndex + resolved[l+1:] // link fragment correct
			}
		}
	}

	resolved = strings.TrimPrefix(resolved, "/")
	return resolved
}

func urlRelativeToRoot(url *url.URL) string {
	var rel string
	splits := strings.Split(url.Path, "/")
	for i := range splits {
		if (len(splits[i]) > 0) && (i < len(splits)-1) {
			rel += "../"
		}
	}
	return rel
}

func urlRelativeToOther(src, base *url.URL) string {
	srcSplits := strings.Split(src.Path, "/")
	baseSplits := strings.Split(getPageFilePath(base), "/")

	for {
		if len(srcSplits) == 0 || len(baseSplits) == 0 {
			break
		}
		if len(srcSplits[0]) == 0 {
			srcSplits = srcSplits[1:]
			continue
		}
		if len(baseSplits[0]) == 0 {
			baseSplits = baseSplits[1:]
			continue
		}

		if srcSplits[0] == baseSplits[0] {
			srcSplits = srcSplits[1:]
			baseSplits = baseSplits[1:]
		} else {
			break
		}
	}

	var upLevels string
	for i, split := range baseSplits {
		if split == "" {
			continue
		}
		// Page filename is not a level.
		if i == len(baseSplits)-1 {
			break
		}
		upLevels += "../"
	}

	return upLevels + path.Join(srcSplits...)
}
