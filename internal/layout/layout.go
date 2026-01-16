package layout

import (
	"regexp"
	"strconv"
	"strings"
)

var unitRegex = regexp.MustCompile(`^([+-]?\d*(?:\.\d+)?)(px|%|rem|em|vh|vw|vmin|vmax)$`)

func (box *LayoutBox) Layout(containingBlock Dimensions) {
	ctx := &LayoutContext{
		RootFontSize:   16.0,
		ViewportWidth:  800.0,
		ViewportHeight: 600.0,
	}
	box.LayoutWithContext(containingBlock, ctx)
}

func (box *LayoutBox) LayoutWithContext(containingBlock Dimensions, ctx *LayoutContext) {
	if box.BoxType == BlockBox {
		box.layoutBlock(containingBlock, ctx)
	} else if box.BoxType == InlineBox {
		box.layoutInline(containingBlock, ctx)
	} else if box.BoxType == AnonymousBox {
		box.layoutAnonymous(containingBlock, ctx)
	}
}

func (box *LayoutBox) layoutBlock(containingBlock Dimensions, ctx *LayoutContext) {
	box.calculateBlockDimensions(containingBlock, ctx)

	containingBlockForChildren := box.Dimensions
	containingBlockForChildren.Content.Height = 0
	containingBlockForChildren.Content.Y = box.Dimensions.Content.Y

	var previousMarginBottom float64
	for i, child := range box.Children {
		childMarginTop := ParseLengthWithContext(child.getMarginTop(), 0, box, ctx)
		if i > 0 {
			childMarginTop = collapseMargins(previousMarginBottom, childMarginTop)
		}

		containingBlockForChildren.Content.Y += childMarginTop
		child.LayoutWithContext(containingBlockForChildren, ctx)

		marginBox := child.Dimensions.MarginBox()
		containingBlockForChildren.Content.Y += marginBox.Height
		containingBlockForChildren.Content.Height += marginBox.Height
		previousMarginBottom = child.Dimensions.Margin.Bottom
	}

	if box.getHeight() == "auto" {
		box.Dimensions.Content.Height = containingBlockForChildren.Content.Y - box.Dimensions.Content.Y
	}
}

func collapseMargins(margin1, margin2 float64) float64 {
	if margin1 < 0 && margin2 < 0 {
		if margin1 < margin2 {
			return margin1
		}
		return margin2
	}
	if margin1 < 0 || margin2 < 0 {
		return margin1 + margin2
	}
	if margin1 > margin2 {
		return margin1
	}
	return margin2
}

func (box *LayoutBox) calculateBlockDimensions(containingBlock Dimensions, ctx *LayoutContext) {
	box.calculateBlockWidth(containingBlock, ctx)
	box.calculateBlockPosition(containingBlock, ctx)
	box.calculateBlockHeight(ctx)
}

func (box *LayoutBox) layoutInline(containingBlock Dimensions, ctx *LayoutContext) {
	// Set inline box position relative to containing block
	box.Dimensions.Content.X = containingBlock.Content.X
	box.Dimensions.Content.Y = containingBlock.Content.Y
	box.Dimensions.Content.Width = 0
	box.Dimensions.Content.Height = 0

	if box.StyledNode != nil && box.StyledNode.Node.Type == 0 {
		content := box.StyledNode.Node.Content
		fontSize := ParseLengthWithContext(box.getValue("font-size"), 16, box, ctx)
		box.Dimensions.Content.Width = float64(len(content)) * fontSize * 0.6
		box.Dimensions.Content.Height = fontSize
		// Set text node position
		box.Dimensions.Content.X = containingBlock.Content.X
		box.Dimensions.Content.Y = containingBlock.Content.Y
	}

	currentX := box.Dimensions.Content.X
	for _, child := range box.Children {
		child.LayoutWithContext(Dimensions{
			Content: Rect{
				X:      currentX,
				Y:      containingBlock.Content.Y,
				Width:  containingBlock.Content.Width,
				Height: containingBlock.Content.Height,
			},
		}, ctx)
		marginBox := child.Dimensions.MarginBox()
		currentX += marginBox.Width
		box.Dimensions.Content.Width += marginBox.Width
		if marginBox.Height > box.Dimensions.Content.Height {
			box.Dimensions.Content.Height = marginBox.Height
		}
	}
}

func (box *LayoutBox) layoutAnonymous(containingBlock Dimensions, ctx *LayoutContext) {
	box.Dimensions.Content.Width = containingBlock.Content.Width
	box.Dimensions.Content.Height = 0
	box.Dimensions.Content.X = containingBlock.Content.X
	box.Dimensions.Content.Y = containingBlock.Content.Y

	currentX := box.Dimensions.Content.X
	currentY := box.Dimensions.Content.Y
	maxHeight := 0.0

	for _, child := range box.Children {
		child.LayoutWithContext(Dimensions{
			Content: Rect{
				X:      currentX,
				Y:      currentY,
				Width:  containingBlock.Content.Width,
				Height: containingBlock.Content.Height,
			},
		}, ctx)

		marginBox := child.Dimensions.MarginBox()

		if marginBox.Height > maxHeight {
			maxHeight = marginBox.Height
		}

		// Wrap check: if line overflows, move to new line
		if currentX+marginBox.Width > containingBlock.Content.X+containingBlock.Content.Width {
			currentX = containingBlock.Content.X
			currentY += maxHeight
			maxHeight = marginBox.Height
			// Set position on new line
			child.Dimensions.Content.X = currentX
			child.Dimensions.Content.Y = currentY
		}

		currentX += marginBox.Width
	}

	box.Dimensions.Content.Height = maxHeight
}

func (box *LayoutBox) calculateBlockWidth(containingBlock Dimensions, ctx *LayoutContext) {
	width := box.getWidth()
	marginLeftStr := box.getMarginLeft()
	marginRightStr := box.getMarginRight()
	borderLeft := ParseLengthWithContext(box.getBorderLeft(), 0, box, ctx)
	borderRight := ParseLengthWithContext(box.getBorderRight(), 0, box, ctx)
	paddingLeft := ParseLengthWithContext(box.getPaddingLeft(), 0, box, ctx)
	paddingRight := ParseLengthWithContext(box.getPaddingRight(), 0, box, ctx)

	// Check for auto in margin-left and margin-right
	marginLeftAuto := marginLeftStr == "auto"
	marginRightAuto := marginRightStr == "auto"

	var marginLeft, marginRight float64
	if !marginLeftAuto {
		marginLeft = ParseLengthWithContext(marginLeftStr, 0, box, ctx)
	}
	if !marginRightAuto {
		marginRight = ParseLengthWithContext(marginRightStr, 0, box, ctx)
	}

	totalNonAutoWidth := marginLeft + marginRight + borderLeft + borderRight + paddingLeft + paddingRight

	var contentWidth float64
	if width != "" && width != "auto" {
		contentWidth = ParseLengthWithContext(width, containingBlock.Content.Width, box, ctx)
	} else {
		contentWidth = containingBlock.Content.Width - totalNonAutoWidth
	}

	minWidth := ParseLengthWithContext(box.getValue("min-width"), 0, box, ctx)
	maxWidthStr := box.getValue("max-width")
	var maxWidth float64
	if maxWidthStr != "" && maxWidthStr != "none" {
		maxWidth = ParseLengthWithContext(maxWidthStr, containingBlock.Content.Width, box, ctx)
		if maxWidth <= 0 {
			maxWidth = containingBlock.Content.Width
		}
	} else {
		maxWidth = containingBlock.Content.Width
	}

	if contentWidth < minWidth {
		contentWidth = minWidth
	}
	if maxWidth > 0 && contentWidth > maxWidth {
		contentWidth = maxWidth
	}
	if contentWidth < 0 {
		contentWidth = 0
	}

	// Auto margin calculation: if both are auto, center; if only one is auto, it becomes zero
	availableSpace := containingBlock.Content.Width - (contentWidth + borderLeft + borderRight + paddingLeft + paddingRight)
	if marginLeftAuto && marginRightAuto {
		// If both are auto, center
		marginLeft = availableSpace / 2.0
		marginRight = availableSpace / 2.0
	} else if marginLeftAuto {
		// If only left is auto, give all space to left
		marginLeft = availableSpace - marginRight
		if marginLeft < 0 {
			marginLeft = 0
		}
	} else if marginRightAuto {
		// If only right is auto, give all space to right
		marginRight = availableSpace - marginLeft
		if marginRight < 0 {
			marginRight = 0
		}
	}

	box.Dimensions.Content.Width = contentWidth

	box.Dimensions.Margin.Left = marginLeft
	box.Dimensions.Margin.Right = marginRight
	box.Dimensions.Border.Left = borderLeft
	box.Dimensions.Border.Right = borderRight
	box.Dimensions.Padding.Left = paddingLeft
	box.Dimensions.Padding.Right = paddingRight
}

func (box *LayoutBox) calculateBlockPosition(containingBlock Dimensions, ctx *LayoutContext) {
	marginTop := ParseLengthWithContext(box.getMarginTop(), 0, box, ctx)
	marginLeft := box.Dimensions.Margin.Left
	borderTop := ParseLengthWithContext(box.getBorderTop(), 0, box, ctx)
	borderLeft := box.Dimensions.Border.Left
	paddingTop := ParseLengthWithContext(box.getPaddingTop(), 0, box, ctx)
	paddingLeft := box.Dimensions.Padding.Left

	box.Dimensions.Margin.Top = marginTop
	box.Dimensions.Border.Top = borderTop
	box.Dimensions.Padding.Top = paddingTop

	box.Dimensions.Content.X = containingBlock.Content.X + marginLeft + borderLeft + paddingLeft
	box.Dimensions.Content.Y = containingBlock.Content.Y + marginTop + borderTop + paddingTop
}

func (box *LayoutBox) calculateBlockHeight(ctx *LayoutContext) {
	marginBottom := ParseLengthWithContext(box.getMarginBottom(), 0, box, ctx)
	borderBottom := ParseLengthWithContext(box.getBorderBottom(), 0, box, ctx)
	paddingBottom := ParseLengthWithContext(box.getPaddingBottom(), 0, box, ctx)

	box.Dimensions.Margin.Bottom = marginBottom
	box.Dimensions.Border.Bottom = borderBottom
	box.Dimensions.Padding.Bottom = paddingBottom

	height := box.getHeight()
	if height != "auto" {
		contentHeight := ParseLengthWithContext(height, 0, box, ctx)

		minHeight := ParseLengthWithContext(box.getValue("min-height"), 0, box, ctx)
		maxHeightStr := box.getValue("max-height")
		var maxHeight float64
		if maxHeightStr != "" && maxHeightStr != "none" {
			maxHeight = ParseLengthWithContext(maxHeightStr, 0, box, ctx)
		} else {
			maxHeight = 0
		}

		if contentHeight < minHeight {
			contentHeight = minHeight
		}
		if maxHeight > 0 && contentHeight > maxHeight {
			contentHeight = maxHeight
		}
		if contentHeight < 0 {
			contentHeight = 0
		}

		box.Dimensions.Content.Height = contentHeight
	}
}

func (box *LayoutBox) getValue(property string) string {
	if box.StyledNode == nil {
		return ""
	}
	value, ok := box.StyledNode.SpecifiedValues[property]
	if !ok {
		return ""
	}
	return value
}

func (box *LayoutBox) getWidth() string {
	return box.getValue("width")
}

func (box *LayoutBox) getHeight() string {
	height := box.getValue("height")
	if height == "" {
		return "auto"
	}
	return height
}

func (box *LayoutBox) getMarginLeft() string {
	if margin := box.getValue("margin-left"); margin != "" {
		return margin
	}
	if margin := box.getValue("margin"); margin != "" {
		return parseShorthand(margin, 0)
	}
	return ""
}

func (box *LayoutBox) getMarginRight() string {
	if margin := box.getValue("margin-right"); margin != "" {
		return margin
	}
	if margin := box.getValue("margin"); margin != "" {
		return parseShorthand(margin, 1)
	}
	return ""
}

func (box *LayoutBox) getMarginTop() string {
	if margin := box.getValue("margin-top"); margin != "" {
		return margin
	}
	if margin := box.getValue("margin"); margin != "" {
		return parseShorthand(margin, 2)
	}
	return ""
}

func (box *LayoutBox) getMarginBottom() string {
	if margin := box.getValue("margin-bottom"); margin != "" {
		return margin
	}
	if margin := box.getValue("margin"); margin != "" {
		return parseShorthand(margin, 3)
	}
	return ""
}

func (box *LayoutBox) getPaddingLeft() string {
	if padding := box.getValue("padding-left"); padding != "" {
		return padding
	}
	if padding := box.getValue("padding"); padding != "" {
		return parseShorthand(padding, 0)
	}
	return ""
}

func (box *LayoutBox) getPaddingRight() string {
	if padding := box.getValue("padding-right"); padding != "" {
		return padding
	}
	if padding := box.getValue("padding"); padding != "" {
		return parseShorthand(padding, 1)
	}
	return ""
}

func (box *LayoutBox) getPaddingTop() string {
	if padding := box.getValue("padding-top"); padding != "" {
		return padding
	}
	if padding := box.getValue("padding"); padding != "" {
		return parseShorthand(padding, 2)
	}
	return ""
}

func (box *LayoutBox) getPaddingBottom() string {
	if padding := box.getValue("padding-bottom"); padding != "" {
		return padding
	}
	if padding := box.getValue("padding"); padding != "" {
		return parseShorthand(padding, 3)
	}
	return ""
}

func (box *LayoutBox) getBorderLeft() string {
	if border := box.getValue("border-left-width"); border != "" {
		return border
	}
	if border := box.getValue("border"); border != "" {
		return extractBorderWidth(border)
	}
	return ""
}

func (box *LayoutBox) getBorderRight() string {
	if border := box.getValue("border-right-width"); border != "" {
		return border
	}
	if border := box.getValue("border"); border != "" {
		return extractBorderWidth(border)
	}
	return ""
}

func (box *LayoutBox) getBorderTop() string {
	if border := box.getValue("border-top-width"); border != "" {
		return border
	}
	if border := box.getValue("border"); border != "" {
		return extractBorderWidth(border)
	}
	return ""
}

func (box *LayoutBox) getBorderBottom() string {
	if border := box.getValue("border-bottom-width"); border != "" {
		return border
	}
	if border := box.getValue("border"); border != "" {
		return extractBorderWidth(border)
	}
	return ""
}

func extractBorderWidth(borderValue string) string {
	parts := strings.Fields(borderValue)
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func parseShorthand(value string, index int) string {
	parts := strings.Fields(value)
	if len(parts) == 0 {
		return ""
	}
	if len(parts) == 1 {
		return parts[0]
	}
	if len(parts) == 2 {
		// 2 değer: [top/bottom] [left/right]
		// index: 0=left, 1=right, 2=top, 3=bottom
		if index == 0 || index == 1 {
			// left veya right için ikinci değer
			return parts[1]
		}
		// top veya bottom için ilk değer
		return parts[0]
	}
	if len(parts) == 3 {
		if index == 0 {
			return parts[0]
		}
		if index == 1 || index == 3 {
			return parts[1]
		}
		return parts[2]
	}
	if len(parts) == 4 {
		return parts[index]
	}
	return parts[0]
}

func parseLength(value string, defaultValue float64) float64 {
	ctx := &LayoutContext{RootFontSize: 16.0, ViewportWidth: 800.0, ViewportHeight: 600.0}
	return ParseLengthWithContext(value, defaultValue, nil, ctx)
}

func ParseLengthWithContext(value string, defaultValue float64, box *LayoutBox, ctx *LayoutContext) float64 {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" || value == "auto" {
		return defaultValue
	}

	matches := unitRegex.FindStringSubmatch(value)
	if len(matches) != 3 {
		num, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return defaultValue
		}
		return num
	}

	numStr := matches[1]
	unit := matches[2]
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return defaultValue
	}

	switch unit {
	case "px":
		return num
	case "%":
		return defaultValue * (num / 100.0)
	case "rem":
		return num * ctx.RootFontSize
	case "em":
		fontSize := ctx.RootFontSize
		if box != nil {
			fontSizeValue := box.getValue("font-size")
			if fontSizeValue != "" {
				fontSize = ParseLengthWithContext(fontSizeValue, ctx.RootFontSize, getParentBox(box), ctx)
			}
		}
		return num * fontSize
	case "vh":
		return ctx.ViewportHeight * (num / 100.0)
	case "vw":
		return ctx.ViewportWidth * (num / 100.0)
	case "vmin":
		vmin := ctx.ViewportWidth
		if ctx.ViewportHeight < ctx.ViewportWidth {
			vmin = ctx.ViewportHeight
		}
		return vmin * (num / 100.0)
	case "vmax":
		vmax := ctx.ViewportHeight
		if ctx.ViewportWidth > ctx.ViewportHeight {
			vmax = ctx.ViewportWidth
		}
		return vmax * (num / 100.0)
	default:
		return num
	}
}

func getParentBox(box *LayoutBox) *LayoutBox {
	return nil
}
