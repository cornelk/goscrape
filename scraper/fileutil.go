package scraper

import (
	"bytes"
	"net/url"
	"os"
	"path/filepath"

	"github.com/uber-go/zap"
)

func (s *Scraper) getFilePath(URL *url.URL) string {
	fileName := URL.Path
	if fileName == "" || fileName == "/" {
		fileName = "index.html"
	} else if fileName[len(fileName)-1] == '/' {
		fileName += "index.html"
	}

	return filepath.Join(".", s.URL.Host, fileName)
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
		return err
	}

	return f.Close()
}
