package scraper

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/url"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/matchers"
	"github.com/h2non/filetype/types"
	"go.uber.org/zap"
)

func (s *Scraper) checkImageForRecode(url *url.URL, buf *bytes.Buffer) *bytes.Buffer {
	if s.config.ImageQuality == 0 {
		return buf
	}

	kind, err := filetype.Match(buf.Bytes())
	if err != nil || kind == types.Unknown {
		return buf
	}

	s.log.Debug("File type detected",
		zap.String("Type", kind.MIME.Type),
		zap.String("Subtype", kind.MIME.Subtype))

	if kind.MIME.Type == matchers.TypeJpeg.MIME.Type && kind.MIME.Subtype == matchers.TypeJpeg.MIME.Subtype {
		if recoded := s.recodeJPEG(url, buf.Bytes()); recoded != nil {
			return recoded
		}
		return buf
	}

	if kind.MIME.Type == matchers.TypePng.MIME.Type && kind.MIME.Subtype == matchers.TypePng.MIME.Subtype {
		if recoded := s.recodePNG(url, buf.Bytes()); recoded != nil {
			return recoded
		}
		return buf
	}

	return buf
}

// encodeJPEG encodes a new JPG based on the given quality setting
func (s *Scraper) encodeJPEG(img image.Image) *bytes.Buffer {
	o := &jpeg.Options{
		Quality: int(s.config.ImageQuality),
	}

	outBuf := &bytes.Buffer{}
	if err := jpeg.Encode(outBuf, img, o); err != nil {
		return nil
	}
	return outBuf
}

// recodeJPEG recodes the image and returns it if it is smaller than before
func (s *Scraper) recodeJPEG(url fmt.Stringer, b []byte) *bytes.Buffer {
	inBuf := bytes.NewBuffer(b)
	img, err := jpeg.Decode(inBuf)
	if err != nil {
		return nil
	}

	outBuf := s.encodeJPEG(img)
	if outBuf == nil || outBuf.Len() > len(b) { // only use the new file if it is smaller
		return nil
	}

	s.log.Debug("Recoded JPEG",
		zap.Stringer("URL", url),
		zap.Int("Size old", len(b)),
		zap.Int("Size new", outBuf.Len()))
	return outBuf
}

// recodePNG recodes the image and returns it if it is smaller than before
func (s *Scraper) recodePNG(url fmt.Stringer, b []byte) *bytes.Buffer {
	inBuf := bytes.NewBuffer(b)
	img, err := png.Decode(inBuf)
	if err != nil {
		return nil
	}

	outBuf := s.encodeJPEG(img)
	if outBuf == nil || outBuf.Len() > len(b) { // only use the new file if it is smaller
		return nil
	}

	s.log.Debug("Recoded PNG",
		zap.Stringer("URL", url),
		zap.Int("Size old", len(b)),
		zap.Int("Size new", outBuf.Len()))
	return outBuf
}
