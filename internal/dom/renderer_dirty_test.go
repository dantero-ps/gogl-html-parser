package dom

import (
	"github.com/furkandgn/goglweb/internal/parser/css"
	"github.com/furkandgn/goglweb/internal/parser/html"
	"testing"
)

func TestRendererDirtyFlag(t *testing.T) {
	root := &html.Node{Type: html.ElementNode, TagName: "html"}
	ss := &css.Stylesheet{}
	r := NewRenderer(root, ss, 800, 600)

	if r.IsDirty() {
		t.Error("fresh renderer after NewRenderer should not be dirty (Rebuild clears NeedsLayout, NeedsRepaint was never set)")
	}

	r.MarkDirty()
	if !r.IsDirty() {
		t.Error("renderer should be dirty after MarkDirty()")
	}

	r.Rebuild()
	r.ClearDirty()
	if r.IsDirty() {
		t.Error("renderer should not be dirty after Rebuild() + ClearDirty()")
	}
}

func TestClearDirtyOnFreshRenderer(t *testing.T) {
	root := &html.Node{Type: html.ElementNode, TagName: "html"}
	ss := &css.Stylesheet{}
	r := NewRenderer(root, ss, 800, 600)

	r.ClearDirty()
	if r.IsDirty() {
		t.Error("ClearDirty() on fresh renderer should leave IsDirty() false")
	}
}
