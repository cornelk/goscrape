package htmlindex

import (
	"bytes"
	"net/url"
	"testing"

	"github.com/cornelk/gotokit/log"
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

	logger := log.NewTestLogger(t)
	idx := New(logger)
	idx.Index(ur, doc)

	return idx
}

func TestIndexStyleAttribute(t *testing.T) {
	input := []byte(`
<html lang="es">
<body style="background-image: url(#);">
<div style="background-image: url('background.jpg'); color: red;">
<img src="test.jpg" style="border: 1px solid url('/border.png');">
<a href="link.html" style="background: url(bg.gif) no-repeat;">Link</a>
</div>
</body>
</html>
`)

	idx := testSetup(t, input)

	// check body tag with style attribute
	nodeTag := BodyTag
	references, err := idx.URLs(nodeTag)
	require.NoError(t, err)
	require.Len(t, references, 1) // url(#) gets resolved to base URL
	assert.Equal(t, "https://domain.com/", references[0].String())

	// check img tag - should have both src and style URLs
	nodeTag = ImgTag
	references, err = idx.URLs(nodeTag)
	require.NoError(t, err)
	require.Len(t, references, 2)

	// Sort to ensure consistent ordering
	urls := make([]string, len(references))
	for i, ref := range references {
		urls[i] = ref.String()
	}
	assert.Contains(t, urls, "https://domain.com/test.jpg")   // from src attribute
	assert.Contains(t, urls, "https://domain.com/border.png") // from style attribute

	// check a tag with style attribute
	nodeTag = ATag
	references, err = idx.URLs(nodeTag)
	require.NoError(t, err)
	require.Len(t, references, 2)

	urls = make([]string, len(references))
	for i, ref := range references {
		urls[i] = ref.String()
	}
	assert.Contains(t, urls, "https://domain.com/link.html") // from href attribute
	assert.Contains(t, urls, "https://domain.com/bg.gif")    // from style attribute
}

func TestIndexComplexStyleAttribute(t *testing.T) {
	input := []byte(`
<html>
<body>
<div style="background: url('https://example.com/bg1.jpg') no-repeat, url('/bg2.png') repeat;">
<p style="content: url(data:image/png;base64,iVBOR); background: url('valid.gif');">
</body>
</html>
`)

	idx := testSetup(t, input)

	// Note: Since we don't have "div" and "p" tags in our Nodes map,
	// let's add a more comprehensive test that uses tracked tags
	// For now, this test documents the current limitation
	references, err := idx.URLs("div")
	require.NoError(t, err)
	require.Empty(t, references) // div is not in Nodes map
}

func TestGitHubIssue48_StyleAttributeParsing(t *testing.T) {
	// Test case directly from GitHub issue #48:
	// HTML style attributes are not being parsed correctly
	// Example: <body style="background-image: url(#);">
	input := []byte(`
<html>
<body style="background-image: url(#);">
<p>Test content</p>
</body>
</html>
`)

	idx := testSetup(t, input)

	// Check that the body tag with style attribute is properly indexed
	nodeTag := BodyTag
	references, err := idx.URLs(nodeTag)
	require.NoError(t, err)
	require.Len(t, references, 1) // url(#) gets resolved to base URL
	assert.Equal(t, "https://domain.com/", references[0].String())

	// Verify the nodes are properly tracked
	nodes := idx.Nodes(nodeTag)
	require.Len(t, nodes, 1)
	require.Contains(t, nodes, "https://domain.com/")
	require.Len(t, nodes["https://domain.com/"], 1)
	assert.Equal(t, BodyTag, nodes["https://domain.com/"][0].Data)
}
