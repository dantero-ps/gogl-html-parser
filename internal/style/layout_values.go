package style

import (
	"goglweb/internal/parser/html"
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
)

func (sn *StyledNode) Display() Display {
	d, ok := sn.SpecifiedValues["display"]
	if !ok {
		return Inline
	}
	return Display(d)
}
