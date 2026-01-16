package render

import (
	"goglweb/internal/layout"
	"goglweb/internal/parser/html"
	"goglweb/internal/style"
	"strings"
	"testing"
)

func TestColorParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected Color
	}{
		{"red", Color{255, 0, 0, 255}},
		{"#00ff00", Color{0, 255, 0, 255}},
		{"#00f", Color{0, 0, 255, 255}},
		{"rgb(255, 255, 255)", Color{255, 255, 255, 255}},
		{"rgba(0, 0, 0, 0.5)", Color{0, 0, 0, 127}},
	}

	for _, tc := range tests {
		result := ParseColor(tc.input)
		if result != tc.expected {
			t.Errorf("For %s expected %+v, got %+v", tc.input, tc.expected, result)
		}
	}
}

func TestBuildDisplayList(t *testing.T) {
	// Setup a simple box
	sn := &style.StyledNode{
		Node: &html.Node{Type: 1}, // Element
		SpecifiedValues: map[string]string{
			"background-color": "red",
			"border-color":     "black",
		},
	}

	box := &layout.LayoutBox{
		StyledNode: sn,
		Dimensions: layout.Dimensions{
			Content: layout.Rect{X: 10, Y: 10, Width: 100, Height: 100},
			Border:  layout.EdgeSizes{Left: 2, Right: 2, Top: 2, Bottom: 2},
		},
	}

	dl := BuildDisplayList(box)
	painter := &MockPainter{}
	dl.Execute(painter)

	if len(painter.Operations) < 2 {
		t.Errorf("Expected at least 2 operations (Background + Border), got %d", len(painter.Operations))
	}
}

func TestTextRendering(t *testing.T) {
	sn := &style.StyledNode{
		Node: &html.Node{Type: 0, Content: "Hello World"},
		SpecifiedValues: map[string]string{
			"color":     "blue",
			"font-size": "20px",
		},
	}

	box := &layout.LayoutBox{
		StyledNode: sn,
		Dimensions: layout.Dimensions{
			Content: layout.Rect{X: 50, Y: 50, Width: 200, Height: 20},
		},
	}

	dl := BuildDisplayList(box)
	painter := &MockPainter{}
	dl.Execute(painter)

	foundText := false
	for _, op := range painter.Operations {
		if strings.Contains(op, "DrawText: 'Hello World'") {
			foundText = true
			break
		}
	}

	if !foundText {
		t.Error("DrawText command not found in painter operations")
	}
}
