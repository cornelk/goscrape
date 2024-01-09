package htmlindex

import (
	"bytes"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestIndex(t *testing.T) {
	input := []byte(`
<html lang="es">
<a href="https://domain.com/wp-content/uploads/document.pdf" rel="doc">Guide</a>
<img src="/test.jpg"/> 
</html>
`)

	buf := &bytes.Buffer{}
	_, err := buf.Write(input)
	require.NoError(t, err)

	doc, err := html.Parse(buf)
	require.NoError(t, err)

	ur, err := url.Parse("https://domain.com/")
	require.NoError(t, err)

	idx := New()
	idx.Index(ur, doc)

	// check a tag
	nodeTag := "a"
	references, err := idx.URLs(nodeTag)
	require.NoError(t, err)
	require.Len(t, references, 1)

	tagURL := "https://domain.com/wp-content/uploads/document.pdf"
	assert.Equal(t, tagURL, references[0].String())

	urls := idx.Nodes(nodeTag)
	require.Len(t, urls, 1)
	nodes, ok := urls[tagURL]
	require.True(t, ok)
	require.Len(t, nodes, 1)
	node := nodes[0]
	assert.Equal(t, nodeTag, node.Data)

	// check img tag
	nodeTag = "img"
	references, err = idx.URLs(nodeTag)
	require.NoError(t, err)
	require.Len(t, references, 1)

	tagURL = "https://domain.com/test.jpg"
	assert.Equal(t, tagURL, references[0].String())

	// check for not existing tag
	nodeTag = "not-existing"
	references, err = idx.URLs(nodeTag)
	require.NoError(t, err)
	require.Empty(t, references)
	urls = idx.Nodes(nodeTag)
	require.Empty(t, urls)
}
