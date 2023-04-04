package scraper

import (
	"bytes"
	"net/url"
	"os"

	"github.com/cornelk/gotokit/log"
	"github.com/headzoo/surf/browser"
)

// assetProcessor is a processor of a downloaded asset that can transform
// a downloaded file content before it will be stored on disk.
type assetProcessor func(URL *url.URL, buf *bytes.Buffer) *bytes.Buffer

func (s *Scraper) downloadReferences() {
	for _, image := range s.browser.Images() {
		s.imagesQueue = append(s.imagesQueue, &image.DownloadableAsset)
	}
	for _, stylesheet := range s.browser.Stylesheets() {
		s.downloadAsset(&stylesheet.DownloadableAsset, s.checkCSSForUrls)
	}
	for _, script := range s.browser.Scripts() {
		s.downloadAsset(&script.DownloadableAsset, nil)
	}
	for _, image := range s.imagesQueue {
		s.downloadAsset(image, s.checkImageForRecode)
	}
	s.imagesQueue = nil
}

// downloadAsset downloads an asset if it does not exist on disk yet.
func (s *Scraper) downloadAsset(asset *browser.DownloadableAsset, processor assetProcessor) {
	URL := asset.URL
	u := URL.String()
	if _, ok := s.processed[u]; ok {
		return // was already processed
	}
	s.processed[u] = struct{}{}

	if s.includes != nil && !s.isURLIncluded(URL) {
		return
	}
	if s.excludes != nil && s.isURLExcluded(URL) {
		return
	}

	filePath := s.GetFilePath(URL, false)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		return // exists already on disk
	}

	s.logger.Info("Downloading", log.String("url", u))

	buf := &bytes.Buffer{}
	_, err := asset.Download(buf)
	if err != nil {
		s.logger.Error("Downloading asset failed",
			log.String("url", u),
			log.Err(err))
		return
	}

	if processor != nil {
		buf = processor(URL, buf)
	}

	if err = s.writeFile(filePath, buf); err != nil {
		s.logger.Error("Writing asset file failed",
			log.String("url", u),
			log.String("file", filePath),
			log.Err(err))
	}
}
