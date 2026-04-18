package style

import (
	"github.com/furkandgn/goglweb/internal/parser/html"
)

type PropertyMap map[string]string

type StyledNode struct {
	Node            *html.Node
	SpecifiedValues PropertyMap
	Children        []*StyledNode
}

type Display string

const (
	Block  Display = "block"
	Inline Display = "inline"
	None   Display = "none"
	Flex   Display = "flex"
)

// blockLevelTags are HTML elements that default to display:block per the HTML spec.
var blockLevelTags = map[string]bool{
	"div": true, "p": true, "section": true, "article": true,
	"header": true, "footer": true, "main": true, "nav": true, "aside": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"ul": true, "ol": true, "li": true, "dl": true, "dt": true, "dd": true,
	"blockquote": true, "figure": true, "figcaption": true, "details": true,
	"summary": true, "form": true, "fieldset": true, "table": true, "hr": true,
	"pre": true, "address": true,
}

func (sn *StyledNode) Display() Display {
	d, ok := sn.SpecifiedValues["display"]
	if !ok {
		if sn.Node != nil && blockLevelTags[sn.Node.TagName] {
			return Block
		}
		return Inline
	}
	return Display(d)
}
