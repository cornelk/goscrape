package htmlindex

import (
	"net/url"

	"github.com/cornelk/gotokit/log"
	"golang.org/x/net/html"
)

type nodeAttributeParserData struct {
	logger    *log.Logger
	url       *url.URL
	node      *html.Node
	attribute string
	value     string
}

// nodeAttributeParser returns the URL values of the attribute of the node and
// whether the attribute has been processed.
type nodeAttributeParser func(data nodeAttributeParserData) ([]string, bool)

// Node describes an HTML tag and its attributes that can contain URLs.
type Node struct {
	Attributes []string

	noChildParsing bool
	parser         nodeAttributeParser
}

// nolint: revive
const (
	BackgroundAttribute = "background"
	HrefAttribute       = "href"

	DataSrcAttribute = "data-src"
	SrcAttribute     = "src"

	DataSrcSetAttribute = "data-srcset"
	SrcSetAttribute     = "srcset"
)

// nolint: revive
const (
	ATag      = "a"
	BodyTag   = "body"
	ImgTag    = "img"
	LinkTag   = "link"
	ScriptTag = "script"
	StyleTag  = "style"
)

// Nodes describes the HTML tags and their attributes that can contain URL.
var Nodes = map[string]Node{
	ATag: {
		Attributes: []string{HrefAttribute},
	},
	BodyTag: {
		Attributes: []string{BackgroundAttribute},
	},
	ImgTag: {
		Attributes: []string{SrcAttribute, DataSrcAttribute, SrcSetAttribute, DataSrcSetAttribute},
		parser:     srcSetValueSplitter,
	},
	LinkTag: {
		Attributes: []string{HrefAttribute},
	},
	ScriptTag: {
		Attributes: []string{SrcAttribute},
	},
	StyleTag: {
		noChildParsing: true,
		parser:         styleParser,
	},
}

// SrcSetAttributes contains the attributes that contain srcset values.
var SrcSetAttributes = map[string]struct{}{
	DataSrcSetAttribute: {},
	SrcSetAttribute:     {},
}
