package style

import (
	"github.com/furkandgn/goglweb/internal/parser/css"
	"github.com/furkandgn/goglweb/internal/parser/html"
)

func ComputeStyle(node *html.Node, stylesheet *css.Stylesheet) PropertyMap {
	props := make(PropertyMap)
	for _, rule := range stylesheet.Rules {
		if matches(node, rule.Selector) {
			for _, decl := range rule.Declarations {
				props[decl.Property] = decl.Value
			}
		}
	}
	return props
}

// Inheritable properties - properties that are automatically inherited in CSS
var inheritableProperties = map[string]bool{
	"color":           true,
	"font-family":     true,
	"font-size":       true,
	"font-weight":     true,
	"font-style":      true,
	"line-height":     true,
	"text-align":      true,
	"text-decoration": true,
	"visibility":      true,
}

func BuildStyledTree(node *html.Node, stylesheet *css.Stylesheet) *StyledNode {
	return buildStyledTreeRecursive(node, stylesheet, nil)
}

func buildStyledTreeRecursive(node *html.Node, stylesheet *css.Stylesheet, parentProps PropertyMap) *StyledNode {
	if node.Type == 0 {
		inheritedProps := make(PropertyMap)

		for prop, value := range parentProps {
			if inheritableProperties[prop] {
				inheritedProps[prop] = value
			}
		}
		return &StyledNode{Node: node, SpecifiedValues: inheritedProps}
	}

	styledNode := &StyledNode{
		Node:            node,
		SpecifiedValues: ComputeStyle(node, stylesheet),
	}

	for prop, value := range parentProps {
		if inheritableProperties[prop] {
			// If this property is not already defined, inherit from parent
			if _, exists := styledNode.SpecifiedValues[prop]; !exists {
				styledNode.SpecifiedValues[prop] = value
			}
		}
	}

	for _, child := range node.Children {
		// Don't include display: none in the tree
		childStyled := buildStyledTreeRecursive(child, stylesheet, styledNode.SpecifiedValues)
		if childStyled.Display() != None {
			styledNode.Children = append(styledNode.Children, childStyled)
		}
	}

	return styledNode
}
