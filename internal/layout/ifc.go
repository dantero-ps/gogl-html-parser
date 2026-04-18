package layout

import (
	"strings"
	"unicode"
)

// ---- Inline item stream ----

// InlineItemKind classifies an item in the IFC stream.
type InlineItemKind int8

const (
	IFCTextWord  InlineItemKind = iota // a single word from a text node
	IFCSpanStart                       // start of an inline box (span)
	IFCSpanEnd                         // end of an inline box (span)
	IFCSoftSpace                       // whitespace-only text node between elements
)

// InlineItem is one atom in the flattened IFC stream.
// TextBox and SpanBox are mutually exclusive depending on Kind.
type InlineItem struct {
	Kind InlineItemKind

	// IFCTextWord fields:
	Word       string
	HasSpace   bool // a breakable space follows this word
	HasLeading bool // a space precedes (text node starts with whitespace)
	Width      float64
	SpaceW     float64

	TextBox *LayoutBox // IFCTextWord: source text node LayoutBox
	SpanBox *LayoutBox // IFCSpanStart / IFCSpanEnd: the span LayoutBox

	// Typography (IFCTextWord only)
	FontFamily string
	FontSize   float64
	FontWeight string
	FontStyle  string
	Ascent     float64
	LineHeight float64
}

// ---- Line box output ----

// LineFragment is one positioned text run inside a LineBox.
type LineFragment struct {
	TextBox    *LayoutBox   // source text node LayoutBox
	Spans      []*LayoutBox // span ancestors (innermost last); used to update span Dimensions
	Text       string
	X, Y       float64
	Width      float64
	LineHeight float64 // height of the containing line box (needed for span Dimensions)
	Ascent     float64
	FontFamily string
	FontSize   float64
	FontWeight string
	FontStyle  string
}

// LineBox holds all fragments for one visual line.
type LineBox struct {
	Fragments []LineFragment
	X, Y      float64
	Width     float64 // total used width
	Height    float64 // line box height
	Ascent    float64 // max ascent (baseline reference)
}

// ---- Collector ----

// collectInlineItems flattens box's inline children into an InlineItem stream.
func collectInlineItems(box *LayoutBox, out *[]InlineItem, ctx *LayoutContext) {
	measurer := ctx.GetTextMeasurer()

	for _, child := range box.Children {
		sn := child.StyledNode
		if sn == nil {
			continue
		}

		// Text node
		if sn.Node.Type == 0 {
			text := sn.Node.Content

			// Whitespace-only: emit a soft-space marker instead of dropping it.
			// This preserves the inter-element space for <span>foo</span> <span>bar</span>.
			if strings.TrimSpace(text) == "" {
				*out = append(*out, InlineItem{Kind: IFCSoftSpace})
				continue
			}

			fontFamily := child.getValue("font-family")
			fontSize := ParseLengthWithContext(child.getValue("font-size"), 16, child, ctx)
			fontWeight := child.getValue("font-weight")
			fontStyle := child.getValue("font-style")

			spaceW := measurer.MeasureWord(" ", fontFamily, fontSize, fontWeight, fontStyle)

			hasLeading := len(text) > 0 && unicode.IsSpace(rune(text[0]))
			hasTrailing := len(text) > 0 && unicode.IsSpace(rune(text[len(text)-1]))
			words := strings.Fields(text)
			if len(words) == 0 {
				continue
			}

			for i, w := range words {
				// Single MeasureText call for all metrics.
				m := measurer.MeasureText(w, fontFamily, fontSize, fontWeight, fontStyle)
				hasSpace := i < len(words)-1 || hasTrailing

				*out = append(*out, InlineItem{
					Kind:       IFCTextWord,
					Word:       w,
					HasSpace:   hasSpace,
					HasLeading: i == 0 && hasLeading,
					Width:      m.Width,
					SpaceW:     spaceW,
					TextBox:    child,
					FontFamily: fontFamily,
					FontSize:   fontSize,
					FontWeight: fontWeight,
					FontStyle:  fontStyle,
					Ascent:     m.Ascent,
					LineHeight: m.LineHeight,
				})
			}
			continue
		}

		// Block child inside an inline context: skip.
		// TODO: CSS requires promoting this to an anonymous block box.
		if child.BoxType != InlineBox {
			continue
		}

		// Inline element (span etc.)
		*out = append(*out, InlineItem{Kind: IFCSpanStart, SpanBox: child})
		collectInlineItems(child, out, ctx)
		*out = append(*out, InlineItem{Kind: IFCSpanEnd, SpanBox: child})
	}
}

// ---- Line Breaker ----

// breakIntoLines runs the greedy IFC line-breaking algorithm and returns LineBoxes.
// textAlign ("left"/"center"/"right") shifts fragment X positions per line.
// TODO: add "justify" support (distribute free space as inter-word gaps).
func breakIntoLines(items []InlineItem, availWidth, startX, startY float64, textAlign string) []LineBox {
	var lines []LineBox

	curX := startX
	curY := startY
	curAscent := 0.0
	curLineH := 0.0
	var curFrags []LineFragment

	// Active fragment accumulation (consecutive words from the same source TextBox
	// within the same line).
	var fragTextBox *LayoutBox
	var fragSpans []*LayoutBox // copy of spanStack at fragment start
	var fragWords []string
	var fragX float64
	var fragW float64
	var fragAscent float64
	var fragFontFamily string
	var fragFontSize float64
	var fragFontWeight string
	var fragFontStyle string

	// span stack: open span boxes, innermost last
	var spanStack []*LayoutBox

	// pendingSpace: a space should precede the next word placed on the line.
	pendingSpace := false

	flushFrag := func() {
		if fragTextBox == nil || len(fragWords) == 0 {
			return
		}
		// fragSpans was already snapshot-copied at capture time; use directly.
		curFrags = append(curFrags, LineFragment{
			TextBox:    fragTextBox,
			Spans:      fragSpans,
			Text:       strings.Join(fragWords, " "),
			X:          fragX,
			Y:          curY,
			Width:      fragW,
			LineHeight: 0, // filled in by flushLine
			Ascent:     fragAscent,
			FontFamily: fragFontFamily,
			FontSize:   fragFontSize,
			FontWeight: fragFontWeight,
			FontStyle:  fragFontStyle,
		})
		fragTextBox = nil
		fragWords = nil
		fragW = 0
	}

	flushLine := func() {
		flushFrag()
		if len(curFrags) == 0 {
			// Reset per-line accumulators even when nothing was emitted so that
			// back-to-back fragment-less flushLine calls stay consistent.
			curX = startX
			curAscent = 0
			curLineH = 0
			pendingSpace = false
			return
		}
		h := curLineH
		if h <= 0 {
			h = curAscent * 1.2
		}
		usedW := curX - startX

		// Stamp line height into every fragment (needed for span Dimensions).
		for i := range curFrags {
			curFrags[i].LineHeight = h
		}

		// Apply text-align X offset.
		freeSpace := availWidth - usedW
		var offsetX float64
		switch textAlign {
		case "center":
			offsetX = freeSpace / 2
		case "right":
			offsetX = freeSpace
		}
		if offsetX > 0 {
			for i := range curFrags {
				curFrags[i].X += offsetX
			}
		}

		lines = append(lines, LineBox{
			Fragments: curFrags,
			X:         startX,
			Y:         curY,
			Width:     usedW,
			Height:    h,
			Ascent:    curAscent,
		})
		curFrags = nil
		curY += h
		curX = startX
		curAscent = 0
		curLineH = 0
		pendingSpace = false
	}

	for _, item := range items {
		switch item.Kind {
		case IFCSoftSpace:
			// Inter-element whitespace: signal a space before the next word.
			pendingSpace = true

		case IFCSpanStart:
			// Flush active fragment — new span means potentially different style.
			flushFrag()
			spanStack = append(spanStack, item.SpanBox)

		case IFCSpanEnd:
			flushFrag()
			if len(spanStack) > 0 {
				spanStack = spanStack[:len(spanStack)-1]
			}

		case IFCTextWord:
			// Determine preceding space width.
			needsSpace := pendingSpace || item.HasLeading
			spaceW := 0.0
			if curX > startX && needsSpace {
				spaceW = item.SpaceW
			}

			// Wrap if word overflows available width.
			if curX+spaceW+item.Width > startX+availWidth && curX > startX {
				flushLine()
				spaceW = 0
			}

			// Update line height metrics AFTER any wrap so they apply to the
			// line this word actually lands on (flushLine resets them).
			if item.Ascent > curAscent {
				curAscent = item.Ascent
			}
			if item.LineHeight > curLineH {
				curLineH = item.LineHeight
			}

			curX += spaceW

			// Continue the active fragment or start a new one.
			if item.TextBox != fragTextBox {
				flushFrag()
				fragTextBox = item.TextBox
				// Snapshot the span stack immediately so it's unaffected by future
				// IFCSpanEnd shrinks or sibling IFCSpanStart appends.
				fragSpans = append([]*LayoutBox(nil), spanStack...)
				fragX = curX
				fragFontFamily = item.FontFamily
				fragFontSize = item.FontSize
				fragFontWeight = item.FontWeight
				fragFontStyle = item.FontStyle
				fragAscent = item.Ascent
				fragW = 0
			} else if len(fragWords) > 0 {
				// Same TextBox on same line: include inter-word space in fragment width.
				fragW += spaceW
				// Keep fragAscent as the max across all words in this fragment.
				if item.Ascent > fragAscent {
					fragAscent = item.Ascent
				}
			}

			fragWords = append(fragWords, item.Word)
			fragW += item.Width
			curX += item.Width
			pendingSpace = item.HasSpace
		}
	}

	flushLine()
	return lines
}

// ---- IFC Layout Entry Point ----

// layoutAnonymousIFC is the IFC-based implementation of anonymous box layout.
func (box *LayoutBox) layoutAnonymousIFC(containingBlock Dimensions, ctx *LayoutContext) {
	box.Dimensions.Content.Width = containingBlock.Content.Width
	box.Dimensions.Content.X = containingBlock.Content.X
	box.Dimensions.Content.Y = containingBlock.Content.Y
	box.Dimensions.Content.Height = 0

	var items []InlineItem
	collectInlineItems(box, &items, ctx)

	if len(items) == 0 {
		return
	}

	// Resolve text-align from parent block container (anonymous box has no StyledNode).
	textAlign := ""
	if p := box.Parent; p != nil && p.StyledNode != nil {
		textAlign = p.StyledNode.SpecifiedValues["text-align"]
	}

	lines := breakIntoLines(items, containingBlock.Content.Width,
		containingBlock.Content.X, containingBlock.Content.Y, textAlign)

	box.LineBoxes = lines

	totalH := 0.0
	for _, lb := range lines {
		totalH += lb.Height
	}
	box.Dimensions.Content.Height = totalH

	// Update child LayoutBox Dimensions from IFC results.
	ifcSetChildPositions(lines)
}

// ifcSetChildPositions updates Dimensions on all source LayoutBoxes referenced in
// the line boxes. Both text nodes and span boxes are processed in a single pass.
//
// TODO: text nodes that span multiple lines get a union bounding rect here.
// This means hit-testing may fire for the empty space between the two fragments
// (e.g. a short second line). Fix requires storing per-fragment rects on the
// LayoutBox rather than a single Dimensions.
//
// TODO: span Dimensions do not account for padding/border; the CSS inline box model
// requires per-fragment border rendering (border-left on first, border-right on last
// fragment) — tracked for a future inline-box fragmentation pass.
func ifcSetChildPositions(lines []LineBox) {
	type bbox struct {
		minX, minY, maxX, maxY float64
		ascent                 float64
	}
	textBB := make(map[*LayoutBox]*bbox)
	spanBB := make(map[*LayoutBox]*bbox)

	expandBB := func(m map[*LayoutBox]*bbox, box *LayoutBox, frag LineFragment) {
		bottom := frag.Y + frag.LineHeight
		right := frag.X + frag.Width
		bb, ok := m[box]
		if !ok {
			m[box] = &bbox{frag.X, frag.Y, right, bottom, frag.Ascent}
			return
		}
		if frag.X < bb.minX {
			bb.minX = frag.X
		}
		if frag.Y < bb.minY {
			bb.minY = frag.Y
		}
		if right > bb.maxX {
			bb.maxX = right
		}
		if bottom > bb.maxY {
			bb.maxY = bottom
		}
		if frag.Ascent > bb.ascent {
			bb.ascent = frag.Ascent
		}
	}

	for _, lb := range lines {
		for _, frag := range lb.Fragments {
			expandBB(textBB, frag.TextBox, frag)
			for _, span := range frag.Spans {
				expandBB(spanBB, span, frag)
			}
		}
	}

	for tb, bb := range textBB {
		tb.Dimensions.Content.X = bb.minX
		tb.Dimensions.Content.Y = bb.minY
		tb.Dimensions.Content.Width = bb.maxX - bb.minX
		tb.Dimensions.Content.Height = bb.maxY - bb.minY
		tb.TextAscent = bb.ascent
	}

	for span, bb := range spanBB {
		span.Dimensions.Content.X = bb.minX
		span.Dimensions.Content.Y = bb.minY
		span.Dimensions.Content.Width = bb.maxX - bb.minX
		span.Dimensions.Content.Height = bb.maxY - bb.minY
	}
}
