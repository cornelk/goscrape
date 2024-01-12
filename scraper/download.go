package scraper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/cornelk/goscrape/htmlindex"
	"github.com/cornelk/gotokit/log"
)

// assetProcessor is a processor of a downloaded asset that can transform
// a downloaded file content before it will be stored on disk.
type assetProcessor func(URL *url.URL, buf *bytes.Buffer) *bytes.Buffer

func (s *Scraper) downloadReferences(ctx context.Context, index *htmlindex.Index) error {
	references, err := index.URLs("img")
	if err != nil {
		s.logger.Error("Getting img nodes URLs failed", log.Err(err))
	}
	s.imagesQueue = append(s.imagesQueue, references...)

	references, err = index.URLs("link")
	if err != nil {
		s.logger.Error("Getting link nodes URLs failed", log.Err(err))
	}
	for _, ur := range references {
		if err := s.downloadAsset(ctx, ur, s.checkCSSForUrls); err != nil && errors.Is(err, context.Canceled) {
			return err
		}
	}

	references, err = index.URLs("script")
	if err != nil {
		s.logger.Error("Getting script nodes URLs failed", log.Err(err))
	}
	for _, ur := range references {
		if err := s.downloadAsset(ctx, ur, nil); err != nil && errors.Is(err, context.Canceled) {
			return err
		}
	}

	for _, image := range s.imagesQueue {
		if err := s.downloadAsset(ctx, image, s.checkImageForRecode); err != nil && errors.Is(err, context.Canceled) {
			return err
		}
	}
	s.imagesQueue = nil
	return nil
}

// downloadAsset downloads an asset if it does not exist on disk yet.
func (s *Scraper) downloadAsset(ctx context.Context, u *url.URL, processor assetProcessor) error {
	urlFull := u.String()

	if !s.shouldURLBeDownloaded(u, 0, true) {
		return nil
	}

	filePath := s.getFilePath(u, false)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return nil // exists already on disk
	}

	s.logger.Info("Downloading asset", log.String("url", urlFull))

	buf := &bytes.Buffer{}
	_, err := s.sendHTTPRequest(ctx, u, buf)
	if err != nil {
		s.logger.Error("Downloading asset failed",
			log.String("url", urlFull),
			log.Err(err))
		return fmt.Errorf("downloading asset: %w", err)
	}

	if processor != nil {
		buf = processor(u, buf)
	}

	if err = s.writeFile(filePath, buf); err != nil {
		s.logger.Error("Writing asset file failed",
			log.String("url", urlFull),
			log.String("file", filePath),
			log.Err(err))
	}

	return nil
}
