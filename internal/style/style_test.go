package style

import (
	"github.com/furkandgn/goglweb/internal/parser/css"
	"github.com/furkandgn/goglweb/internal/parser/html"
	"testing"
)

func TestComputeStyle(t *testing.T) {
	node := &html.Node{
		TagName: "div",
		Attr:    map[string]string{"class": "btn", "id": "main"},
		Type:    1,
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "div", Declarations: []css.Declaration{{Property: "display", Value: "block"}}},
			{Selector: ".btn", Declarations: []css.Declaration{{Property: "color", Value: "red"}}},
			{Selector: "#main", Declarations: []css.Declaration{{Property: "color", Value: "blue"}}},
		},
	}

	style := ComputeStyle(node, stylesheet)

	// Cascade testi: Sonraki kural veya daha spesifik olan (bu basit yapıda son gelen) kazanır
	if style["color"] != "blue" {
		t.Errorf("Beklenen color blue, alınan %s", style["color"])
	}

	if style["display"] != "block" {
		t.Errorf("Beklenen display block, alınan %s", style["display"])
	}
}

func TestBuildStyledTree(t *testing.T) {
	// <div class="container"><p></p></div>
	root := &html.Node{
		TagName: "div",
		Attr:    map[string]string{"class": "container"},
		Type:    1,
		Children: []*html.Node{
			{TagName: "p", Type: 1},
		},
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: ".container", Declarations: []css.Declaration{{Property: "padding", Value: "10px"}}},
			{Selector: "p", Declarations: []css.Declaration{{Property: "margin", Value: "5px"}}},
		},
	}

	styledTree := BuildStyledTree(root, stylesheet)

	if styledTree.SpecifiedValues["padding"] != "10px" {
		t.Error("Root stili hatalı")
	}

	if len(styledTree.Children) != 1 {
		t.Fatalf("Çocuk düğüm sayısı hatalı: %d", len(styledTree.Children))
	}

	if styledTree.Children[0].SpecifiedValues["margin"] != "5px" {
		t.Error("Çocuk düğüm stili hatalı")
	}
}

func TestDisplayNoneFiltering(t *testing.T) {
	root := &html.Node{
		TagName: "div",
		Type:    1,
		Children: []*html.Node{
			{TagName: "span", Type: 1}, // Gizlenecek
		},
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "span", Declarations: []css.Declaration{{Property: "display", Value: "none"}}},
		},
	}

	styledTree := BuildStyledTree(root, stylesheet)

	if len(styledTree.Children) != 0 {
		t.Errorf("display:none olan eleman ağaçtan elenmeliydi, mevcut çocuk: %d", len(styledTree.Children))
	}
}
