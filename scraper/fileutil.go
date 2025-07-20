package scraper

import (
	"fmt"
	"hash/fnv"
	"net/url"
	"path/filepath"
	"strings"
)

const (
	// PageExtension is the file extension that downloaded pages get.
	PageExtension = ".html"
	// PageDirIndex is the file name of the index file for every dir.
	PageDirIndex = "index" + PageExtension
	// MaxFilenameLength is the maximum length for a filename component to ensure filesystem compatibility.
	MaxFilenameLength = 200
)

// getFilePath returns a file path for a URL to store the URL content in.
func (s *Scraper) getFilePath(url *url.URL, isAPage bool) string {
	fileName := url.Path
	if isAPage {
		fileName = getPageFilePath(url)
	}

	var externalHost string
	if url.Host != s.URL.Host {
		externalHost = "_" + url.Host // _ is a prefix for external domains on the filesystem
	}

	// Split the file path into directory and filename components
	dir := filepath.Dir(fileName)
	base := filepath.Base(fileName)

	// Truncate the filename component if it's too long
	truncatedBase := truncateFilename(base)

	// Reconstruct the path with the truncated filename
	if dir == "." {
		fileName = truncatedBase
	} else {
		fileName = filepath.Join(dir, truncatedBase)
	}

	return filepath.Join(s.config.OutputDirectory, s.URL.Host, externalHost, fileName)
}

// getPageFilePath returns a filename for a URL that represents a page.
func getPageFilePath(url *url.URL) string {
	fileName := url.Path

	// root of domain will be index.html
	switch {
	case fileName == "" || fileName == "/":
		fileName = PageDirIndex
		// directory index will be index.html in the directory

	case fileName[len(fileName)-1] == '/':
		fileName += PageDirIndex

	default:
		ext := filepath.Ext(fileName)
		// if file extension is missing add .html, otherwise keep the existing file extension
		if ext == "" {
			fileName += PageExtension
		}
	}

	return fileName
}

// truncateFilename truncates a filename if it exceeds MaxFilenameLength while preserving the extension.
func truncateFilename(filename string) string {
	if len(filename) <= MaxFilenameLength {
		return filename
	}

	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)

	// Calculate how much space we need for hash and extension
	hashLength := 8 // Using first 8 hex characters (from 32-bit FNV)
	reservedLength := hashLength + len(ext)

	// If the extension alone is too long, truncate it too
	if reservedLength > MaxFilenameLength {
		ext = ext[:MaxFilenameLength-hashLength]
		reservedLength = hashLength + len(ext)
	}

	maxBaseLength := MaxFilenameLength - reservedLength
	if maxBaseLength <= 0 {
		maxBaseLength = 1
	}

	truncatedBase := baseName[:maxBaseLength]

	// Generate FNV-1a hash of original filename to ensure uniqueness
	h := fnv.New32a()
	_, _ = h.Write([]byte(filename))
	hashStr := fmt.Sprintf("%08x", h.Sum32())[:hashLength]

	return truncatedBase + hashStr + ext
}
