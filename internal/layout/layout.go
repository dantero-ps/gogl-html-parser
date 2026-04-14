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
	} else if box.BoxType == PositionedBox {
		// Absolute and fixed boxes are laid out by their containing block in Pass 2.
		// Nothing to do here during normal flow.
		return
	} else if box.BoxType == FlexBox {
		box.layoutFlex(containingBlock, ctx)
	}
}

func (box *LayoutBox) layoutBlock(containingBlock Dimensions, ctx *LayoutContext) {
	box.calculateBlockDimensions(containingBlock, ctx)

	// Push this box onto the positioned ancestor stack if it is a positioned element
	isPositioned := box.PositionType == "relative" || box.PositionType == "absolute" || box.PositionType == "fixed"
	if isPositioned {
		ctx.PositionedAncestors = append(ctx.PositionedAncestors, box)
	}

	containingBlockForChildren := box.Dimensions
	containingBlockForChildren.Content.Height = 0
	containingBlockForChildren.Content.Y = box.Dimensions.Content.Y

	var previousMarginBottom float64
	for i, child := range box.Children {
		// Defer absolute/fixed children to Pass 2
		if child.PositionType == "absolute" || child.PositionType == "fixed" {
			nearest := nearestPositionedAncestor(ctx)
			if nearest != nil {
				nearest.DeferredAbsolute = append(nearest.DeferredAbsolute, child)
			}
			continue
		}

		if i > 0 {
			childMarginTop := ParseLengthWithContext(child.getMarginTop(), 0, box, ctx)
			collapsedMargin := collapseMargins(previousMarginBottom, childMarginTop)
			overlap := previousMarginBottom + childMarginTop - collapsedMargin
			containingBlockForChildren.Content.Y -= overlap
		}

		child.LayoutWithContext(containingBlockForChildren, ctx)

		marginBox := child.Dimensions.MarginBox()
		containingBlockForChildren.Content.Y += marginBox.Height
		containingBlockForChildren.Content.Height += marginBox.Height
		previousMarginBottom = child.Dimensions.Margin.Bottom
	}

	totalChildrenHeight := containingBlockForChildren.Content.Y - box.Dimensions.Content.Y
	box.ChildrenHeight = totalChildrenHeight
	if box.getHeight() == "auto" {
		box.Dimensions.Content.Height = totalChildrenHeight
	}

	// Pop from positioned ancestor stack and layout deferred absolute/fixed children
	if isPositioned {
		if len(ctx.PositionedAncestors) > 0 {
			ctx.PositionedAncestors = ctx.PositionedAncestors[:len(ctx.PositionedAncestors)-1]
		}
		box.layoutDeferredAbsolute(ctx)
	}

	// Apply relative offset after layout is complete
	if box.PositionType == "relative" {
		box.applyRelativeOffset(ctx)
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
	box.calculateBlockHeight(containingBlock, ctx)
}

func (box *LayoutBox) layoutInline(containingBlock Dimensions, ctx *LayoutContext) {
	box.Dimensions.Content.X = containingBlock.Content.X
	box.Dimensions.Content.Y = containingBlock.Content.Y
	box.Dimensions.Content.Width = 0
	box.Dimensions.Content.Height = 0

	if box.StyledNode != nil && box.StyledNode.Node.Type == 0 {
		content := box.StyledNode.Node.Content
		fontSize := ParseLengthWithContext(box.getValue("font-size"), 16, box, ctx)
		fontFamily := box.getValue("font-family")
		fontWeight := box.getValue("font-weight")
		fontStyle := box.getValue("font-style")
		measurer := ctx.GetTextMeasurer()

		maxWidth := containingBlock.Content.Width
		var totalHeight float64
		var maxLineWidth float64
		var lastLineWidth float64
		numLines := 1

		if maxWidth > 0 {
			lines := WordWrap(content, fontFamily, fontSize, maxWidth, measurer, fontWeight, fontStyle)
			numLines = len(lines)
			for _, line := range lines {
				m := measurer.MeasureText(line, fontFamily, fontSize, fontWeight, fontStyle)
				totalHeight += m.LineHeight
				if m.Width > maxLineWidth {
					maxLineWidth = m.Width
				}
				lastLineWidth = m.Width
			}
			if numLines == 0 {
				totalHeight = fontSize
				numLines = 1
			}
		} else {
			metrics := measurer.MeasureText(content, fontFamily, fontSize, fontWeight, fontStyle)
			maxLineWidth = metrics.Width
			lastLineWidth = metrics.Width
			totalHeight = metrics.Height
		}

		if maxLineWidth > maxWidth && maxWidth > 0 {
			maxLineWidth = maxWidth
		}
		box.Dimensions.Content.Width = maxLineWidth
		box.Dimensions.Content.Height = totalHeight
		box.LastLineWidth = lastLineWidth
		box.NumLines = numLines

		firstMetrics := measurer.MeasureText(content, fontFamily, fontSize, fontWeight, fontStyle)
		box.TextAscent = firstMetrics.Ascent

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

	// Propagate inline flow info from children (for span elements wrapping text).
	if box.NumLines == 0 {
		box.NumLines = 1
		box.LastLineWidth = box.Dimensions.Content.Width
		// If we have children, inherit from last child
		if len(box.Children) > 0 {
			last := box.Children[len(box.Children)-1]
			if last.NumLines > 1 {
				box.NumLines = last.NumLines
				box.LastLineWidth = last.LastLineWidth
			}
		}
	}
}

func (box *LayoutBox) layoutAnonymous(containingBlock Dimensions, ctx *LayoutContext) {
	box.Dimensions.Content.Width = containingBlock.Content.Width
	box.Dimensions.Content.Height = 0
	box.Dimensions.Content.X = containingBlock.Content.X
	box.Dimensions.Content.Y = containingBlock.Content.Y

	left := box.Dimensions.Content.X
	right := left + containingBlock.Content.Width
	currentX := left
	currentY := box.Dimensions.Content.Y
	lineHeight := 0.0

	// Track the most recent text-node child that started a line at `left`,
	// so we can re-wrap it when a subsequent sibling overflows the line.
	var lastTextChild *LayoutBox
	lastTextChildLineY := currentY

	for _, child := range box.Children {
		// Layout child with currentX so it positions itself correctly.
		// Full container width is passed so the child can word-wrap normally.
		child.LayoutWithContext(Dimensions{
			Content: Rect{
				X:      currentX,
				Y:      currentY,
				Width:  containingBlock.Content.Width,
				Height: containingBlock.Content.Height,
			},
		}, ctx)

		childLineH := child.Dimensions.Content.Height
		if child.NumLines > 1 {
			childLineH = child.Dimensions.Content.Height / float64(child.NumLines)
		}
		if childLineH > lineHeight {
			lineHeight = childLineH
		}

		// How this child affects the cursor depends on whether it wraps.
		if child.NumLines > 1 {
			// Multi-line text: move cursor to start of last line.
			wrappedLineH := child.Dimensions.Content.Height / float64(child.NumLines)
			advanceY := wrappedLineH * float64(child.NumLines-1)
			currentY += advanceY
			currentX = left + child.LastLineWidth
			lineHeight = wrappedLineH
			// A multi-line text child now acts as the "line start" text node.
			if isTextNode(child) {
				lastTextChild = child
				lastTextChildLineY = currentY - advanceY // start of this whole block
			} else {
				lastTextChild = nil
			}
		} else {
			// Single-line child: advance X by its width.
			childWidth := child.Dimensions.Content.Width
			if child.LastLineWidth > 0 {
				childWidth = child.LastLineWidth
			}

			// Overflow check.
			if currentX+childWidth > right && currentX > left {
				// Before wrapping to a new line, try re-wrapping the most recent
				// text-node sibling (lastTextChild) with a reduced width so that
				// this child can follow on the same line.
				rewrapped := false
				if lastTextChild != nil && isTextNode(lastTextChild) {
					spaceNeeded := childWidth
					maxWidthForPrev := right - left - spaceNeeded
					// Only attempt if the reduced width would actually cause text to wrap.
					if maxWidthForPrev > 0 && lastTextChild.Dimensions.Content.Width > maxWidthForPrev {
						lastTextChild.WrapWidth = maxWidthForPrev
						lastTextChild.LayoutWithContext(Dimensions{
							Content: Rect{
								X:      left,
								Y:      lastTextChildLineY,
								Width:  maxWidthForPrev,
								Height: containingBlock.Content.Height,
							},
						}, ctx)
						if lastTextChild.NumLines > 1 {
							lh := lastTextChild.Dimensions.Content.Height / float64(lastTextChild.NumLines)
							currentY = lastTextChildLineY + lh*float64(lastTextChild.NumLines-1)
							currentX = left + lastTextChild.LastLineWidth
							lineHeight = lh
						} else {
							currentX = left + lastTextChild.LastLineWidth
							currentY = lastTextChildLineY
						}
						// Re-layout the current child at the updated position.
						child.LayoutWithContext(Dimensions{
							Content: Rect{
								X:      currentX,
								Y:      currentY,
								Width:  containingBlock.Content.Width,
								Height: containingBlock.Content.Height,
							},
						}, ctx)
						childLineH = child.Dimensions.Content.Height
						if child.NumLines > 1 {
							childLineH = child.Dimensions.Content.Height / float64(child.NumLines)
						}
						if childLineH > lineHeight {
							lineHeight = childLineH
						}
						childWidth = child.Dimensions.Content.Width
						if child.LastLineWidth > 0 {
							childWidth = child.LastLineWidth
						}
						if child.NumLines > 1 {
							wrappedLineH := child.Dimensions.Content.Height / float64(child.NumLines)
							advanceY := wrappedLineH * float64(child.NumLines-1)
							currentY += advanceY
							currentX = left + child.LastLineWidth
							lineHeight = wrappedLineH
						} else {
							currentX += childWidth
						}
						lastTextChild = nil
						rewrapped = true
					}
				}

				if !rewrapped {
					// Regular line wrap.
					currentY += lineHeight
					currentX = left
					lineHeight = 0
					lastTextChild = nil
					lastTextChildLineY = currentY

					child.LayoutWithContext(Dimensions{
						Content: Rect{
							X:      currentX,
							Y:      currentY,
							Width:  containingBlock.Content.Width,
							Height: containingBlock.Content.Height,
						},
					}, ctx)

					childLineH = child.Dimensions.Content.Height
					if child.NumLines > 1 {
						childLineH = child.Dimensions.Content.Height / float64(child.NumLines)
					}
					if childLineH > lineHeight {
						lineHeight = childLineH
					}
					if child.NumLines > 1 {
						wrappedLineH := child.Dimensions.Content.Height / float64(child.NumLines)
						advanceY := wrappedLineH * float64(child.NumLines-1)
						currentY += advanceY
						currentX = left + child.LastLineWidth
						lineHeight = wrappedLineH
					} else {
						childWidth = child.Dimensions.Content.Width
						if child.LastLineWidth > 0 {
							childWidth = child.LastLineWidth
						}
						currentX += childWidth
					}
				}
			} else {
				currentX += childWidth
				// Track text children that start at the left edge of a line.
				if isTextNode(child) && currentX-childWidth == left {
					lastTextChild = child
					lastTextChildLineY = currentY
				}
			}
		}
	}

	totalHeight := (currentY - box.Dimensions.Content.Y) + lineHeight
	box.Dimensions.Content.Height = totalHeight
}

// isTextNode reports whether box is a pure text node (no tag, just content).
func isTextNode(box *LayoutBox) bool {
	return box != nil && box.StyledNode != nil && box.StyledNode.Node.Type == 0
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

func (box *LayoutBox) calculateBlockHeight(containingBlock Dimensions, ctx *LayoutContext) {
	marginBottom := ParseLengthWithContext(box.getMarginBottom(), 0, box, ctx)
	borderBottom := ParseLengthWithContext(box.getBorderBottom(), 0, box, ctx)
	paddingBottom := ParseLengthWithContext(box.getPaddingBottom(), 0, box, ctx)

	box.Dimensions.Margin.Bottom = marginBottom
	box.Dimensions.Border.Bottom = borderBottom
	box.Dimensions.Padding.Bottom = paddingBottom

	height := box.getHeight()
	if height != "auto" {
		contentHeight := ParseLengthWithContext(height, containingBlock.Content.Height, box, ctx)

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
		// CSS: [top] [left/right] [bottom]
		if index == 0 {
			return parts[1]
		}
		if index == 1 {
			return parts[1]
		}
		if index == 2 {
			return parts[0]
		}
		return parts[2]
	}
	if len(parts) == 4 {
		// CSS: [top] [right] [bottom] [left]
		// index: 0=left, 1=right, 2=top, 3=bottom
		return []string{parts[3], parts[1], parts[0], parts[2]}[index]
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
	if box == nil {
		return nil
	}
	return box.Parent
}

// nearestPositionedAncestor returns the top of the PositionedAncestors stack, or nil.
func nearestPositionedAncestor(ctx *LayoutContext) *LayoutBox {
	if len(ctx.PositionedAncestors) == 0 {
		return nil
	}
	return ctx.PositionedAncestors[len(ctx.PositionedAncestors)-1]
}

// viewportDimensions returns a Dimensions representing the viewport.
func viewportDimensions(ctx *LayoutContext) Dimensions {
	return Dimensions{Content: Rect{X: 0, Y: 0, Width: ctx.ViewportWidth, Height: ctx.ViewportHeight}}
}

// layoutDeferredAbsolute lays out all absolute/fixed children deferred to this box.
func (box *LayoutBox) layoutDeferredAbsolute(ctx *LayoutContext) {
	for _, child := range box.DeferredAbsolute {
		cb := box.Dimensions // containing block is this box
		if child.PositionType == "fixed" {
			cb = viewportDimensions(ctx)
		}
		child.layoutAbsolute(cb, ctx)
	}
}

// layoutAbsolute sizes and positions an absolute or fixed box within its containing block.
func (box *LayoutBox) layoutAbsolute(containingBlock Dimensions, ctx *LayoutContext) {
	// Calculate width
	w := box.getWidth()
	if w != "" && w != "auto" {
		box.Dimensions.Content.Width = ParseLengthWithContext(w, containingBlock.Content.Width, box, ctx)
	} else {
		box.Dimensions.Content.Width = 0 // shrink-to-content placeholder
	}

	// Calculate margins, borders, padding
	box.calculateBlockWidth(containingBlock, ctx)
	box.calculateBlockPosition(containingBlock, ctx)

	// Set containing block for children
	containingBlockForChildren := box.Dimensions
	containingBlockForChildren.Content.Height = 0
	containingBlockForChildren.Content.Y = box.Dimensions.Content.Y

	// Push this box onto positioned ancestor stack for nested absolute children
	if box.PositionType == "absolute" || box.PositionType == "fixed" {
		ctx.PositionedAncestors = append(ctx.PositionedAncestors, box)
	}

	// Lay out children to determine height
	for _, child := range box.Children {
		if child.PositionType == "absolute" || child.PositionType == "fixed" {
			nearest := nearestPositionedAncestor(ctx)
			if nearest != nil {
				nearest.DeferredAbsolute = append(nearest.DeferredAbsolute, child)
			}
			continue
		}
		child.LayoutWithContext(containingBlockForChildren, ctx)
		marginBox := child.Dimensions.MarginBox()
		containingBlockForChildren.Content.Y += marginBox.Height
		containingBlockForChildren.Content.Height += marginBox.Height
	}

	// Pop from positioned ancestor stack
	if box.PositionType == "absolute" || box.PositionType == "fixed" {
		if len(ctx.PositionedAncestors) > 0 {
			ctx.PositionedAncestors = ctx.PositionedAncestors[:len(ctx.PositionedAncestors)-1]
		}
		box.layoutDeferredAbsolute(ctx)
	}

	// Calculate height
	height := box.getHeight()
	if height == "auto" {
		box.Dimensions.Content.Height = containingBlockForChildren.Content.Height
	} else {
		box.Dimensions.Content.Height = ParseLengthWithContext(height, containingBlock.Content.Height, box, ctx)
	}

	// Calculate border/padding bottom
	borderBottom := ParseLengthWithContext(box.getBorderBottom(), 0, box, ctx)
	paddingBottom := ParseLengthWithContext(box.getPaddingBottom(), 0, box, ctx)
	marginBottom := ParseLengthWithContext(box.getMarginBottom(), 0, box, ctx)
	box.Dimensions.Border.Bottom = borderBottom
	box.Dimensions.Padding.Bottom = paddingBottom
	box.Dimensions.Margin.Bottom = marginBottom

	// Resolve top/left/bottom/right to position within containing block
	top := box.getValue("top")
	left := box.getValue("left")
	bottom := box.getValue("bottom")
	right := box.getValue("right")

	cbLeft := containingBlock.Content.X
	cbTop := containingBlock.Content.Y
	cbW := containingBlock.Content.Width
	cbH := containingBlock.Content.Height

	if left != "" && left != "auto" {
		box.Dimensions.Content.X = cbLeft + box.Dimensions.Margin.Left + box.Dimensions.Border.Left + box.Dimensions.Padding.Left + ParseLengthWithContext(left, cbW, box, ctx)
	} else if right != "" && right != "auto" {
		box.Dimensions.Content.X = cbLeft + cbW - box.Dimensions.Margin.Right - box.Dimensions.Border.Right - box.Dimensions.Padding.Right - box.Dimensions.Content.Width - ParseLengthWithContext(right, cbW, box, ctx)
	}
	if top != "" && top != "auto" {
		box.Dimensions.Content.Y = cbTop + box.Dimensions.Margin.Top + box.Dimensions.Border.Top + box.Dimensions.Padding.Top + ParseLengthWithContext(top, cbH, box, ctx)
	} else if bottom != "" && bottom != "auto" {
		box.Dimensions.Content.Y = cbTop + cbH - box.Dimensions.Margin.Bottom - box.Dimensions.Border.Bottom - box.Dimensions.Padding.Bottom - box.Dimensions.Content.Height - ParseLengthWithContext(bottom, cbH, box, ctx)
	}
}

// applyRelativeOffset shifts a relatively positioned box by its top/left/bottom/right offsets.
func (box *LayoutBox) applyRelativeOffset(ctx *LayoutContext) {
	top := box.getValue("top")
	left := box.getValue("left")
	bottom := box.getValue("bottom")
	right := box.getValue("right")

	if top != "" && top != "auto" {
		box.Dimensions.Content.Y += ParseLengthWithContext(top, 0, box, ctx)
	} else if bottom != "" && bottom != "auto" {
		box.Dimensions.Content.Y -= ParseLengthWithContext(bottom, 0, box, ctx)
	}
	if left != "" && left != "auto" {
		box.Dimensions.Content.X += ParseLengthWithContext(left, 0, box, ctx)
	} else if right != "" && right != "auto" {
		box.Dimensions.Content.X -= ParseLengthWithContext(right, 0, box, ctx)
	}
}

// --- Flex Layout ---

// resolveFlexConfig reads flex container properties from the box's styled node.
func (box *LayoutBox) resolveFlexConfig() FlexConfig {
	dir := box.getValue("flex-direction")
	if dir == "" {
		dir = "row"
	}
	wrap := box.getValue("flex-wrap")
	if wrap == "" {
		wrap = "nowrap"
	}
	jc := box.getValue("justify-content")
	if jc == "" {
		jc = "flex-start"
	}
	ai := box.getValue("align-items")
	if ai == "" {
		ai = "stretch"
	}
	return FlexConfig{Direction: dir, Wrap: wrap, JustifyContent: jc, AlignItems: ai}
}

// resolveFlexItemBasis computes the base size of a flex item along the main axis.
func resolveFlexItemBasis(item *LayoutBox, mainAxisSize float64, ctx *LayoutContext) float64 {
	basis := item.getValue("flex-basis")
	if basis == "" || basis == "auto" {
		w := item.getValue("width")
		if w != "" && w != "auto" {
			return ParseLengthWithContext(w, mainAxisSize, item, ctx)
		}
		return 0 // shrink-to-content; filled after child layout
	}
	return ParseLengthWithContext(basis, mainAxisSize, item, ctx)
}

// layoutFlex implements the CSS Flexbox layout algorithm.
func (box *LayoutBox) layoutFlex(containingBlock Dimensions, ctx *LayoutContext) {
	// --- Container sizing ---
	box.calculateBlockWidth(containingBlock, ctx)
	box.calculateBlockPosition(containingBlock, ctx)
	box.calculateBlockHeight(containingBlock, ctx)

	cfg := box.resolveFlexConfig()
	isRow := cfg.Direction == "row" || cfg.Direction == "row-reverse"

	mainSize := box.Dimensions.Content.Width
	if !isRow {
		mainSize = box.Dimensions.Content.Height
	}

	// --- Step 1: resolve base sizes and gather items ---
	type flexItem struct {
		box       *LayoutBox
		baseSize  float64
		grow      float64
		shrink    float64
		crossSize float64
	}
	items := make([]flexItem, 0, len(box.Children))
	for _, child := range box.Children {
		if child == nil {
			continue
		}
		// Skip positioned children in flex layout
		if child.PositionType == "absolute" || child.PositionType == "fixed" {
			nearest := nearestPositionedAncestor(ctx)
			if nearest != nil {
				nearest.DeferredAbsolute = append(nearest.DeferredAbsolute, child)
			}
			continue
		}
		grow := ParseLengthWithContext(child.getValue("flex-grow"), 0, child, ctx)
		shrink := ParseLengthWithContext(child.getValue("flex-shrink"), 1, child, ctx)
		base := resolveFlexItemBasis(child, mainSize, ctx)
		items = append(items, flexItem{box: child, baseSize: base, grow: grow, shrink: shrink})
	}

	// --- Step 2: lay out each item to get intrinsic cross size ---
	for i := range items {
		cb := box.Dimensions
		if isRow {
			if items[i].baseSize > 0 {
				cb.Content.Width = items[i].baseSize
			}
		} else {
			if items[i].baseSize > 0 {
				cb.Content.Height = items[i].baseSize
			}
		}
		items[i].box.LayoutWithContext(cb, ctx)
		if isRow {
			if items[i].baseSize == 0 {
				items[i].baseSize = items[i].box.Dimensions.MarginBox().Width
			}
			items[i].crossSize = items[i].box.Dimensions.MarginBox().Height
		} else {
			if items[i].baseSize == 0 {
				items[i].baseSize = items[i].box.Dimensions.MarginBox().Height
			}
			items[i].crossSize = items[i].box.Dimensions.MarginBox().Width
		}
	}

	// --- Step 3: wrap lines ---
	type flexLine struct {
		items     []int
		mainUsed  float64
		crossSize float64
	}
	var lines []flexLine
	if cfg.Wrap == "nowrap" {
		line := flexLine{}
		for i := range items {
			line.items = append(line.items, i)
			line.mainUsed += items[i].baseSize
		}
		lines = []flexLine{line}
	} else {
		current := flexLine{}
		for i := range items {
			if current.mainUsed+items[i].baseSize > mainSize && len(current.items) > 0 {
				lines = append(lines, current)
				current = flexLine{}
			}
			current.items = append(current.items, i)
			current.mainUsed += items[i].baseSize
		}
		if len(current.items) > 0 {
			lines = append(lines, current)
		}
	}

	// --- Step 4: grow/shrink within each line ---
	for li := range lines {
		free := mainSize - lines[li].mainUsed
		if free > 0 {
			totalGrow := 0.0
			for _, idx := range lines[li].items {
				totalGrow += items[idx].grow
			}
			if totalGrow > 0 {
				for _, idx := range lines[li].items {
					items[idx].baseSize += free * (items[idx].grow / totalGrow)
				}
			}
		} else if free < 0 {
			totalShrink := 0.0
			for _, idx := range lines[li].items {
				totalShrink += items[idx].shrink
			}
			if totalShrink > 0 {
				for _, idx := range lines[li].items {
					items[idx].baseSize += free * (items[idx].shrink / totalShrink)
					if items[idx].baseSize < 0 {
						items[idx].baseSize = 0
					}
				}
			}
		}
		// compute line cross size
		for _, idx := range lines[li].items {
			if items[idx].crossSize > lines[li].crossSize {
				lines[li].crossSize = items[idx].crossSize
			}
		}
	}

	// --- Step 5: position items ---
	crossOrigin := box.Dimensions.Content.Y
	if !isRow {
		crossOrigin = box.Dimensions.Content.X
	}

	for _, line := range lines {
		// justify-content offsets
		usedMain := 0.0
		for _, idx := range line.items {
			usedMain += items[idx].baseSize
		}
		freeMain := mainSize - usedMain
		var startOffset, gap float64
		switch cfg.JustifyContent {
		case "flex-end":
			startOffset = freeMain
		case "center":
			startOffset = freeMain / 2
		case "space-between":
			if len(line.items) > 1 {
				gap = freeMain / float64(len(line.items)-1)
			}
		case "space-around":
			if len(line.items) > 0 {
				gap = freeMain / float64(len(line.items))
				startOffset = gap / 2
			}
		}

		mainCursor := startOffset
		if isRow {
			mainCursor += box.Dimensions.Content.X
		} else {
			mainCursor += box.Dimensions.Content.Y
		}

		for _, idx := range line.items {
			item := items[idx]
			// Re-layout with final main-axis size to get correct cross size.
			finalCB := box.Dimensions
			if isRow {
				finalCB.Content.Width = item.baseSize
				finalCB.Content.X = mainCursor
				finalCB.Content.Y = crossOrigin
			} else {
				finalCB.Content.Height = item.baseSize
				finalCB.Content.Y = mainCursor
				finalCB.Content.X = crossOrigin
			}
			item.box.LayoutWithContext(finalCB, ctx)

			// align-items cross placement
			if isRow {
				switch cfg.AlignItems {
				case "center":
					item.box.Dimensions.Content.Y = crossOrigin + (line.crossSize-item.box.Dimensions.MarginBox().Height)/2
				case "flex-end":
					item.box.Dimensions.Content.Y = crossOrigin + line.crossSize - item.box.Dimensions.MarginBox().Height
				case "stretch":
					stretchH := line.crossSize - item.box.Dimensions.Margin.Top - item.box.Dimensions.Margin.Bottom - item.box.Dimensions.Padding.Top - item.box.Dimensions.Padding.Bottom - item.box.Dimensions.Border.Top - item.box.Dimensions.Border.Bottom
					if stretchH > 0 {
						item.box.Dimensions.Content.Height = stretchH
					}
				}
				mainCursor += item.box.Dimensions.MarginBox().Width + gap
			} else {
				switch cfg.AlignItems {
				case "center":
					item.box.Dimensions.Content.X = crossOrigin + (line.crossSize-item.box.Dimensions.MarginBox().Width)/2
				case "flex-end":
					item.box.Dimensions.Content.X = crossOrigin + line.crossSize - item.box.Dimensions.MarginBox().Width
				case "stretch":
					stretchW := line.crossSize - item.box.Dimensions.Margin.Left - item.box.Dimensions.Margin.Right - item.box.Dimensions.Padding.Left - item.box.Dimensions.Padding.Right - item.box.Dimensions.Border.Left - item.box.Dimensions.Border.Right
					if stretchW > 0 {
						item.box.Dimensions.Content.Width = stretchW
					}
				}
				mainCursor += item.box.Dimensions.MarginBox().Height + gap
			}
		}

		if isRow {
			crossOrigin += line.crossSize
		} else {
			crossOrigin += line.crossSize
		}
	}

	// --- Step 6: set container height (if auto) ---
	if box.getHeight() == "auto" {
		if isRow {
			totalCross := 0.0
			for _, line := range lines {
				totalCross += line.crossSize
			}
			box.Dimensions.Content.Height = totalCross
		} else {
			totalMain := 0.0
			for _, line := range lines {
				for _, idx := range line.items {
					totalMain += items[idx].baseSize
				}
			}
			box.Dimensions.Content.Height = totalMain
		}
	}
}
