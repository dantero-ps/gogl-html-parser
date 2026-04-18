package render

import (
	"github.com/furkandgn/goglweb/internal/layout"
)

// blockContainerWidth walks up the parent chain to find the nearest
// block or anonymous box and returns its content width. This is used
// for text-align so the offset is relative to the containing block,
// not the text node's own measured width.
func blockContainerWidth(box *layout.LayoutBox) float64 {
	b := box.Parent
	for b != nil {
		if b.BoxType == layout.BlockBox || b.BoxType == layout.AnonymousBox || b.BoxType == layout.FlexBox {
			return b.Dimensions.Content.Width
		}
		b = b.Parent
	}
	return box.Dimensions.Content.Width
}

// RenderText handles the conversion of text-specific styles into DisplayList commands.
func RenderText(box *layout.LayoutBox, dl *DisplayList) {
	if box.StyledNode == nil || box.StyledNode.Node.Content == "" {
		return
	}

	text := box.StyledNode.Node.Content

	// Font Size - layout.go'daki parseLengthWithContext kullanarak tüm birimleri destekle
	fontSizeStr := getStyleValue(box, "font-size")
	ctx := &layout.LayoutContext{
		RootFontSize:   16.0,
		ViewportWidth:  800.0,
		ViewportHeight: 600.0,
	}
	fontSize := layout.ParseLengthWithContext(fontSizeStr, 16.0, box, ctx)

	// Text Color
	textColorStr := getStyleValue(box, "color")
	textColor := ParseColor(textColorStr)
	if textColorStr == "" {
		textColor = Color{0, 0, 0, 255} // Default black
	}

	// Font Family
	fontFamily := getStyleValue(box, "font-family")

	fontWeight := getStyleValue(box, "font-weight")
	fontStyle := getStyleValue(box, "font-style")

	ascent := box.TextAscent + 3
	if ascent <= 0 {
		ascent = fontSize
	}

	textAlign := getStyleValue(box, "text-align")
	containerWidth := blockContainerWidth(box)
	// If the layout engine re-wrapped this text node with a reduced width to make
	// room for inline siblings, use that width so the GPU wraps identically.
	if box.WrapWidth > 0 {
		containerWidth = box.WrapWidth
	}

	dl.Add(DrawTextCmd{
		Text:           text,
		X:              box.Dimensions.Content.X,
		Y:              box.Dimensions.Content.Y + ascent,
		FontSize:       fontSize,
		Color:          textColor,
		FontFamily:     fontFamily,
		TextAlign:      textAlign,
		ContainerWidth: containerWidth,
		FontWeight:     fontWeight,
		FontStyle:      fontStyle,
	})
}
