package dom

import (
	"goglweb/internal/layout"
	"goglweb/internal/parser/html"
)

// HitTestResult represents the result of a hit test
type HitTestResult struct {
	LayoutBox *layout.LayoutBox
	HTMLNode  *html.Node
	Point     struct {
		X, Y float64
	}
}

// HitTest finds which layout box contains a point
// Returns the topmost element according to z-order
func HitTest(layoutBox *layout.LayoutBox, x, y float64) *HitTestResult {
	if layoutBox == nil {
		return nil
	}

	// First check children (z-order: children on top)
	for i := len(layoutBox.Children) - 1; i >= 0; i-- {
		child := layoutBox.Children[i]
		result := HitTest(child, x, y)
		if result != nil {
			return result
		}
	}

	// If not inside any child, check this box
	if isPointInBox(layoutBox, x, y) {
		var htmlNode *html.Node
		if layoutBox.StyledNode != nil {
			htmlNode = layoutBox.StyledNode.Node
			// If it's a text node, find parent element
			if htmlNode != nil && htmlNode.Type == html.TextNode {
				// Find parent element containing the text node
				// Since there's no parent pointer in HTML tree, we should search from root
				// However, HitTest function doesn't know the root. So to find the element box
				// containing the text node, we should search backwards in the layout tree.
				// Since there's no parent pointer, we ignore text nodes and only return element nodes.
				// For text click events, they need to be mapped to parent element.
				// This can be done by the caller of HitTest.
				return nil
			}
		}
		// Only return element nodes
		if htmlNode != nil && htmlNode.Type == html.ElementNode {
			return &HitTestResult{
				LayoutBox: layoutBox,
				HTMLNode:  htmlNode,
				Point:     struct{ X, Y float64 }{X: x, Y: y},
			}
		}
	}

	return nil
}

// isPointInBox checks if a point is inside a box
func isPointInBox(box *layout.LayoutBox, x, y float64) bool {
	if box == nil {
		return false
	}

	// Use border box (margin excluded, border included)
	borderBox := box.Dimensions.BorderBox()

	return x >= borderBox.X &&
		x <= borderBox.X+borderBox.Width &&
		y >= borderBox.Y &&
		y <= borderBox.Y+borderBox.Height
}

// FindLayoutBoxByNode finds the layout box corresponding to an HTML node
func FindLayoutBoxByNode(root *layout.LayoutBox, targetNode *html.Node) *layout.LayoutBox {
	if root == nil || targetNode == nil {
		return nil
	}

	// Does this box's node match?
	if root.StyledNode != nil && root.StyledNode.Node == targetNode {
		return root
	}

	// Search in children
	for _, child := range root.Children {
		result := FindLayoutBoxByNode(child, targetNode)
		if result != nil {
			return result
		}
	}

	return nil
}
