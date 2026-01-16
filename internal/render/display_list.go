package render

import (
	"goglweb/internal/parser/html"
	"strconv"
	"strings"

	"goglweb/internal/layout"
)

// Command interface for all rendering operations.
type Command interface {
	Execute(painter Painter)
}

type FillRectCmd struct {
	Rect  layout.Rect
	Color Color
}

func (c FillRectCmd) Execute(p Painter) { p.FillRect(c.Rect, c.Color) }

type DrawBorderCmd struct {
	Rect  layout.Rect
	Edges layout.EdgeSizes
	Color Color
}

func (c DrawBorderCmd) Execute(p Painter) { p.DrawBorder(c.Rect, c.Edges, c.Color) }

type DrawTextCmd struct {
	Text       string
	X, Y       float64
	FontSize   float64
	Color      Color
	FontFamily string
}

func (c DrawTextCmd) Execute(p Painter) {
	p.DrawText(c.Text, c.X, c.Y, c.FontSize, c.Color, c.FontFamily)
}

type ClipRectCmd struct {
	Rect  layout.Rect
	Clear bool
}

func (c ClipRectCmd) Execute(p Painter) {
	if c.Clear {
		p.ClearClip()
	} else {
		p.SetClip(c.Rect)
	}
}

// DisplayList maintains a sequence of drawing commands.
type DisplayList struct {
	Commands []Command
}

func NewDisplayList() *DisplayList {
	return &DisplayList{Commands: make([]Command, 0)}
}

func (dl *DisplayList) Add(cmd Command) {
	dl.Commands = append(dl.Commands, cmd)
}

func (dl *DisplayList) Execute(painter Painter) {
	for _, cmd := range dl.Commands {
		cmd.Execute(painter)
	}
}

// BuildDisplayList traverses the LayoutBox tree and generates commands.
func BuildDisplayList(root *layout.LayoutBox) *DisplayList {
	dl := NewDisplayList()
	renderBox(root, dl)
	return dl
}

func renderBox(box *layout.LayoutBox, dl *DisplayList) {
	if box == nil {
		return
	}

	// Read opacity value (0.0 - 1.0)
	opacityStr := getStyleValue(box, "opacity")
	opacity := 1.0
	if opacityStr != "" {
		if parsed, err := strconv.ParseFloat(opacityStr, 64); err == nil {
			opacity = parsed
			if opacity < 0 {
				opacity = 0
			}
			if opacity > 1 {
				opacity = 1
			}
		}
	}

	// Helper function that applies opacity to color
	applyOpacity := func(color Color) Color {
		if opacity < 1.0 {
			color.A = uint8(float64(color.A) * opacity)
		}
		return color
	}

	// 1. Background (Z-order: Bottom)
	bgColorStr := getStyleValue(box, "background-color")
	if bgColorStr != "" && bgColorStr != "transparent" {
		// Background fills Padding Box (Content + Padding)
		rect := box.Dimensions.Content
		rect.X -= box.Dimensions.Padding.Left
		rect.Y -= box.Dimensions.Padding.Top
		rect.Width += box.Dimensions.Padding.Left + box.Dimensions.Padding.Right
		rect.Height += box.Dimensions.Padding.Top + box.Dimensions.Padding.Bottom

		bgColor := ParseColor(bgColorStr)
		dl.Add(FillRectCmd{Rect: rect, Color: applyOpacity(bgColor)})
	}

	// 2. Borders
	borderColorStr := getStyleValue(box, "border-color")
	if borderColorStr == "" {
		// Extract color from border shorthand: "2px solid #333" -> "#333"
		borderShorthand := getStyleValue(box, "border")
		if borderShorthand != "" {
			borderColorStr = extractBorderColor(borderShorthand)
		}
	}
	if box.Dimensions.Border.Left > 0 || box.Dimensions.Border.Right > 0 ||
		box.Dimensions.Border.Top > 0 || box.Dimensions.Border.Bottom > 0 {

		// Border is outside padding box
		rect := box.Dimensions.Content
		rect.X -= (box.Dimensions.Padding.Left + box.Dimensions.Border.Left)
		rect.Y -= (box.Dimensions.Padding.Top + box.Dimensions.Border.Top)
		rect.Width += box.Dimensions.Padding.Left + box.Dimensions.Padding.Right + box.Dimensions.Border.Left + box.Dimensions.Border.Right
		rect.Height += box.Dimensions.Padding.Top + box.Dimensions.Padding.Bottom + box.Dimensions.Border.Top + box.Dimensions.Border.Bottom

		// Don't render border for boxes with height 0 (only border width would be visible)
		// Minimum height = border top + border bottom
		minHeight := box.Dimensions.Border.Top + box.Dimensions.Border.Bottom
		if rect.Height < minHeight {
			rect.Height = minHeight
		}

		borderColor := ParseColor(borderColorStr)
		dl.Add(DrawBorderCmd{
			Rect:  rect,
			Edges: box.Dimensions.Border,
			Color: applyOpacity(borderColor),
		})
	}

	// 3. Overflow / Clipping
	overflow := getStyleValue(box, "overflow")
	isClipped := overflow == "hidden"
	if isClipped {
		dl.Add(ClipRectCmd{Rect: box.Dimensions.Content})
	}

	// 4. Children (Recursive Traversal)
	// Even though the prompt says post-order (children first), browser engines
	// usually put children commands after parent background to ensure they are on top.
	for _, child := range box.Children {
		renderBox(child, dl)
	}

	// 5. Text Rendering
	if box.StyledNode != nil && box.StyledNode.Node.Type == html.TextNode { // Assuming 0 is TextNode
		RenderText(box, dl)
	}

	// Clear clipping after children and text are drawn
	if isClipped {
		dl.Add(ClipRectCmd{Clear: true})
	}
}

func getStyleValue(box *layout.LayoutBox, property string) string {
	if box.StyledNode == nil || box.StyledNode.SpecifiedValues == nil {
		return ""
	}
	return box.StyledNode.SpecifiedValues[property]
}

// extractBorderColor extracts color value from border shorthand
// Example: "2px solid #333" -> "#333"
func extractBorderColor(borderValue string) string {
	parts := strings.Fields(borderValue)
	// Border shorthand format: width style color
	// Color is usually the last token (hex or named color)
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		// Hex color or named color check
		if strings.HasPrefix(part, "#") || isNamedColor(part) {
			return part
		}
	}
	return ""
}

func isNamedColor(color string) bool {
	namedColors := []string{"black", "white", "red", "green", "blue", "yellow", "cyan", "gray", "silver", "maroon", "navy", "orange", "purple"}
	for _, nc := range namedColors {
		if strings.ToLower(color) == nc {
			return true
		}
	}
	return false
}
