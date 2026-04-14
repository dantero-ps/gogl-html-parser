package render

import (
	"github.com/furkandgn/goglweb/internal/layout"
	"github.com/furkandgn/goglweb/internal/parser/html"
	"github.com/furkandgn/goglweb/internal/style"
	"testing"
)

// TestScrollCommandsEmitted verifies that a scrollable container emits
// BeginScrollCommand and EndScrollCommand in the display list.
func TestScrollCommandsEmitted(t *testing.T) {
	// Build a scrollable LayoutBox manually with proper HTML nodes
	childHTMLNode := &html.Node{TagName: "div", Type: html.ElementNode}
	scrollableHTMLNode := &html.Node{TagName: "div", Type: html.ElementNode}

	childStyled := &style.StyledNode{
		Node: childHTMLNode,
		SpecifiedValues: map[string]string{
			"display":          "block",
			"width":            "50px",
			"height":           "50px",
			"background-color": "#ff0000",
		},
	}
	childBox := &layout.LayoutBox{
		BoxType:    layout.BlockBox,
		StyledNode: childStyled,
		Dimensions: layout.Dimensions{
			Content: layout.Rect{X: 0, Y: 0, Width: 50, Height: 50},
		},
	}

	scrollableStyled := &style.StyledNode{
		Node: scrollableHTMLNode,
		SpecifiedValues: map[string]string{
			"display":          "block",
			"width":            "200px",
			"height":           "200px",
			"overflow":         "scroll",
			"background-color": "#00ff00",
		},
	}
	scrollableBox := &layout.LayoutBox{
		BoxType:    layout.BlockBox,
		StyledNode: scrollableStyled,
		Dimensions: layout.Dimensions{
			Content: layout.Rect{X: 0, Y: 0, Width: 200, Height: 200},
		},
		Overflow:      "scroll",
		ScrollOffsetX: 0,
		ScrollOffsetY: 10,
		Children:      []*layout.LayoutBox{childBox},
	}
	childBox.Parent = scrollableBox

	// Build display list
	dl := BuildDisplayList(scrollableBox)

	// Execute with MockPainter
	painter := &MockPainter{}
	dl.Execute(painter)

	// Verify BeginScroll and EndScroll were called
	if painter.BeginScrollCalls != 1 {
		t.Errorf("Expected 1 BeginScroll call, got %d", painter.BeginScrollCalls)
	}
	if painter.EndScrollCalls != 1 {
		t.Errorf("Expected 1 EndScroll call, got %d", painter.EndScrollCalls)
	}
}

// TestNoScrollCommandsForNonScrollable verifies that non-scrollable boxes
// do not emit scroll commands.
func TestNoScrollCommandsForNonScrollable(t *testing.T) {
	boxHTMLNode := &html.Node{TagName: "div", Type: html.ElementNode}
	boxStyled := &style.StyledNode{
		Node: boxHTMLNode,
		SpecifiedValues: map[string]string{
			"display":          "block",
			"width":            "200px",
			"height":           "200px",
			"background-color": "#00ff00",
		},
	}
	box := &layout.LayoutBox{
		BoxType:    layout.BlockBox,
		StyledNode: boxStyled,
		Dimensions: layout.Dimensions{
			Content: layout.Rect{X: 0, Y: 0, Width: 200, Height: 200},
		},
		Overflow: "visible",
	}

	dl := BuildDisplayList(box)
	painter := &MockPainter{}
	dl.Execute(painter)

	if painter.BeginScrollCalls != 0 {
		t.Errorf("Expected 0 BeginScroll calls for non-scrollable, got %d", painter.BeginScrollCalls)
	}
	if painter.EndScrollCalls != 0 {
		t.Errorf("Expected 0 EndScroll calls for non-scrollable, got %d", painter.EndScrollCalls)
	}
}
