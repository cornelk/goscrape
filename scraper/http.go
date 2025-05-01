package scraper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cornelk/gotokit/app"
	"github.com/cornelk/gotokit/log"
)

var (
	maxRetries = 10
	retryDelay = 1500 * time.Millisecond

	errExhaustedRetries = errors.New("exhausted retries")
)

func (s *Scraper) downloadURL(ctx context.Context, u *url.URL) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
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
		return nil, fmt.Errorf("executing HTTP request: %w", err)
	}

	return resp, nil
}

func (s *Scraper) downloadURLWithRetries(ctx context.Context, u *url.URL) ([]byte, *url.URL, error) {
	var err error
	var resp *http.Response

	for retries := range maxRetries + 2 {
		if retries == maxRetries+1 {
			return nil, nil, fmt.Errorf("%w for URL %s", errExhaustedRetries, u)
		}

		resp, err = s.downloadURL(ctx, u)
		if err != nil {
			return nil, nil, err
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			s.logger.Warn("Too Many Requests. Retrying again",
				log.Int("num", retries+1),
				log.Int("max", maxRetries),
				log.String("url", u.String()))

			// Wait a bit and try again using exponential backoff on each retry
			if err := app.Sleep(ctx, (time.Duration(retries)+1)*retryDelay); err != nil {
				return nil, nil, fmt.Errorf("sleeping between retries: %w", err)
			}
			continue
		}
		break
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

// Headers converts a slice of strings to a http.Header.
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
