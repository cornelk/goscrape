package scraper

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/cornelk/gotokit/log"
)

func (s *Scraper) downloadURL(ctx context.Context, u *url.URL) ([]byte, *url.URL, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("creating HTTP request: %w", err)
	}

	req.Header.Set("User-Agent", s.config.UserAgent)
	if s.auth != "" {
		req.Header.Set("Authorization", s.auth)
	}

	for key, values := range s.config.Header {
		for _, value := range values {
			req.Header.Set(key, value)
		}
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("sending HTTP request: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.logger.Error("Closing HTTP Request body failed",
				log.String("url", u.String()),
				log.Err(err))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("unexpected HTTP request status code %d", resp.StatusCode)
	}

	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, resp.Body); err != nil {
		return nil, nil, fmt.Errorf("reading HTTP request body: %w", err)
	}
	return buf.Bytes(), resp.Request.URL, nil
}

func Headers(headers []string) http.Header {
	h := http.Header{}
	for _, header := range headers {
		sl := strings.SplitN(header, ":", 2)
		if len(sl) == 2 {
			h.Set(sl[0], sl[1])
		}
	}
	return h
}
