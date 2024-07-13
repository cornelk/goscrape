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

var Nodes = map[string]Node{
	"a": {
		Attributes: []string{HrefAttribute},
	},
	"body": {
		Attributes: []string{BackgroundAttribute},
	},
	"img": {
		Attributes: []string{SrcAttribute, DataSrcAttribute, SrcSetAttribute, DataSrcSetAttribute},
		parser:     srcSetValueSplitter,
	},
	"link": {
		Attributes: []string{HrefAttribute},
	},
	"script": {
		Attributes: []string{SrcAttribute},
	},
}
