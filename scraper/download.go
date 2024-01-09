package scraper

import (
	"bytes"
	"context"
	"net/url"
	"os"

	"github.com/cornelk/goscrape/htmlindex"
	"github.com/cornelk/gotokit/log"
)

// assetProcessor is a processor of a downloaded asset that can transform
// a downloaded file content before it will be stored on disk.
type assetProcessor func(URL *url.URL, buf *bytes.Buffer) *bytes.Buffer

func (s *Scraper) downloadReferences(ctx context.Context, index *htmlindex.Index) {
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
		s.downloadAsset(ctx, ur, s.checkCSSForUrls)
	}

	references, err = index.URLs("script")
	if err != nil {
		s.logger.Error("Getting script nodes URLs failed", log.Err(err))
	}
	for _, ur := range references {
		s.downloadAsset(ctx, ur, nil)
	}

	for _, image := range s.imagesQueue {
		s.downloadAsset(ctx, image, s.checkImageForRecode)
	}
	s.imagesQueue = nil
}

// downloadAsset downloads an asset if it does not exist on disk yet.
func (s *Scraper) downloadAsset(ctx context.Context, u *url.URL, processor assetProcessor) {
	urlFull := u.String()
	if _, ok := s.processed[u.String()]; ok {
		return // was already processed
	}
	s.processed[urlFull] = struct{}{}

	if s.includes != nil && !s.isURLIncluded(u) {
		return
	}
	if s.excludes != nil && s.isURLExcluded(u) {
		return
	}

	filePath := s.GetFilePath(u, false)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return // exists already on disk
	}

	s.logger.Info("Downloading", log.String("url", urlFull))

	buf := &bytes.Buffer{}
	_, err := s.sendHTTPRequest(ctx, u, buf)
	if err != nil {
		s.logger.Error("Downloading asset failed",
			log.String("url", urlFull),
			log.Err(err))
		return
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
}
