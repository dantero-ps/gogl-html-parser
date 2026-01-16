package dom

import (
	"goglweb/internal/layout"
	"goglweb/internal/parser/css"
	"goglweb/internal/parser/html"
	"goglweb/internal/render"
	"goglweb/internal/style"
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

// Rebuild recreates the entire rendering pipeline
func (r *Renderer) Rebuild() {
	// 1. Rebuild styled tree
	r.StyledTree = style.BuildStyledTree(r.HTMLRoot, r.Stylesheet)

	// 2. Rebuild layout tree
	r.LayoutTree = layout.BuildLayoutTree(r.StyledTree)

	// 3. Calculate layout
	r.LayoutTree.LayoutWithContext(r.ContainingBlock, r.LayoutCtx)

	// 4. Create display list
	r.DisplayList = render.BuildDisplayList(r.LayoutTree)

	r.NeedsLayout = false
}

// MarkDirty marks that layout needs to be recalculated
func (r *Renderer) MarkDirty() {
	r.NeedsLayout = true
}

// UpdateStylesheet updates the stylesheet and rebuilds
func (r *Renderer) UpdateStylesheet(stylesheet *css.Stylesheet) {
	r.Stylesheet = stylesheet
	r.MarkDirty()
}

// UpdateViewport updates the viewport size
func (r *Renderer) UpdateViewport(width, height float64) {
	r.LayoutCtx = layout.NewLayoutContext(width, height)
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
