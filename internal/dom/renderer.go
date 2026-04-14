package dom

import (
	"github.com/furkandgn/goglweb/internal/layout"
	"github.com/furkandgn/goglweb/internal/parser/css"
	"github.com/furkandgn/goglweb/internal/parser/html"
	"github.com/furkandgn/goglweb/internal/render"
	"github.com/furkandgn/goglweb/internal/style"
)

// Renderer manages DOM changes and coordinates re-rendering
type Renderer struct {
	HTMLRoot        *html.Node
	Stylesheet      *css.Stylesheet
	StyledTree      *style.StyledNode
	LayoutTree      *layout.LayoutBox
	DisplayList     *render.DisplayList
	LayoutCtx       *layout.LayoutContext
	ContainingBlock layout.Dimensions
	NeedsLayout     bool
	NeedsRepaint    bool
}

// NewRenderer creates a new renderer
func NewRenderer(htmlRoot *html.Node, stylesheet *css.Stylesheet, viewportWidth, viewportHeight float64) *Renderer {
	r := &Renderer{
		HTMLRoot:    htmlRoot,
		Stylesheet:  stylesheet,
		NeedsLayout: true,
	}
	r.LayoutCtx = layout.NewLayoutContext(viewportWidth, viewportHeight)
	r.ContainingBlock = layout.Dimensions{
		Content: layout.Rect{
			X:      0,
			Y:      0,
			Width:  viewportWidth,
			Height: viewportHeight,
		},
	}
	r.Rebuild()
	return r
}

// collectScrollOffsets walks the layout tree and saves scroll offsets for
// every scrollable box, keyed by its stable *html.Node pointer.
func collectScrollOffsets(box *layout.LayoutBox, out map[*html.Node][2]float64) {
	if box == nil {
		return
	}
	if box.IsScrollable() && box.StyledNode != nil && box.StyledNode.Node != nil {
		out[box.StyledNode.Node] = [2]float64{box.ScrollOffsetX, box.ScrollOffsetY}
	}
	for _, c := range box.Children {
		collectScrollOffsets(c, out)
	}
}

// restoreScrollOffsets walks the rebuilt layout tree and re-applies saved offsets
// by matching the stable *html.Node pointer.
func restoreScrollOffsets(box *layout.LayoutBox, saved map[*html.Node][2]float64) {
	if box == nil {
		return
	}
	if box.IsScrollable() && box.StyledNode != nil && box.StyledNode.Node != nil {
		if offsets, ok := saved[box.StyledNode.Node]; ok {
			box.ScrollOffsetX = offsets[0]
			box.ScrollOffsetY = offsets[1]
		}
	}
	for _, c := range box.Children {
		restoreScrollOffsets(c, saved)
	}
}

// Rebuild recreates the entire rendering pipeline
func (r *Renderer) Rebuild() {
	// Save scroll offsets before rebuilding so they survive the new layout tree.
	scrollOffsets := make(map[*html.Node][2]float64)
	if r.LayoutTree != nil {
		collectScrollOffsets(r.LayoutTree, scrollOffsets)
	}

	// 1. Rebuild styled tree
	r.StyledTree = style.BuildStyledTree(r.HTMLRoot, r.Stylesheet)

	// 2. Rebuild layout tree
	r.LayoutTree = layout.BuildLayoutTree(r.StyledTree)

	// 3. Calculate layout
	r.LayoutTree.LayoutWithContext(r.ContainingBlock, r.LayoutCtx)

	// Restore scroll offsets onto the new layout tree.
	if len(scrollOffsets) > 0 {
		restoreScrollOffsets(r.LayoutTree, scrollOffsets)
	}

	// 4. Create display list (after restoring offsets so BeginScrollCmd gets correct values)
	r.DisplayList = render.BuildDisplayList(r.LayoutTree)

	r.NeedsLayout = false
}

// MarkDirty marks that layout needs to be recalculated and the screen repainted.
func (r *Renderer) MarkDirty() {
	r.NeedsLayout = true
	r.NeedsRepaint = true
}

// IsDirty returns true if layout or repaint is needed.
func (r *Renderer) IsDirty() bool {
	return r.NeedsLayout || r.NeedsRepaint
}

// ClearDirty resets the repaint flag after a successful render.
func (r *Renderer) ClearDirty() {
	r.NeedsRepaint = false
}

// UpdateStylesheet updates the stylesheet and rebuilds
func (r *Renderer) UpdateStylesheet(stylesheet *css.Stylesheet) {
	r.Stylesheet = stylesheet
	r.MarkDirty()
}

// UpdateViewport updates the viewport size
func (r *Renderer) UpdateViewport(width, height float64) {
	// Preserve TextMeasurer so font metrics survive the new context
	prevMeasurer := r.LayoutCtx.TextMeasurer
	r.LayoutCtx = layout.NewLayoutContext(width, height)
	r.LayoutCtx.TextMeasurer = prevMeasurer
	r.ContainingBlock = layout.Dimensions{
		Content: layout.Rect{
			X:      0,
			Y:      0,
			Width:  width,
			Height: height,
		},
	}
	r.MarkDirty()
}

// GetDisplayList returns the display list (rebuilds if necessary)
func (r *Renderer) GetDisplayList() *render.DisplayList {
	if r.NeedsLayout {
		r.Rebuild()
	}
	return r.DisplayList
}

// GetLayoutTree returns the layout tree (rebuilds if necessary)
func (r *Renderer) GetLayoutTree() *layout.LayoutBox {
	if r.NeedsLayout {
		r.Rebuild()
	}
	return r.LayoutTree
}
