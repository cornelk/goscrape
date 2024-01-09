package scraper

import (
	"bytes"
	"testing"

	"github.com/cornelk/goscrape/htmlindex"
	"github.com/cornelk/gotokit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
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

	doc, err := html.Parse(buf)
	require.NoError(t, err)

	index := htmlindex.New()
	index.Index(s.URL, doc)

	html, fixed, err := s.fixURLReferences(s.URL, doc, index)
	require.NoError(t, err)
	assert.True(t, fixed)

	expected := "<html lang=\"es\"><head></head><body><a href=\"wp-content/uploads/document.pdf\" rel=\"doc\">Guide</a>\n\n</body></html>"
	assert.Equal(t, expected, html)
}
