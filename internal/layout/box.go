package layout

import "goglweb/internal/style"

type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

type LayoutContext struct {
	RootFontSize   float64
	ViewportWidth  float64
	ViewportHeight float64
}

func NewLayoutContext(viewportWidth, viewportHeight float64) *LayoutContext {
	return &LayoutContext{
		RootFontSize:   16.0,
		ViewportWidth:  viewportWidth,
		ViewportHeight: viewportHeight,
	}
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
)

type LayoutBox struct {
	BoxType    BoxType
	StyledNode *style.StyledNode
	Dimensions Dimensions
	Children   []*LayoutBox
}
