package layout

import "github.com/furkandgn/goglweb/internal/style"

type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

type LayoutContext struct {
	RootFontSize        float64
	ViewportWidth       float64
	ViewportHeight      float64
	TextMeasurer        TextMeasurer
	PositionedAncestors []*LayoutBox // stack; top is nearest positioned ancestor
}

func NewLayoutContext(viewportWidth, viewportHeight float64) *LayoutContext {
	return &LayoutContext{
		RootFontSize:   16.0,
		ViewportWidth:  viewportWidth,
		ViewportHeight: viewportHeight,
	}
}

// GetTextMeasurer returns the configured TextMeasurer, falling back to FallbackMeasurer.
func (ctx *LayoutContext) GetTextMeasurer() TextMeasurer {
	if ctx.TextMeasurer != nil {
		return ctx.TextMeasurer
	}
	return &FallbackMeasurer{}
}

type EdgeSizes struct {
	Left   float64
	Right  float64
	Top    float64
	Bottom float64
}

type Dimensions struct {
	Content Rect
	Padding EdgeSizes
	Border  EdgeSizes
	Margin  EdgeSizes
}

func (d Dimensions) PaddingBox() Rect {
	return Rect{
		X:      d.Content.X - d.Padding.Left,
		Y:      d.Content.Y - d.Padding.Top,
		Width:  d.Content.Width + d.Padding.Left + d.Padding.Right,
		Height: d.Content.Height + d.Padding.Top + d.Padding.Bottom,
	}
}

func (d Dimensions) BorderBox() Rect {
	paddingBox := d.PaddingBox()
	return Rect{
		X:      paddingBox.X - d.Border.Left,
		Y:      paddingBox.Y - d.Border.Top,
		Width:  paddingBox.Width + d.Border.Left + d.Border.Right,
		Height: paddingBox.Height + d.Border.Top + d.Border.Bottom,
	}
}

func (d Dimensions) MarginBox() Rect {
	borderBox := d.BorderBox()
	return Rect{
		X:      borderBox.X - d.Margin.Left,
		Y:      borderBox.Y - d.Margin.Top,
		Width:  borderBox.Width + d.Margin.Left + d.Margin.Right,
		Height: borderBox.Height + d.Margin.Top + d.Margin.Bottom,
	}
}

type BoxType int

const (
	BlockBox BoxType = iota
	InlineBox
	AnonymousBox
	PositionedBox // position: absolute or fixed
	FlexBox       // display: flex container
)

type LayoutBox struct {
	BoxType    BoxType
	StyledNode *style.StyledNode
	Dimensions Dimensions
	Children   []*LayoutBox
	Parent     *LayoutBox
	TextAscent float64 // distance from content top to baseline, set for text nodes

	// Inline flow fields — set by layoutInline for text nodes that wrap.
	// LastLineWidth is the width of the final wrapped line (used by anonymous
	// box flow so the next sibling continues on the same line).
	// NumLines is how many wrapped lines the text produces.
	LastLineWidth float64
	NumLines      int

	// Positioned layout fields
	PositionType     string       // "static" | "relative" | "absolute" | "fixed"
	DeferredAbsolute []*LayoutBox // absolute/fixed children deferred to second pass

	// WrapWidth is the effective word-wrap width for this text node when it has
	// been re-wrapped to make room for subsequent inline siblings. When non-zero,
	// the GPU renderer uses this instead of the full block-container width.
	WrapWidth float64

	// Scroll fields
	ScrollOffsetX  float64 // pixels scrolled horizontally (positive = scrolled right)
	ScrollOffsetY  float64 // pixels scrolled vertically (positive = scrolled down)
	Overflow       string  // "" | "visible" | "hidden" | "scroll" | "auto"
	ChildrenHeight float64 // total height occupied by children (used for scroll bounds)
}

func (box *LayoutBox) IsScrollable() bool {
	return box.Overflow == "scroll" || box.Overflow == "auto"
}

// ClampScrollOffset clamps a scroll offset to valid bounds.
func ClampScrollOffset(offset, contentHeight, visibleHeight float64) float64 {
	if offset < 0 {
		return 0
	}
	max := contentHeight - visibleHeight
	if max < 0 {
		max = 0
	}
	if offset > max {
		return max
	}
	return offset
}

// FlexConfig holds resolved flex container properties.
type FlexConfig struct {
	Direction      string // "row" | "column"
	Wrap           string // "nowrap" | "wrap" | "wrap-reverse"
	JustifyContent string // "flex-start" | "flex-end" | "center" | "space-between" | "space-around"
	AlignItems     string // "stretch" | "flex-start" | "flex-end" | "center"
}
