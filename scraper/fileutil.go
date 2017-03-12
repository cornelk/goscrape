package scraper

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"

	"github.com/uber-go/zap"
	"gopkg.in/h2non/filetype.v1"
	"gopkg.in/h2non/filetype.v1/matchers"
	"gopkg.in/h2non/filetype.v1/types"
)

var (
	// PageExtension is the file extension that downloaded pages get
	PageExtension = ".html"
	// PageDirIndex is the file name of the index file for every dir
	PageDirIndex = "index" + PageExtension
)

// GetFilePath returns a file path for a URL to store the URL content in
func (s *Scraper) GetFilePath(URL *url.URL, isAPage bool) string {
	fileName := URL.Path
	if isAPage {
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
			} else {
				// replace any other extension with .html
				if ext != PageExtension {
					fileName = fileName[:len(fileName)-len(ext)] + PageExtension
				}
			}
		}
	}

	var externalHost string
	if URL.Host != s.URL.Host {
		externalHost = "_" + URL.Host // _ is a prefix for external domains on the filesystem
	}

	return filepath.Join(".", s.URL.Host, externalHost, fileName)
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

	_, err = f.Write(buf.Bytes())
	if err != nil {
		_ = f.Close()
		_ = os.Remove(filePath)
		return err
	}

	return f.Close()
}

func (s *Scraper) checkFileTypeForRecode(filePath string, buf *bytes.Buffer) *bytes.Buffer {
	if s.ImageQuality == 0 {
		return buf
	}

	kind, err := filetype.Match(buf.Bytes())
	if err != nil || kind == types.Unknown {
		return buf
	}

	s.log.Debug("File type detected", zap.String("Type", kind.MIME.Type), zap.String("Subtype", kind.MIME.Subtype))

	if kind.MIME.Type == matchers.TypeJpeg.MIME.Type && kind.MIME.Subtype == matchers.TypeJpeg.MIME.Subtype {
		recoded := s.recodeJPEG(filePath, buf.Bytes())
		if recoded != nil {
			return recoded
		}
		return buf
	}

	if kind.MIME.Type == matchers.TypePng.MIME.Type && kind.MIME.Subtype == matchers.TypePng.MIME.Subtype {
		recoded := s.recodePNG(filePath, buf.Bytes())
		if recoded != nil {
			return recoded
		}
		return buf
	}

	return buf
}
