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

	idx := testSetup(t, input)

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

func TestIndexImg(t *testing.T) {
	input := []byte(`
<html lang="es">
<body background="bg.jpg"></body>
<img src="test.jpg" srcset="test-480w.jpg 480w, test-800w.jpg 800w"/> 
</body>
</html>
`)

	idx := testSetup(t, input)
	references, err := idx.URLs(ImgTag)
	require.NoError(t, err)
	require.Len(t, references, 3)
	assert.Equal(t, "https://domain.com/test-480w.jpg", references[0].String())
	assert.Equal(t, "https://domain.com/test-800w.jpg", references[1].String())
	assert.Equal(t, "https://domain.com/test.jpg", references[2].String())

	references, err = idx.URLs(BodyTag)
	require.NoError(t, err)
	require.Len(t, references, 1)
	assert.Equal(t, "https://domain.com/bg.jpg", references[0].String())
}

func testSetup(t *testing.T, input []byte) *Index {
	t.Helper()

	buf := &bytes.Buffer{}
	_, err := buf.Write(input)
	require.NoError(t, err)

	doc, err := html.Parse(buf)
	require.NoError(t, err)

	ur, err := url.Parse("https://domain.com/")
	require.NoError(t, err)

	idx := New()
	idx.Index(ur, doc)

	return idx
}
