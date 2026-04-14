package goglweb

import (
	"github.com/furkandgn/goglweb/internal/dom"
	"github.com/furkandgn/goglweb/internal/layout"
	"github.com/furkandgn/goglweb/internal/parser/css"
	"github.com/furkandgn/goglweb/internal/parser/html"
	"github.com/furkandgn/goglweb/internal/render"
)

// Painter is the drawing interface. Implement this for custom rendering backends.
type Painter = render.Painter

// Rect represents a rectangle in layout coordinates.
type Rect = layout.Rect

// Color represents an RGBA color.
type Color = render.Color

// EdgeSizes represents box edge dimensions (padding, border, margin).
type EdgeSizes = layout.EdgeSizes

// NodeRef is an opaque handle to a DOM node. Internal types are not exposed.
type NodeRef struct {
	node *html.Node
}

// HitTestResult represents the result of a hit test query.
type HitTestResult struct {
	Node NodeRef
	X, Y float64
}

// Renderer is the low-level rendering engine. Use this when you own the GL context.
type Renderer struct {
	renderer   *dom.Renderer
	htmlRoot   *html.Node
	stylesheet *css.Stylesheet
	cfg        config
}

// NewRenderer creates a new rendering engine from HTML and CSS source strings.
func NewRenderer(htmlSource, cssSource string, opts ...Option) (*Renderer, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	htmlParser := html.NewParser(htmlSource)
	htmlRoot := htmlParser.Parse()

	cssParser := css.NewParser(cssSource)
	stylesheet := cssParser.Parse()

	renderer := dom.NewRenderer(htmlRoot, stylesheet, cfg.viewportWidth, cfg.viewportHeight)

	return &Renderer{
		renderer:   renderer,
		htmlRoot:   htmlRoot,
		stylesheet: stylesheet,
		cfg:        cfg,
	}, nil
}

// Root returns a NodeRef for the root HTML node.
func (r *Renderer) Root() NodeRef {
	return NodeRef{node: r.htmlRoot}
}

// AppendChild adds a child node to a parent node.
func (r *Renderer) AppendChild(parent, child NodeRef) {
	dom.AppendChild(parent.node, child.node)
}

// RemoveChild removes a child node from a parent node.
func (r *Renderer) RemoveChild(parent, child NodeRef) bool {
	return dom.RemoveChild(parent.node, child.node)
}

// SetAttribute sets an attribute on a node.
func (r *Renderer) SetAttribute(node NodeRef, key, value string) {
	dom.SetAttribute(node.node, key, value)
}

// GetAttribute returns the value of a node's attribute.
func (r *Renderer) GetAttribute(node NodeRef, key string) string {
	return dom.GetAttribute(node.node, key)
}

// AddClass adds a CSS class to a node.
func (r *Renderer) AddClass(node NodeRef, class string) {
	dom.AddClass(node.node, class)
}

// RemoveClass removes a CSS class from a node.
func (r *Renderer) RemoveClass(node NodeRef, class string) {
	dom.RemoveClass(node.node, class)
}

// ToggleClass toggles a CSS class on a node.
func (r *Renderer) ToggleClass(node NodeRef, class string) {
	dom.ToggleClass(node.node, class)
}

// SetTextContent sets the text content of a node.
func (r *Renderer) SetTextContent(node NodeRef, text string) {
	dom.SetTextContent(node.node, text)
}

// GetTextContent returns the text content of a node.
func (r *Renderer) GetTextContent(node NodeRef) string {
	return dom.GetTextContent(node.node)
}

// NewElement creates a new element node with the given tag name.
func (r *Renderer) NewElement(tag string) NodeRef {
	return NodeRef{node: html.NewElement(tag)}
}

// NewText creates a new text node with the given content.
func (r *Renderer) NewText(text string) NodeRef {
	return NodeRef{node: html.NewText(text)}
}

// FindNodeByID walks the DOM tree and returns the first node with the given id attribute.
func (r *Renderer) FindNodeByID(id string) NodeRef {
	n := findNodeByID(r.htmlRoot, id)
	if n == nil {
		return NodeRef{}
	}
	return NodeRef{node: n}
}

// findNodeByID recursively walks the DOM tree looking for a node with the given id.
func findNodeByID(root *html.Node, id string) *html.Node {
	if root == nil {
		return nil
	}
	if dom.GetAttribute(root, "id") == id {
		return root
	}
	for _, child := range root.Children {
		if result := findNodeByID(child, id); result != nil {
			return result
		}
	}
	return nil
}

// MarkDirty marks that the renderer needs to recalculate layout and repaint.
func (r *Renderer) MarkDirty() {
	r.renderer.MarkDirty()
}

// IsDirty returns true if layout or repaint is needed.
func (r *Renderer) IsDirty() bool {
	return r.renderer.IsDirty()
}

// ClearDirty resets the dirty flag after a successful render.
func (r *Renderer) ClearDirty() {
	r.renderer.ClearDirty()
}

// UpdateStylesheet parses new CSS source and updates the renderer's stylesheet.
func (r *Renderer) UpdateStylesheet(cssSource string) {
	cssParser := css.NewParser(cssSource)
	stylesheet := cssParser.Parse()
	r.stylesheet = stylesheet
	r.renderer.UpdateStylesheet(stylesheet)
}

// UpdateViewport updates the viewport dimensions.
func (r *Renderer) UpdateViewport(w, h float64) {
	r.renderer.UpdateViewport(w, h)
}

// Render executes one frame: builds display list (if dirty) and renders via the Painter.
func (r *Renderer) Render(painter Painter) {
	dl := r.renderer.GetDisplayList()
	dl.Execute(painter)
}

// HitTest performs a hit test at the given coordinates.
func (r *Renderer) HitTest(x, y float64) *HitTestResult {
	layoutTree := r.renderer.GetLayoutTree()
	if layoutTree == nil {
		return nil
	}
	result := dom.HitTest(layoutTree, x, y)
	if result == nil {
		return nil
	}
	return &HitTestResult{
		Node: NodeRef{node: result.HTMLNode},
		X:    result.Point.X,
		Y:    result.Point.Y,
	}
}

// ClampScrollOffset clamps a scroll offset to valid bounds.
func ClampScrollOffset(offset, contentHeight, visibleHeight float64) float64 {
	return layout.ClampScrollOffset(offset, contentHeight, visibleHeight)
}
