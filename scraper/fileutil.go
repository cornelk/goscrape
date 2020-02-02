package scraper

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

const (
	// PageExtension is the file extension that downloaded pages get
	PageExtension = ".html"
	// PageDirIndex is the file name of the index file for every dir
	PageDirIndex = "index" + PageExtension
)

// GetPageFilePath returns a filename for a URL that represents a page.
func GetPageFilePath(url *url.URL) string {
	fileName := url.Path

	// root of domain will be index.html
	if fileName == "" || fileName == "/" {
		fileName = PageDirIndex
		// directory index will be index.html in the directory
	} else if fileName[len(fileName)-1] == '/' {
		fileName += PageDirIndex
	} else {
		ext := filepath.Ext(fileName)
		// if file extension is missing add .html
		if ext == "" {
			fileName += PageExtension
		} else if ext != PageExtension { // replace any other extension with .html
			fileName = fileName[:len(fileName)-len(ext)] + PageExtension
		}
	}

	return fileName
}

// GetFilePath returns a file path for a URL to store the URL content in
func (s *Scraper) GetFilePath(url *url.URL, isAPage bool) string {
	fileName := url.Path
	if isAPage {
		fileName = GetPageFilePath(url)
	}

	var externalHost string
	if url.Host != s.URL.Host {
		externalHost = "_" + url.Host // _ is a prefix for external domains on the filesystem
	}

	return filepath.Join(s.config.OutputDirectory, s.URL.Host, externalHost, fileName)
}

func (s *Scraper) writeFile(filePath string, buf *bytes.Buffer) error {
	dir := filepath.Dir(filePath)
	if len(dir) < len(s.URL.Host) { // nothing to append if it is the root dir
		dir = filepath.Join(".", s.URL.Host, dir)
	}
	s.log.Debug("Creating dir", zap.String("Path", dir))
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	s.log.Debug("Creating file", zap.String("Path", filePath))
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}

	if _, err = f.Write(buf.Bytes()); err != nil {
		_ = f.Close() // try to close and remove file but return the first error
		_ = os.Remove(filePath)
		return err
	}

	return f.Close()
}
