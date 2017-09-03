package scraper

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"

	"go.uber.org/zap"
)

// encodeJPEG encodes a new JPG based on the given quality setting
func (s *Scraper) encodeJPEG(img image.Image) *bytes.Buffer {
	o := &jpeg.Options{
		Quality: int(s.ImageQuality),
	}

	outBuf := &bytes.Buffer{}
	err := jpeg.Encode(outBuf, img, o)
	if err != nil {
		return nil
	}
	return outBuf
}

// recodeJPEG recodes the image and returns it if it is smaller than before
func (s *Scraper) recodeJPEG(filePath string, b []byte) *bytes.Buffer {
	inBuf := bytes.NewBuffer(b)
	img, err := jpeg.Decode(inBuf)
	if err != nil {
		return nil
	}

	outBuf := s.encodeJPEG(img)
	if outBuf == nil || outBuf.Len() > len(b) { // only use the new file if it is smaller
		return nil
	}

	s.log.Debug("Recoded JPEG", zap.String("File", filePath), zap.Int("Size old", len(b)), zap.Int("Size new", outBuf.Len()))
	return outBuf
}

// recodePNG recodes the image and returns it if it is smaller than before
func (s *Scraper) recodePNG(filePath string, b []byte) *bytes.Buffer {
	inBuf := bytes.NewBuffer(b)
	img, err := png.Decode(inBuf)
	if err != nil {
		return nil
	}

	outBuf := s.encodeJPEG(img)
	if outBuf == nil || outBuf.Len() > len(b) { // only use the new file if it is smaller
		return nil
	}

	s.log.Debug("Recoded PNG", zap.String("File", filePath), zap.Int("Size old", len(b)), zap.Int("Size new", outBuf.Len()))
	return outBuf
}
