package scraper

import (
	"net/url"
	"testing"

	"github.com/cornelk/gotokit/log"
	"github.com/cornelk/gotokit/set"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeURLPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "/"},
		{"/", "/"},
		{"/path", "/path"},
		{"/path/", "/path"},
		{"/path/to/resource", "/path/to/resource"},
		{"/path/to/resource/", "/path/to/resource"},
		{"/category/blog-post", "/category/blog-post"},
		{"/category/blog-post/", "/category/blog-post"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := normalizeURLPath(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestShouldURLBeDownloaded_TrailingSlashDuplicates(t *testing.T) {
	logger := log.NewTestLogger(t)
	cfg := Config{
		URL: "https://example.com",
	}
	scraper, err := New(logger, cfg)
	require.NoError(t, err)
	require.NotNil(t, scraper)

	// Initialize empty processed set
	scraper.processed = set.New[string]()

	// Test that URLs with and without trailing slashes are treated as duplicates
	url1, err := url.Parse("https://example.com/category/blog-post")
	require.NoError(t, err)

	url2, err := url.Parse("https://example.com/category/blog-post/")
	require.NoError(t, err)

	// First URL should be downloadable
	should1 := scraper.shouldURLBeDownloaded(url1, 0, false)
	assert.True(t, should1, "First URL should be downloadable")

	// Second URL with trailing slash should be treated as duplicate
	should2 := scraper.shouldURLBeDownloaded(url2, 0, false)
	assert.False(t, should2, "Second URL with trailing slash should be treated as duplicate")

	// Verify that the normalized path is in the processed set
	assert.True(t, scraper.processed.Contains("/category/blog-post"))
}

func TestShouldURLBeDownloaded_TrailingSlashDuplicatesReverse(t *testing.T) {
	logger := log.NewTestLogger(t)
	cfg := Config{
		URL: "https://example.com",
	}
	scraper, err := New(logger, cfg)
	require.NoError(t, err)
	require.NotNil(t, scraper)

	// Initialize empty processed set
	scraper.processed = set.New[string]()

	// Test reverse order - trailing slash first, then without
	url1, err := url.Parse("https://example.com/category/blog-post/")
	require.NoError(t, err)

	url2, err := url.Parse("https://example.com/category/blog-post")
	require.NoError(t, err)

	// First URL with trailing slash should be downloadable
	should1 := scraper.shouldURLBeDownloaded(url1, 0, false)
	assert.True(t, should1, "First URL with trailing slash should be downloadable")

	// Second URL without trailing slash should be treated as duplicate
	should2 := scraper.shouldURLBeDownloaded(url2, 0, false)
	assert.False(t, should2, "Second URL without trailing slash should be treated as duplicate")

	// Verify that the normalized path is in the processed set
	assert.True(t, scraper.processed.Contains("/category/blog-post"))
}

func TestShouldURLBeDownloaded_RootPath(t *testing.T) {
	logger := log.NewTestLogger(t)
	cfg := Config{
		URL: "https://example.com",
	}
	scraper, err := New(logger, cfg)
	require.NoError(t, err)
	require.NotNil(t, scraper)

	// Initialize empty processed set
	scraper.processed = set.New[string]()

	// Test root path normalization
	url1, err := url.Parse("https://example.com/")
	require.NoError(t, err)

	url2, err := url.Parse("https://example.com")
	require.NoError(t, err)

	// First root URL should be downloadable
	should1 := scraper.shouldURLBeDownloaded(url1, 0, false)
	assert.True(t, should1, "First root URL should be downloadable")

	// Second root URL should be treated as duplicate
	should2 := scraper.shouldURLBeDownloaded(url2, 0, false)
	assert.False(t, should2, "Second root URL should be treated as duplicate")

	// Verify that the normalized root path is in the processed set
	assert.True(t, scraper.processed.Contains("/"))
}

func TestShouldURLBeDownloaded_ExternalURLs(t *testing.T) {
	logger := log.NewTestLogger(t)
	cfg := Config{
		URL: "https://example.com",
	}
	scraper, err := New(logger, cfg)
	require.NoError(t, err)
	require.NotNil(t, scraper)

	// Initialize empty processed set
	scraper.processed = set.New[string]()

	// Test external URLs with trailing slashes as assets
	url1, err := url.Parse("https://external.com/path.css")
	require.NoError(t, err)

	url2, err := url.Parse("https://external.com/path.css/")
	require.NoError(t, err)

	// First external asset should be downloadable (if it passes other checks)
	should1 := scraper.shouldURLBeDownloaded(url1, 0, true) // asset = true

	// Second external asset with trailing slash should be treated as duplicate
	should2 := scraper.shouldURLBeDownloaded(url2, 0, true) // asset = true

	// First should pass, second should be blocked as duplicate
	assert.True(t, should1, "First external asset should be downloadable")
	assert.False(t, should2, "Second external asset with trailing slash should be treated as duplicate")

	// Verify that the normalized external URL is in the processed set
	normalizedURL1 := normalizeURLPath(url1.String())
	assert.True(t, scraper.processed.Contains(normalizedURL1))
}
