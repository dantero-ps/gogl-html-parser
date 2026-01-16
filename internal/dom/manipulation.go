package dom

import (
	"goglweb/internal/parser/html"
	"strings"
)

// Helper functions for DOM manipulation

// AppendChild adds a node to the parent's children list
func AppendChild(parent, child *html.Node) {
	if parent == nil || child == nil {
		return
	}
	parent.Children = append(parent.Children, child)
}

// RemoveChild removes a node from the parent's children list
func RemoveChild(parent, child *html.Node) bool {
	if parent == nil || child == nil {
		return false
	}
	for i, c := range parent.Children {
		if c == child {
			parent.Children = append(parent.Children[:i], parent.Children[i+1:]...)
			return true
		}
	}
	return false
}

// SetAttribute adds or updates an attribute on a node
func SetAttribute(node *html.Node, key, value string) {
	if node == nil {
		return
	}
	if node.Attr == nil {
		node.Attr = make(map[string]string)
	}
	node.Attr[key] = value
}

// GetAttribute returns a node's attribute value
func GetAttribute(node *html.Node, key string) string {
	if node == nil || node.Attr == nil {
		return ""
	}
	return node.Attr[key]
}

// RemoveAttribute removes an attribute from a node
func RemoveAttribute(node *html.Node, key string) {
	if node == nil || node.Attr == nil {
		return
	}
	delete(node.Attr, key)
}

// SetTextContent sets the content of a text node or the first text child of an element node
func SetTextContent(node *html.Node, text string) {
	if node == nil {
		return
	}
	if node.Type == html.TextNode {
		node.Content = text
		return
	}
	// If element node, remove existing text children and add a new text node
	// First remove existing text nodes
	newChildren := []*html.Node{}
	for _, child := range node.Children {
		if child.Type != html.TextNode {
			newChildren = append(newChildren, child)
		}
	}
	node.Children = newChildren
	// Add new text node (append to end, to preserve original order)
	if text != "" {
		textNode := html.NewText(text)
		node.Children = append(node.Children, textNode)
	}
}

// GetTextContent returns a node's text content (iterative version)
func GetTextContent(node *html.Node) string {
	if node == nil {
		return ""
	}
	if node.Type == html.TextNode {
		return node.Content
	}

	// If element node, iteratively combine all text children's content
	result := ""
	stack := []*html.Node{node}

	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		for _, child := range current.Children {
			if child.Type == html.TextNode {
				result += child.Content
			} else {
				stack = append(stack, child)
			}
		}
	}

	return result
}

// AddClass adds a class to a node
func AddClass(node *html.Node, className string) {
	if node == nil {
		return
	}
	classes := GetClasses(node)
	for _, c := range classes {
		if c == className {
			return // Zaten var
		}
	}
	classes = append(classes, className)
	SetAttribute(node, "class", joinClasses(classes))
}

// RemoveClass removes a class from a node
func RemoveClass(node *html.Node, className string) {
	if node == nil {
		return
	}
	classes := GetClasses(node)
	newClasses := []string{}
	for _, c := range classes {
		if c != className {
			newClasses = append(newClasses, c)
		}
	}
	SetAttribute(node, "class", joinClasses(newClasses))
}

// ToggleClass toggles a class on a node
func ToggleClass(node *html.Node, className string) {
	if node == nil {
		return
	}
	classes := GetClasses(node)
	found := false
	for _, c := range classes {
		if c == className {
			found = true
			break
		}
	}
	if found {
		RemoveClass(node, className)
	} else {
		AddClass(node, className)
	}
}

// GetClasses returns a node's class list
func GetClasses(node *html.Node) []string {
	if node == nil {
		return []string{}
	}
	classAttr := GetAttribute(node, "class")
	if classAttr == "" {
		return []string{}
	}
	// Use strings.Fields (automatically handles whitespace, filters empty strings)
	return strings.Fields(classAttr)
}

func joinClasses(classes []string) string {
	result := ""
	for i, c := range classes {
		if i > 0 {
			result += " "
		}
		result += c
	}
	return result
}
