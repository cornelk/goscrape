package scraper

import (
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/cornelk/gotokit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFilePath(t *testing.T) {
	type filePathFixture struct {
		BaseURL          string
		DownloadURL      string
		ExpectedFilePath string
	}

	pathSeparator := string(os.PathSeparator)
	expectedBasePath := "google.com" + pathSeparator
	var fixtures = []filePathFixture{
		{"https://google.com/", "https://github.com/", expectedBasePath + "_github.com" + pathSeparator + "index.html"},
		{"https://google.com/", "https://github.com/#fragment", expectedBasePath + "_github.com" + pathSeparator + "index.html"},
		{"https://google.com/", "https://github.com/test", expectedBasePath + "_github.com" + pathSeparator + "test.html"},
		{"https://google.com/", "https://github.com/test/", expectedBasePath + "_github.com" + pathSeparator + "test" + pathSeparator + "index.html"},
		{"https://google.com/", "https://github.com/test.aspx", expectedBasePath + "_github.com" + pathSeparator + "test.aspx"},
		{"https://google.com/", "https://google.com/settings", expectedBasePath + "settings.html"},
	}

	var cfg Config
	logger := log.NewTestLogger(t)
	for _, fix := range fixtures {
		cfg.URL = fix.BaseURL
		s, err := New(logger, cfg)
		require.NoError(t, err)

		URL, err := url.Parse(fix.DownloadURL)
		require.NoError(t, err)

		output := s.getFilePath(URL, true)
		assert.Equal(t, fix.ExpectedFilePath, output)
	}
}

func TestTruncateFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected func(string) bool // function to validate the result
	}{
		{
			name:     "short filename unchanged",
			filename: "test.css",
			expected: func(result string) bool { return result == "test.css" },
		},
		{
			name:     "long filename gets truncated",
			filename: "very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-long-filename.css",
			expected: func(result string) bool {
				return len(result) <= MaxFilenameLength &&
					len(result) > 0 &&
					result != "very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-long-filename.css" &&
					result[len(result)-4:] == ".css"
			},
		},
		{
			name:     "filename without extension",
			filename: "very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-long-filename",
			expected: func(result string) bool {
				return len(result) <= MaxFilenameLength &&
					len(result) > 0 &&
					result != "very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-very-long-filename"
			},
		},
		{
			name:     "empty filename",
			filename: "",
			expected: func(result string) bool { return result == "" },
		},
		{
			name:     "filename at max length",
			filename: strings.Repeat("a", MaxFilenameLength),
			expected: func(result string) bool { return len(result) == MaxFilenameLength },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateFilename(tt.filename)
			assert.True(t, tt.expected(result), "filename: %q, result: %q", tt.filename, result)
		})
	}
}

func TestTruncateFilenameUniqueness(t *testing.T) {
	// Test that different long filenames with the same prefix produce different results
	longPrefix := "this-is-a-very-long-filename-prefix-that-will-be-truncated-and-should-produce-different-results-based-on-the-hash-suffix-when-the-full-filename-is-different"

	filename1 := longPrefix + "-file1.css"
	filename2 := longPrefix + "-file2.css"

	result1 := truncateFilename(filename1)
	result2 := truncateFilename(filename2)

	assert.NotEqual(t, result1, result2, "Different long filenames should produce different truncated results")
	assert.LessOrEqual(t, len(result1), MaxFilenameLength, "Result1 should be within max length")
	assert.LessOrEqual(t, len(result2), MaxFilenameLength, "Result2 should be within max length")
}
