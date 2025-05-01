package scraper

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/cornelk/gotokit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {
	headers := Headers([]string{"a:b", "c:d:e"})
	assert.Equal(t, "b", headers.Get("a"))
	assert.Equal(t, "d:e", headers.Get("c"))
}

func TestDownloadURLWithRetries(t *testing.T) {
	ctx := context.Background()
	expected := "ok"

	var retry int
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if retry < maxRetries {
			retry++
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, err := fmt.Fprint(w, expected)
		assert.NoError(t, err)
	}))
	defer svr.Close()

	ur, err := url.Parse(svr.URL)
	require.NoError(t, err)

	maxRetries = 2
	retryDelay = time.Millisecond

	var cfg Config
	logger := log.NewTestLogger(t)
	s, err := New(logger, cfg)
	require.NoError(t, err)

	// download works after 2 retries
	b, urActual, err := s.downloadURLWithRetries(ctx, ur)
	require.NoError(t, err)
	require.NotNil(t, urActual)
	assert.Equal(t, svr.URL, urActual.String())
	assert.Equal(t, expected, string(b))
	assert.Equal(t, retry, maxRetries)

	// download fails after 3 retries
	retry = -100
	_, _, err = s.downloadURLWithRetries(ctx, ur)
	assert.ErrorIs(t, err, errExhaustedRetries)
}
