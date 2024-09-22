package scraper

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cornelk/gotokit/log"
)

// createDownloadPath creates the download path if it does not exist yet.
func (s *Scraper) createDownloadPath(path string) error {
	if path == "" {
		return nil
	}

	s.logger.Debug("Creating dir", log.String("path", path))
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory '%s': %w", path, err)
	}
	return nil
}

func (s *Scraper) writeFile(filePath string, data []byte) error {
	dir := filepath.Dir(filePath)
	if len(dir) < len(s.URL.Host) { // nothing to append if it is the root dir
		dir = filepath.Join(".", s.URL.Host, dir)
	}

	if err := s.dirCreator(dir); err != nil {
		return err
	}

	s.logger.Debug("Creating file", log.String("path", filePath))
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating file '%s': %w", filePath, err)
	}

	if _, err = f.Write(data); err != nil {
		// nolint: wrapcheck
		_ = f.Close() // try to close and remove file but return the first error
		_ = os.Remove(filePath)
		return fmt.Errorf("writing to file: %w", err)
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("closing file: %w", err)
	}
	return nil
}

func (s *Scraper) fileExists(filePath string) bool {
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return true
	}
	return false
}
