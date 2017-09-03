package scraper

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"net/url"

	"go.uber.org/zap"
	"gopkg.in/h2non/filetype.v1"
	"gopkg.in/h2non/filetype.v1/matchers"
	"gopkg.in/h2non/filetype.v1/types"
)

func (s *Scraper) checkImageForRecode(URL *url.URL, buf *bytes.Buffer) *bytes.Buffer {
	if s.ImageQuality == 0 {
		return buf
	}

	kind, err := filetype.Match(buf.Bytes())
	if err != nil || kind == types.Unknown {
		return buf
	}

	s.log.Debug("File type detected", zap.String("Type", kind.MIME.Type), zap.String("Subtype", kind.MIME.Subtype))

	if kind.MIME.Type == matchers.TypeJpeg.MIME.Type && kind.MIME.Subtype == matchers.TypeJpeg.MIME.Subtype {
		if recoded := s.recodeJPEG(URL, buf.Bytes()); recoded != nil {
			return recoded
		}
		return buf
	}

	if kind.MIME.Type == matchers.TypePng.MIME.Type && kind.MIME.Subtype == matchers.TypePng.MIME.Subtype {
		if recoded := s.recodePNG(URL, buf.Bytes()); recoded != nil {
			return recoded
		}
		return buf
	}

	return buf
}

// encodeJPEG encodes a new JPG based on the given quality setting
func (s *Scraper) encodeJPEG(img image.Image) *bytes.Buffer {
	o := &jpeg.Options{
		Quality: int(s.ImageQuality),
	}

	outBuf := &bytes.Buffer{}
	if err := jpeg.Encode(outBuf, img, o); err != nil {
		return nil
	}
	return outBuf
}

// recodeJPEG recodes the image and returns it if it is smaller than before
func (s *Scraper) recodeJPEG(URL *url.URL, b []byte) *bytes.Buffer {
	inBuf := bytes.NewBuffer(b)
	img, err := jpeg.Decode(inBuf)
	if err != nil {
		return nil
	}

	outBuf := s.encodeJPEG(img)
	if outBuf == nil || outBuf.Len() > len(b) { // only use the new file if it is smaller
		return nil
	}

	s.log.Debug("Recoded JPEG", zap.Stringer("URL", URL), zap.Int("Size old", len(b)), zap.Int("Size new", outBuf.Len()))
	return outBuf
}

// recodePNG recodes the image and returns it if it is smaller than before
func (s *Scraper) recodePNG(URL *url.URL, b []byte) *bytes.Buffer {
	inBuf := bytes.NewBuffer(b)
	img, err := png.Decode(inBuf)
	if err != nil {
		return nil
	}

	outBuf := s.encodeJPEG(img)
	if outBuf == nil || outBuf.Len() > len(b) { // only use the new file if it is smaller
		return nil
	}

	s.log.Debug("Recoded PNG", zap.Stringer("URL", URL), zap.Int("Size old", len(b)), zap.Int("Size new", outBuf.Len()))
	return outBuf
}
