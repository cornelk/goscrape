package htmlindex

// nodeAttributeParser returns the URL values of the attribute of the node and
// whether the attribute has been processed.
type nodeAttributeParser func(attribute, value string) ([]string, bool)

type Node struct {
	Attributes []string

	parser nodeAttributeParser
}

const (
	BackgroundAttribute = "background"
	HrefAttribute       = "href"

	DataSrcAttribute = "data-src"
	SrcAttribute     = "src"

	DataSrcSetAttribute = "data-srcset"
	SrcSetAttribute     = "srcset"
)

const (
	ATag      = "a"
	BodyTag   = "body"
	ImgTag    = "img"
	LinkTag   = "link"
	ScriptTag = "script"
)

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
}
