package render

import (
	"goglweb/internal/layout"
)

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
	if fontFamily == "" {
		fontFamily = "Helvatica"
	}

	// Add DrawText command using the Content box coordinates
	dl.Add(DrawTextCmd{
		Text:       text,
		X:          box.Dimensions.Content.X,
		Y:          box.Dimensions.Content.Y + fontSize, // Baseline adjustment
		FontSize:   fontSize,
		Color:      textColor,
		FontFamily: fontFamily,
	})
}
