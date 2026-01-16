package html

type NodeType int

const (
	TextNode NodeType = iota
	ElementNode
)

type Node struct {
	Type     NodeType
	TagName  string
	Content  string
	Attr     map[string]string
	Children []*Node
}

func NewElement(tagName string) *Node {
	return &Node{
		Type:    ElementNode,
		TagName: tagName,
		Attr:    make(map[string]string),
	}
}

func NewText(content string) *Node {
	return &Node{
		Type:    TextNode,
		Content: content,
	}
}
