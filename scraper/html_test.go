package scraper

import (
	"bytes"
	"testing"

	"github.com/cornelk/gotokit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixFileReferences(t *testing.T) {
	logger := log.NewTestLogger(t)
	cfg := Config{
		URL: "http://domain.com",
	}
	s, err := New(logger, cfg)
	require.NoError(t, err)

	b := []byte(`
<html lang="es">
<a href="https://domain.com/wp-content/uploads/document.pdf" rel="doc">Guide</a>
</html>
`)

	buf := &bytes.Buffer{}
	_, err = buf.Write(b)
	require.NoError(t, err)

	html, fixed, err := s.fixURLReferences(s.URL, buf)
	require.NoError(t, err)
	assert.True(t, fixed)

	expected := "<html lang=\"es\"><head></head><body><a href=\"wp-content/uploads/document.pdf\" rel=\"doc\">Guide</a>\n\n</body></html>"
	assert.Equal(t, expected, html)
}
