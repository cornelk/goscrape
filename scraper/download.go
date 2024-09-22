package scraper

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/cornelk/goscrape/htmlindex"
	"github.com/cornelk/gotokit/log"
)

// assetProcessor is a processor of a downloaded asset that can transform
// a downloaded file content before it will be stored on disk.
type assetProcessor func(URL *url.URL, data []byte) []byte

var tagsWithReferences = []string{
	htmlindex.LinkTag,
	htmlindex.ScriptTag,
	htmlindex.BodyTag,
}

func (s *Scraper) downloadReferences(ctx context.Context, index *htmlindex.Index) error {
	references, err := index.URLs(htmlindex.BodyTag)
	if err != nil {
		s.logger.Error("Getting body node URLs failed", log.Err(err))
	}
	s.imagesQueue = append(s.imagesQueue, references...)

	references, err = index.URLs(htmlindex.ImgTag)
	if err != nil {
		s.logger.Error("Getting img node URLs failed", log.Err(err))
	}
	s.imagesQueue = append(s.imagesQueue, references...)

	for _, tag := range tagsWithReferences {
		references, err = index.URLs(tag)
		if err != nil {
			s.logger.Error("Getting node URLs failed",
				log.String("node", tag),
				log.Err(err))
		}

		var processor assetProcessor
		if tag == htmlindex.LinkTag {
			processor = s.checkCSSForUrls
		}
		for _, ur := range references {
			if err := s.downloadAsset(ctx, ur, processor); err != nil && errors.Is(err, context.Canceled) {
				return err
			}
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
	u.Fragment = ""
	urlFull := u.String()

	if !s.shouldURLBeDownloaded(u, 0, true) {
		return nil
	}

	filePath := s.getFilePath(u, false)
	if s.fileExists(filePath) {
		return nil
	}

	s.logger.Info("Downloading asset", log.String("url", urlFull))
	data, _, err := s.httpDownloader(ctx, u)
	if err != nil {
		s.logger.Error("Downloading asset failed",
			log.String("url", urlFull),
			log.Err(err))
		return fmt.Errorf("downloading asset: %w", err)
	}

	if processor != nil {
		data = processor(u, data)
	}

	if err = s.fileWriter(filePath, data); err != nil {
		s.logger.Error("Writing asset file failed",
			log.String("url", urlFull),
			log.String("file", filePath),
			log.Err(err))
	}

	return nil
}
