package render

import (
	"fmt"
	"github.com/furkandgn/goglweb/internal/parser/html"
	"slices"
	"strconv"
	"strings"

	"github.com/furkandgn/goglweb/internal/layout"
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
	Text           string
	X, Y           float64
	FontSize       float64
	Color          Color
	FontFamily     string
	TextAlign      string
	ContainerWidth float64
	FontWeight     string
	FontStyle      string
}

func (c DrawTextCmd) Execute(p Painter) {
	p.DrawText(c.Text, c.X, c.Y, c.FontSize, c.Color, c.FontFamily, c.TextAlign, c.ContainerWidth, c.FontWeight, c.FontStyle)
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

// BeginScrollCmd activates scroll clipping and translation for a scrollable container.
type BeginScrollCmd struct {
	ClipRect layout.Rect
	OffsetX  float64
	OffsetY  float64
}

func (c BeginScrollCmd) Execute(p Painter) {
	p.BeginScroll(c.ClipRect, c.OffsetX, c.OffsetY)
}

// EndScrollCmd deactivates scroll clipping for a scrollable container.
type EndScrollCmd struct{}

func (c EndScrollCmd) Execute(p Painter) {
	p.EndScroll()
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

func (dl *DisplayList) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("DisplayList (%d commands):\n", len(dl.Commands)))

	for i, cmd := range dl.Commands {
		sb.WriteString(fmt.Sprintf("[%d] ", i))

		switch c := cmd.(type) {
		case FillRectCmd:
			sb.WriteString(fmt.Sprintf("FillRect: Rect(%v, %v, %v, %v) Color(R:%d, G:%d, B:%d, A:%d)\n",
				c.Rect.X, c.Rect.Y, c.Rect.Width, c.Rect.Height,
				c.Color.R, c.Color.G, c.Color.B, c.Color.A))
		case DrawBorderCmd:
			sb.WriteString(fmt.Sprintf("DrawBorder: Rect(%v, %v, %v, %v) Edges(L:%v, R:%v, T:%v, B:%v) Color(R:%d, G:%d, B:%d, A:%d)\n",
				c.Rect.X, c.Rect.Y, c.Rect.Width, c.Rect.Height,
				c.Edges.Left, c.Edges.Right, c.Edges.Top, c.Edges.Bottom,
				c.Color.R, c.Color.G, c.Color.B, c.Color.A))
		case DrawTextCmd:
			sb.WriteString(fmt.Sprintf("DrawText: Text(%q) Pos(%v, %v) FontSize(%v) Color(R:%d, G:%d, B:%d, A:%d) Font(%s, %s, %s) Align(%s) ContainerWidth(%v)\n",
				c.Text, c.X, c.Y, c.FontSize,
				c.Color.R, c.Color.G, c.Color.B, c.Color.A,
				c.FontFamily, c.FontWeight, c.FontStyle, c.TextAlign, c.ContainerWidth))
		case ClipRectCmd:
			if c.Clear {
				sb.WriteString("ClipRect: ClearClip\n")
			} else {
				sb.WriteString(fmt.Sprintf("ClipRect: Rect(%v, %v, %v, %v)\n",
					c.Rect.X, c.Rect.Y, c.Rect.Width, c.Rect.Height))
			}
		case BeginScrollCmd:
			sb.WriteString(fmt.Sprintf("BeginScroll: ClipRect(%v, %v, %v, %v) Offset(%v, %v)\n",
				c.ClipRect.X, c.ClipRect.Y, c.ClipRect.Width, c.ClipRect.Height,
				c.OffsetX, c.OffsetY))
		case EndScrollCmd:
			sb.WriteString("EndScroll\n")
		default:
			sb.WriteString(fmt.Sprintf("Unknown command: %T\n", cmd))
		}
	}

	return sb.String()
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
	// Clip if overflow:hidden is set, OR if the box has an explicit CSS width (non-auto).
	// Block boxes with a fixed width don't let content visually escape; this mirrors
	// how browsers enforce the box model even without explicit overflow:hidden.
	overflow := getStyleValue(box, "overflow")
	explicitWidth := getStyleValue(box, "width")
	isClipped := overflow == "hidden" || (overflow != "visible" && explicitWidth != "" && explicitWidth != "auto")
	if isClipped {
		clipRect := box.Dimensions.Content
		// Expand clip to include padding so background is not clipped
		clipRect.X -= box.Dimensions.Padding.Left
		clipRect.Y -= box.Dimensions.Padding.Top
		clipRect.Width += box.Dimensions.Padding.Left + box.Dimensions.Padding.Right
		clipRect.Height += box.Dimensions.Padding.Top + box.Dimensions.Padding.Bottom
		dl.Add(ClipRectCmd{Rect: clipRect})
	}

	// 3b. Scrollable containers: emit BeginScroll/EndScroll
	isScrollable := box.IsScrollable()
	if isScrollable {
		borderBox := box.Dimensions.BorderBox()
		dl.Add(BeginScrollCmd{
			ClipRect: borderBox,
			OffsetX:  box.ScrollOffsetX,
			OffsetY:  box.ScrollOffsetY,
		})
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

	// End scroll after children and text
	if isScrollable {
		dl.Add(EndScrollCmd{})
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
	return slices.Contains(namedColors, strings.ToLower(color))
}
