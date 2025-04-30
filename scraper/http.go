package scraper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cornelk/gotokit/log"
)

func (s *Scraper) downloadURL(ctx context.Context, u *url.URL) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
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

	return resp, nil
}

func (s *Scraper) downloadURLWithRetriesFor429(ctx context.Context, u *url.URL) (*http.Response, error) {
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		resp, err := s.downloadURL(ctx, u)
		if resp != nil {
			if resp.StatusCode == http.StatusTooManyRequests {
				s.logger.Warn(fmt.Sprintf("Too Many Requests. Retrying again (%s/%s)", strconv.Itoa(i+1), strconv.Itoa(maxRetries)), log.String("url", u.String()))
				// Wait a bit and try again
				time.Sleep(time.Duration((i+1)*1500) * time.Millisecond) // max total of 82.5 seconds within 10 retries using exponential backoff on each retry
				continue
			} else {
				return resp, err
			}
		} else if err != nil {
			return nil, err
		}
		// Success
		return resp, nil
	}
	// Exhausted retries
	err := errors.New("Exhausted retries for URL " + u.String())
	return nil, err
}

func (s *Scraper) downloadURLWithRetries(ctx context.Context, u *url.URL) ([]byte, *url.URL, error) {

	resp, err := s.downloadURLWithRetriesFor429(ctx, u)
	if err != nil {
		return nil, nil, fmt.Errorf("sending HTTP request: %w", err)
	}
	if resp == nil {
		return nil, nil, fmt.Errorf("HTTP response is nil for %v", u)
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
