package layout

import (
	"goglweb/internal/parser/css"
	"goglweb/internal/parser/html"
	"goglweb/internal/style"
	"strings"
	"testing"
)

func TestBuildLayoutTree(t *testing.T) {
	// <div><p>Hello</p></div>
	root := &html.Node{
		TagName: "div",
		Type:    html.ElementNode,
		Children: []*html.Node{
			{
				TagName: "p",
				Type:    html.ElementNode,
				Children: []*html.Node{
					{Type: html.TextNode, Content: "Hello"},
				},
			},
		},
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "div", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "100px"},
				{Property: "height", Value: "50px"},
			}},
			{Selector: "p", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	if layoutTree == nil {
		t.Fatal("Layout tree nil olmamalı")
	}

	if layoutTree.BoxType != BlockBox {
		t.Errorf("Root box type block olmalı, alınan: %v", layoutTree.BoxType)
	}

	if len(layoutTree.Children) != 1 {
		t.Fatalf("Çocuk box sayısı 1 olmalı, alınan: %d", len(layoutTree.Children))
	}

	if layoutTree.Children[0].BoxType != BlockBox {
		t.Errorf("Çocuk box type block olmalı, alınan: %v", layoutTree.Children[0].BoxType)
	}
}

func TestBlockLayout(t *testing.T) {
	// <div style="width: 200px; height: 100px;"><p style="width: 150px;"></p></div>
	root := &html.Node{
		TagName: "div",
		Type:    html.ElementNode,
		Children: []*html.Node{
			{TagName: "p", Type: html.ElementNode},
		},
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "div", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "200px"},
				{Property: "height", Value: "100px"},
			}},
			{Selector: "p", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "150px"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}

	layoutTree.Layout(containingBlock)

	if layoutTree.Dimensions.Content.Width != 200 {
		t.Errorf("Root genişliği 200px olmalı, alınan: %f", layoutTree.Dimensions.Content.Width)
	}

	if layoutTree.Dimensions.Content.Height != 100 {
		t.Errorf("Root yüksekliği 100px olmalı, alınan: %f", layoutTree.Dimensions.Content.Height)
	}

	if len(layoutTree.Children) > 0 {
		childBox := layoutTree.Children[0]
		if childBox.Dimensions.Content.Width != 150 {
			t.Errorf("Çocuk genişliği 150px olmalı, alınan: %f", childBox.Dimensions.Content.Width)
		}
	}
}

func TestInlineLayout(t *testing.T) {
	// <p><span>Hello</span> <span>World</span></p>
	root := &html.Node{
		TagName: "p",
		Type:    html.ElementNode,
		Children: []*html.Node{
			{
				TagName:  "span",
				Type:     html.ElementNode,
				Children: []*html.Node{{Type: html.TextNode, Content: "Hello"}},
			},
			{
				TagName:  "span",
				Type:     html.ElementNode,
				Children: []*html.Node{{Type: html.TextNode, Content: " World"}},
			},
		},
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "p", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
			}},
			{Selector: "span", Declarations: []css.Declaration{
				{Property: "display", Value: "inline"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}

	layoutTree.Layout(containingBlock)

	if len(layoutTree.Children) == 0 {
		t.Fatal("Inline box'lar oluşturulmalı")
	}
}

func TestMarginPaddingBorder(t *testing.T) {
	root := &html.Node{
		TagName: "div",
		Type:    html.ElementNode,
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "div", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "100px"},
				{Property: "height", Value: "50px"},
				{Property: "margin", Value: "10px"},
				{Property: "padding", Value: "5px"},
				{Property: "border", Value: "2px solid black"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}

	layoutTree.Layout(containingBlock)

	// Margin kontrolü
	if layoutTree.Dimensions.Margin.Top != 10 {
		t.Errorf("Margin top 10px olmalı, alınan: %f", layoutTree.Dimensions.Margin.Top)
	}

	// Padding kontrolü
	if layoutTree.Dimensions.Padding.Top != 5 {
		t.Errorf("Padding top 5px olmalı, alınan: %f", layoutTree.Dimensions.Padding.Top)
	}

	// Border kontrolü
	if layoutTree.Dimensions.Border.Top != 2 {
		t.Errorf("Border top 2px olmalı, alınan: %f", layoutTree.Dimensions.Border.Top)
	}
}

func TestAutoWidth(t *testing.T) {
	root := &html.Node{
		TagName: "div",
		Type:    html.ElementNode,
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "div", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "auto"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}

	layoutTree.Layout(containingBlock)

	// Auto width, containing block'un genişliğini almalı
	if layoutTree.Dimensions.Content.Width != 800 {
		t.Errorf("Auto width containing block genişliğini almalı (800px), alınan: %f", layoutTree.Dimensions.Content.Width)
	}
}

func TestVerticalStacking(t *testing.T) {
	// <div><p></p><p></p></div>
	root := &html.Node{
		TagName: "div",
		Type:    html.ElementNode,
		Children: []*html.Node{
			{TagName: "p", Type: html.ElementNode},
			{TagName: "p", Type: html.ElementNode},
		},
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "div", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "200px"},
			}},
			{Selector: "p", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "height", Value: "50px"},
				{Property: "margin", Value: "10px"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}

	layoutTree.Layout(containingBlock)

	if len(layoutTree.Children) != 2 {
		t.Fatalf("2 çocuk box olmalı, alınan: %d", len(layoutTree.Children))
	}

	firstChild := layoutTree.Children[0]
	secondChild := layoutTree.Children[1]

	// İkinci çocuk, birincinin altında olmalı
	if secondChild.Dimensions.Content.Y <= firstChild.Dimensions.Content.Y+firstChild.Dimensions.Content.Height {
		t.Errorf("Block elementler dikey olarak sıralanmalı")
	}
}

func TestAnonymousBoxCreation(t *testing.T) {
	// <div><span>Hello</span><span>World</span></div>
	root := &html.Node{
		TagName: "div",
		Type:    html.ElementNode,
		Children: []*html.Node{
			{
				TagName:  "span",
				Type:     html.ElementNode,
				Children: []*html.Node{{Type: html.TextNode, Content: "Hello"}},
			},
			{
				TagName:  "span",
				Type:     html.ElementNode,
				Children: []*html.Node{{Type: html.TextNode, Content: "World"}},
			},
		},
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "div", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
			}},
			{Selector: "span", Declarations: []css.Declaration{
				{Property: "display", Value: "inline"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	if len(layoutTree.Children) != 1 {
		t.Fatalf("Anonymous box oluşturulmalı, çocuk sayısı 1 olmalı, alınan: %d", len(layoutTree.Children))
	}

	anonymousBox := layoutTree.Children[0]
	if anonymousBox.BoxType != AnonymousBox {
		t.Errorf("Anonymous box type olmalı, alınan: %v", anonymousBox.BoxType)
	}

	if len(anonymousBox.Children) != 2 {
		t.Errorf("Anonymous box içinde 2 inline box olmalı, alınan: %d", len(anonymousBox.Children))
	}
}

func TestPercentageWidth(t *testing.T) {
	root := &html.Node{
		TagName: "div",
		Type:    html.ElementNode,
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "div", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "50%"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}

	layoutTree.Layout(containingBlock)

	expectedWidth := 800 * 0.5
	if layoutTree.Dimensions.Content.Width != expectedWidth {
		t.Errorf("Yüzde genişlik doğru hesaplanmalı, beklenen: %f, alınan: %f", expectedWidth, layoutTree.Dimensions.Content.Width)
	}
}

func TestNestedBlocks(t *testing.T) {
	// <div><div><div></div></div></div>
	root := &html.Node{
		TagName: "div",
		Type:    html.ElementNode,
		Children: []*html.Node{
			{
				TagName: "div",
				Type:    html.ElementNode,
				Children: []*html.Node{
					{TagName: "div", Type: html.ElementNode},
				},
			},
		},
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "div", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "100px"},
				{Property: "height", Value: "50px"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}

	layoutTree.Layout(containingBlock)

	if len(layoutTree.Children) != 1 {
		t.Fatalf("1 çocuk olmalı, alınan: %d", len(layoutTree.Children))
	}

	firstChild := layoutTree.Children[0]
	if firstChild.Dimensions.Content.Width != 100 {
		t.Errorf("İç içe block genişliği 100px olmalı, alınan: %f", firstChild.Dimensions.Content.Width)
	}

	if len(firstChild.Children) != 1 {
		t.Fatalf("İç içe block'un 1 çocuğu olmalı, alınan: %d", len(firstChild.Children))
	}
}

func TestEmUnit(t *testing.T) {
	root := &html.Node{
		TagName: "div",
		Type:    html.ElementNode,
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "div", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "font-size", Value: "20px"},
				{Property: "width", Value: "10em"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	ctx := NewLayoutContext(800, 600)
	ctx.RootFontSize = 16.0

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}

	layoutTree.LayoutWithContext(containingBlock, ctx)

	expectedWidth := 20.0 * 10.0
	if layoutTree.Dimensions.Content.Width != expectedWidth {
		t.Errorf("em birimi doğru hesaplanmalı, beklenen: %f, alınan: %f", expectedWidth, layoutTree.Dimensions.Content.Width)
	}
}

func TestRemUnit(t *testing.T) {
	root := &html.Node{
		TagName: "div",
		Type:    html.ElementNode,
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "div", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "5rem"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	ctx := NewLayoutContext(800, 600)
	ctx.RootFontSize = 16.0

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}

	layoutTree.LayoutWithContext(containingBlock, ctx)

	expectedWidth := 16.0 * 5.0
	if layoutTree.Dimensions.Content.Width != expectedWidth {
		t.Errorf("rem birimi doğru hesaplanmalı, beklenen: %f, alınan: %f", expectedWidth, layoutTree.Dimensions.Content.Width)
	}
}

func TestViewportUnits(t *testing.T) {
	root := &html.Node{
		TagName: "div",
		Type:    html.ElementNode,
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "div", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "50vw"},
				{Property: "height", Value: "30vh"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	ctx := NewLayoutContext(1000, 800)
	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 1000, Height: 800},
	}

	layoutTree.LayoutWithContext(containingBlock, ctx)

	expectedWidth := 1000.0 * 0.5
	expectedHeight := 800.0 * 0.3

	if layoutTree.Dimensions.Content.Width != expectedWidth {
		t.Errorf("vw birimi doğru hesaplanmalı, beklenen: %f, alınan: %f", expectedWidth, layoutTree.Dimensions.Content.Width)
	}

	if layoutTree.Dimensions.Content.Height != expectedHeight {
		t.Errorf("vh birimi doğru hesaplanmalı, beklenen: %f, alınan: %f", expectedHeight, layoutTree.Dimensions.Content.Height)
	}
}

func TestVminVmaxUnits(t *testing.T) {
	root := &html.Node{
		TagName: "div",
		Type:    html.ElementNode,
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "div", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "20vmin"},
				{Property: "height", Value: "20vmax"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	ctx := NewLayoutContext(800, 1000)
	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 1000},
	}

	layoutTree.LayoutWithContext(containingBlock, ctx)

	expectedWidth := 800.0 * 0.2
	expectedHeight := 1000.0 * 0.2

	if layoutTree.Dimensions.Content.Width != expectedWidth {
		t.Errorf("vmin birimi doğru hesaplanmalı, beklenen: %f, alınan: %f", expectedWidth, layoutTree.Dimensions.Content.Width)
	}

	if layoutTree.Dimensions.Content.Height != expectedHeight {
		t.Errorf("vmax birimi doğru hesaplanmalı, beklenen: %f, alınan: %f", expectedHeight, layoutTree.Dimensions.Content.Height)
	}
}

// TestWhitespaceOnlyTextNodesInLayout - Tests whether whitespace-only text nodes create boxes in layout
func TestWhitespaceOnlyTextNodesInLayout(t *testing.T) {
	// Multi-line HTML - contains whitespace-only text nodes
	htmlSource := `<div class="container">
		</div>`

	htmlParser := html.NewParser(htmlSource)
	htmlRoot := htmlParser.Parse()

	// Count whitespace-only text nodes
	whitespaceOnlyCount := 0
	var countWhitespaceNodes func(*html.Node)
	countWhitespaceNodes = func(n *html.Node) {
		if n.Type == html.TextNode && strings.TrimSpace(n.Content) == "" && n.Content != "" {
			whitespaceOnlyCount++
			t.Logf("Found whitespace-only text node: '%s' (len=%d)",
				strings.ReplaceAll(strings.ReplaceAll(n.Content, "\n", "\\n"), "\t", "\\t"), len(n.Content))
		}
		for _, child := range n.Children {
			countWhitespaceNodes(child)
		}
	}
	countWhitespaceNodes(htmlRoot)

	if whitespaceOnlyCount == 0 {
		t.Log("No whitespace-only text nodes found in HTML")
		return
	}

	t.Logf("Found %d whitespace-only text nodes in HTML", whitespaceOnlyCount)

	// Create CSS and styled tree
	cssSource := `
		.container {
			display: block;
			width: 800px;
			margin: 50px auto;
			padding: 20px;
			background-color: white;
			border: 2px solid #333;
		}
	`
	cssParser := css.NewParser(cssSource)
	stylesheet := cssParser.Parse()

	styledTree := style.BuildStyledTree(htmlRoot, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	// Calculate layout
	ctx := NewLayoutContext(1200, 800)
	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 1200, Height: 800},
	}
	layoutTree.LayoutWithContext(containingBlock, ctx)

	// Find boxes created from whitespace-only text nodes in layout tree
	whitespaceBoxCount := 0
	var countWhitespaceBoxes func(*LayoutBox)
	countWhitespaceBoxes = func(box *LayoutBox) {
		if box.StyledNode != nil && box.StyledNode.Node != nil {
			if box.StyledNode.Node.Type == html.TextNode {
				content := box.StyledNode.Node.Content
				if strings.TrimSpace(content) == "" && content != "" {
					whitespaceBoxCount++
					t.Logf("Found whitespace-only text box: '%s' (len=%d), Dimensions: W=%.2f H=%.2f",
						strings.ReplaceAll(strings.ReplaceAll(content, "\n", "\\n"), "\t", "\\t"),
						len(content),
						box.Dimensions.Content.Width,
						box.Dimensions.Content.Height)

					// If box width or height is greater than 0, this may cause black rectangle to appear
					if box.Dimensions.Content.Width > 0 || box.Dimensions.Content.Height > 0 {
						t.Logf("  ⚠️  This box is renderable (W=%.2f, H=%.2f) - may cause black rectangle to appear!",
							box.Dimensions.Content.Width, box.Dimensions.Content.Height)
					}
				}
			}
		}
		for _, child := range box.Children {
			countWhitespaceBoxes(child)
		}
	}
	countWhitespaceBoxes(layoutTree)

	t.Logf("Found %d whitespace-only text boxes in layout tree", whitespaceBoxCount)

	if whitespaceBoxCount > 0 {
		t.Errorf("BUG DETECTED: %d whitespace-only text boxes were created in layout. These boxes cause black rectangles to appear when rendered.", whitespaceBoxCount)
	}
}
