package layout

import (
	"goglweb/internal/parser/css"
	"goglweb/internal/parser/html"
	"goglweb/internal/style"
	"math"
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

// TestNestedEmResolution tests that em resolves against parent font-size, not root.
// Parent has font-size: 1.5em (= 24px with root 16), child has font-size: 2em (= 48px).
func TestNestedEmResolution(t *testing.T) {
	parentNode := &html.Node{
		TagName: "div",
		Type:    html.ElementNode,
		Children: []*html.Node{
			{TagName: "span", Type: html.ElementNode},
		},
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "div", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "font-size", Value: "1.5em"},
				{Property: "width", Value: "800px"},
			}},
			{Selector: "span", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "font-size", Value: "2em"},
				{Property: "width", Value: "10em"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(parentNode, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	ctx := NewLayoutContext(1000, 800)
	ctx.RootFontSize = 16.0

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 1000, Height: 800},
	}
	layoutTree.LayoutWithContext(containingBlock, ctx)

	// Child should be 10em * (16 * 1.5 * 2) = 10 * 48 = 480px
	if len(layoutTree.Children) == 0 {
		t.Fatal("Expected child box")
	}
	childBox := layoutTree.Children[0]
	expectedWidth := 10.0 * 16.0 * 1.5 * 2.0 // = 480
	if math.Abs(childBox.Dimensions.Content.Width-expectedWidth) > 0.01 {
		t.Errorf("Nested em resolution: expected %.2f, got %.2f", expectedWidth, childBox.Dimensions.Content.Width)
	}
}

// TestPercentageHeight tests that height: 50% resolves against containing block height.
func TestPercentageHeight(t *testing.T) {
	root := &html.Node{
		TagName: "div",
		Type:    html.ElementNode,
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "div", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "height", Value: "50%"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 200},
	}
	layoutTree.Layout(containingBlock)

	expectedHeight := 100.0 // 50% of 200
	if math.Abs(layoutTree.Dimensions.Content.Height-expectedHeight) > 0.01 {
		t.Errorf("Percentage height: expected %.2f, got %.2f", expectedHeight, layoutTree.Dimensions.Content.Height)
	}
}

// TestTextMetricsInterface verifies FallbackMeasurer produces expected widths.
func TestTextMetricsInterface(t *testing.T) {
	m := &FallbackMeasurer{}
	metrics := m.MeasureText("Hello", "", 16.0)
	expected := float64(5) * 16.0 * 0.6
	if math.Abs(metrics.Width-expected) > 0.01 {
		t.Errorf("FallbackMeasurer: expected width %.2f, got %.2f", expected, metrics.Width)
	}
	if metrics.Height != 16.0 {
		t.Errorf("FallbackMeasurer: expected height 16, got %.2f", metrics.Height)
	}
}

// TestMultiLineWrappingHeight verifies that multi-line text boxes accumulate height across all lines.
func TestMultiLineWrappingHeight(t *testing.T) {
	m := &FallbackMeasurer{}
	// Each word is ~4 chars * 16 * 0.6 = 38.4px. With maxWidth=50, "The" fits alone, etc.
	lines := WordWrap("The quick brown fox", "", 16.0, 50.0, m)
	if len(lines) <= 1 {
		t.Errorf("Expected multiple lines for narrow container, got %d", len(lines))
	}
}

// --- Plan 02-01: Positioned Layout Tests ---

// TestPositionedRelativeOffset verifies that position:relative; top:20px; left:10px shifts the box
// without affecting sibling positions.
func TestPositionedRelativeOffset(t *testing.T) {
	// Manually construct styled tree for precise control
	parentNode := &html.Node{TagName: "div", Type: html.ElementNode}
	child1Node := &html.Node{TagName: "div", Type: html.ElementNode}
	child2Node := &html.Node{TagName: "div", Type: html.ElementNode}
	parentNode.Children = []*html.Node{child1Node, child2Node}

	parentStyled := &style.StyledNode{
		Node: parentNode,
		SpecifiedValues: map[string]string{
			"display": "block",
			"width":   "800px",
		},
	}
	child1Styled := &style.StyledNode{
		Node: child1Node,
		SpecifiedValues: map[string]string{
			"display":  "block",
			"width":    "100px",
			"height":   "50px",
			"position": "relative",
			"top":      "20px",
			"left":     "10px",
		},
	}
	child2Styled := &style.StyledNode{
		Node: child2Node,
		SpecifiedValues: map[string]string{
			"display": "block",
			"width":   "100px",
			"height":  "30px",
		},
	}
	parentStyled.Children = []*style.StyledNode{child1Styled, child2Styled}

	layoutTree := BuildLayoutTree(parentStyled)
	ctx := NewLayoutContext(800, 600)
	containingBlock := Dimensions{Content: Rect{X: 0, Y: 0, Width: 800, Height: 600}}
	layoutTree.LayoutWithContext(containingBlock, ctx)

	if len(layoutTree.Children) < 2 {
		t.Fatalf("Expected at least 2 children, got %d", len(layoutTree.Children))
	}

	child1 := layoutTree.Children[0]
	child2 := layoutTree.Children[1]

	// child1 should have position:relative with PositionType set
	if child1.PositionType != "relative" {
		t.Errorf("First child PositionType should be 'relative', got '%s'", child1.PositionType)
	}

	// child2 should be static
	if child2.PositionType != "static" {
		t.Errorf("Second child PositionType should be 'static', got '%s'", child2.PositionType)
	}

	// child1 should be shifted by (10, 20) from its normal position
	// Normal position would be at Y=0, X=0 (first child in parent)
	// After relative offset: X=10, Y=20
	if child1.Dimensions.Content.X != 10 {
		t.Errorf("First child X should be 10 (left:10px offset), got %.2f", child1.Dimensions.Content.X)
	}
	if child1.Dimensions.Content.Y != 20 {
		t.Errorf("First child Y should be 20 (top:20px offset), got %.2f", child1.Dimensions.Content.Y)
	}

	// child2 should be at the normal flow position (Y = 0 + 50px height = 50)
	// It should NOT be displaced by the relative offset of child1
	normalFlowY := child1.Dimensions.Content.Y - 20 + child1.Dimensions.Content.Height // normal Y + height
	if math.Abs(child2.Dimensions.Content.Y-normalFlowY) > 0.01 {
		t.Errorf("Second child Y should be at normal flow position %.2f, got %.2f (sibling should be unaffected by relative offset)", normalFlowY, child2.Dimensions.Content.Y)
	}
}

// TestPositionedAbsoluteBox verifies that an absolute child is positioned relative to its
// nearest positioned ancestor, not the viewport.
func TestPositionedAbsoluteBox(t *testing.T) {
	// Build: outer block (800px) → inner block (position:relative, 200x100) → child (position:absolute, top:5px, left:5px, 50x50)
	outerNode := &html.Node{TagName: "div", Type: html.ElementNode}
	innerNode := &html.Node{TagName: "div", Type: html.ElementNode}
	absNode := &html.Node{TagName: "div", Type: html.ElementNode}
	outerNode.Children = []*html.Node{innerNode}
	innerNode.Children = []*html.Node{absNode}

	outerStyled := &style.StyledNode{
		Node: outerNode,
		SpecifiedValues: map[string]string{
			"display": "block",
			"width":   "800px",
		},
	}
	innerStyled := &style.StyledNode{
		Node: innerNode,
		SpecifiedValues: map[string]string{
			"display":  "block",
			"width":    "200px",
			"height":   "100px",
			"position": "relative",
		},
	}
	absStyled := &style.StyledNode{
		Node: absNode,
		SpecifiedValues: map[string]string{
			"display":  "block",
			"width":    "50px",
			"height":   "50px",
			"position": "absolute",
			"top":      "5px",
			"left":     "5px",
		},
	}
	outerStyled.Children = []*style.StyledNode{innerStyled}
	innerStyled.Children = []*style.StyledNode{absStyled}

	layoutTree := BuildLayoutTree(outerStyled)
	ctx := NewLayoutContext(800, 600)
	containingBlock := Dimensions{Content: Rect{X: 0, Y: 0, Width: 800, Height: 600}}
	layoutTree.LayoutWithContext(containingBlock, ctx)

	if len(layoutTree.Children) < 1 {
		t.Fatal("Expected at least 1 child (inner block)")
	}
	inner := layoutTree.Children[0]
	innerContentX := inner.Dimensions.Content.X
	innerContentY := inner.Dimensions.Content.Y

	// Inner should have a deferred absolute child
	if len(inner.DeferredAbsolute) < 1 {
		t.Fatalf("Inner should have 1 deferred absolute child, got %d", len(inner.DeferredAbsolute))
	}
	absChild := inner.DeferredAbsolute[0]

	// Absolute child should be positioned at (innerContentX+5, innerContentY+5)
	expectedX := innerContentX + 5
	expectedY := innerContentY + 5
	if math.Abs(absChild.Dimensions.Content.X-expectedX) > 0.01 {
		t.Errorf("Absolute child X should be %.2f (inner X + 5), got %.2f", expectedX, absChild.Dimensions.Content.X)
	}
	if math.Abs(absChild.Dimensions.Content.Y-expectedY) > 0.01 {
		t.Errorf("Absolute child Y should be %.2f (inner Y + 5), got %.2f", expectedY, absChild.Dimensions.Content.Y)
	}

	// Inner block height should still be 100 (absolute child does not expand it)
	if math.Abs(inner.Dimensions.Content.Height-100) > 0.01 {
		t.Errorf("Inner block height should be 100, got %.2f", inner.Dimensions.Content.Height)
	}
}

// TestPositionedFixedBox verifies that a fixed element is positioned relative to the viewport.
func TestPositionedFixedBox(t *testing.T) {
	// Build: outer block (800x600) → inner block (position:relative, 200x100, margin-left:100, margin-top:50) → fixed child (top:0, right:0, 100x40)
	outerNode := &html.Node{TagName: "div", Type: html.ElementNode}
	innerNode := &html.Node{TagName: "div", Type: html.ElementNode}
	fixedNode := &html.Node{TagName: "div", Type: html.ElementNode}
	outerNode.Children = []*html.Node{innerNode}
	innerNode.Children = []*html.Node{fixedNode}

	outerStyled := &style.StyledNode{
		Node: outerNode,
		SpecifiedValues: map[string]string{
			"display": "block",
			"width":   "800px",
		},
	}
	innerStyled := &style.StyledNode{
		Node: innerNode,
		SpecifiedValues: map[string]string{
			"display":     "block",
			"width":       "200px",
			"height":      "100px",
			"position":    "relative",
			"margin-left": "100px",
			"margin-top":  "50px",
		},
	}
	fixedStyled := &style.StyledNode{
		Node: fixedNode,
		SpecifiedValues: map[string]string{
			"display":  "block",
			"width":    "100px",
			"height":   "40px",
			"position": "fixed",
			"top":      "0px",
			"right":    "0px",
		},
	}
	outerStyled.Children = []*style.StyledNode{innerStyled}
	innerStyled.Children = []*style.StyledNode{fixedStyled}

	layoutTree := BuildLayoutTree(outerStyled)
	ctx := NewLayoutContext(800, 600)
	containingBlock := Dimensions{Content: Rect{X: 0, Y: 0, Width: 800, Height: 600}}
	layoutTree.LayoutWithContext(containingBlock, ctx)

	inner := layoutTree.Children[0]

	// Inner should have the fixed child in DeferredAbsolute
	if len(inner.DeferredAbsolute) < 1 {
		t.Fatalf("Inner should have 1 deferred fixed child, got %d", len(inner.DeferredAbsolute))
	}
	fixedChild := inner.DeferredAbsolute[0]

	// Fixed child should be at viewport top-right: X = 800 - 100 = 700, Y = 0
	if math.Abs(fixedChild.Dimensions.Content.X-700) > 0.01 {
		t.Errorf("Fixed child X should be 700 (viewport right - width), got %.2f", fixedChild.Dimensions.Content.X)
	}
	if math.Abs(fixedChild.Dimensions.Content.Y-0) > 0.01 {
		t.Errorf("Fixed child Y should be 0 (viewport top), got %.2f", fixedChild.Dimensions.Content.Y)
	}
}

// --- Plan 02-02: Flexbox Layout Tests ---

// TestFlexRowDistribution verifies that three 100px children in a 300px flex row
// are distributed left to right at X positions 0, 100, 200.
func TestFlexRowDistribution(t *testing.T) {
	containerNode := &html.Node{TagName: "div", Type: html.ElementNode}
	childNodes := make([]*html.Node, 3)
	for i := range 3 {
		childNodes[i] = &html.Node{TagName: "div", Type: html.ElementNode}
	}
	containerNode.Children = []*html.Node{childNodes[0], childNodes[1], childNodes[2]}

	childStyles := make([]*style.StyledNode, 3)
	for i := range 3 {
		childStyles[i] = &style.StyledNode{
			Node: childNodes[i],
			SpecifiedValues: map[string]string{
				"display": "block",
				"width":   "100px",
				"height":  "50px",
			},
		}
	}

	containerStyled := &style.StyledNode{
		Node: containerNode,
		SpecifiedValues: map[string]string{
			"display": "flex",
			"width":   "300px",
		},
		Children: childStyles,
	}

	layoutTree := BuildLayoutTree(containerStyled)
	ctx := NewLayoutContext(800, 600)
	containingBlock := Dimensions{Content: Rect{X: 0, Y: 0, Width: 800, Height: 600}}
	layoutTree.LayoutWithContext(containingBlock, ctx)

	if layoutTree.BoxType != FlexBox {
		t.Fatalf("Container should be FlexBox, got %v", layoutTree.BoxType)
	}
	if len(layoutTree.Children) != 3 {
		t.Fatalf("Expected 3 children, got %d", len(layoutTree.Children))
	}

	containerX := layoutTree.Dimensions.Content.X
	for i, child := range layoutTree.Children {
		expectedX := containerX + float64(i)*100
		if math.Abs(child.Dimensions.Content.X-expectedX) > 0.01 {
			t.Errorf("Child %d X should be %.2f, got %.2f", i, expectedX, child.Dimensions.Content.X)
		}
	}

	// Container height should be 50 (tallest child)
	if math.Abs(layoutTree.Dimensions.Content.Height-50) > 0.01 {
		t.Errorf("Container height should be 50, got %.2f", layoutTree.Dimensions.Content.Height)
	}
}

// TestFlexJustifyContentSpaceBetween verifies space-between distribution.
func TestFlexJustifyContentSpaceBetween(t *testing.T) {
	containerNode := &html.Node{TagName: "div", Type: html.ElementNode}
	child1Node := &html.Node{TagName: "div", Type: html.ElementNode}
	child2Node := &html.Node{TagName: "div", Type: html.ElementNode}
	containerNode.Children = []*html.Node{child1Node, child2Node}

	containerStyled := &style.StyledNode{
		Node: containerNode,
		SpecifiedValues: map[string]string{
			"display":         "flex",
			"width":           "300px",
			"justify-content": "space-between",
		},
		Children: []*style.StyledNode{
			{Node: child1Node, SpecifiedValues: map[string]string{
				"display": "block", "width": "50px", "height": "40px",
			}},
			{Node: child2Node, SpecifiedValues: map[string]string{
				"display": "block", "width": "50px", "height": "40px",
			}},
		},
	}

	layoutTree := BuildLayoutTree(containerStyled)
	ctx := NewLayoutContext(800, 600)
	containingBlock := Dimensions{Content: Rect{X: 0, Y: 0, Width: 800, Height: 600}}
	layoutTree.LayoutWithContext(containingBlock, ctx)

	containerX := layoutTree.Dimensions.Content.X
	child1 := layoutTree.Children[0]
	child2 := layoutTree.Children[1]

	// First child at start (X = containerX)
	if math.Abs(child1.Dimensions.Content.X-containerX) > 0.01 {
		t.Errorf("Child 1 X should be %.2f, got %.2f", containerX, child1.Dimensions.Content.X)
	}

	// Second child at X = containerX + 250 (300 - 50)
	expectedX2 := containerX + 250
	if math.Abs(child2.Dimensions.Content.X-expectedX2) > 0.01 {
		t.Errorf("Child 2 X should be %.2f, got %.2f", expectedX2, child2.Dimensions.Content.X)
	}
}

// TestFlexAlignItemsCenter verifies vertical centering in a row flex container.
func TestFlexAlignItemsCenter(t *testing.T) {
	containerNode := &html.Node{TagName: "div", Type: html.ElementNode}
	childANode := &html.Node{TagName: "div", Type: html.ElementNode}
	childBNode := &html.Node{TagName: "div", Type: html.ElementNode}
	containerNode.Children = []*html.Node{childANode, childBNode}

	containerStyled := &style.StyledNode{
		Node: containerNode,
		SpecifiedValues: map[string]string{
			"display":     "flex",
			"width":       "300px",
			"align-items": "center",
		},
		Children: []*style.StyledNode{
			{Node: childANode, SpecifiedValues: map[string]string{
				"display": "block", "width": "50px", "height": "100px",
			}},
			{Node: childBNode, SpecifiedValues: map[string]string{
				"display": "block", "width": "50px", "height": "40px",
			}},
		},
	}

	layoutTree := BuildLayoutTree(containerStyled)
	ctx := NewLayoutContext(800, 600)
	containingBlock := Dimensions{Content: Rect{X: 0, Y: 0, Width: 800, Height: 600}}
	layoutTree.LayoutWithContext(containingBlock, ctx)

	containerY := layoutTree.Dimensions.Content.Y
	childA := layoutTree.Children[0]
	childB := layoutTree.Children[1]

	// Item A (100px tall) should be at container top (tallest)
	if math.Abs(childA.Dimensions.Content.Y-containerY) > 0.01 {
		t.Errorf("Item A Y should be %.2f (container top), got %.2f", containerY, childA.Dimensions.Content.Y)
	}

	// Item B (40px tall) should be centered: Y = containerY + (100-40)/2 = containerY + 30
	expectedY := containerY + 30
	if math.Abs(childB.Dimensions.Content.Y-expectedY) > 0.01 {
		t.Errorf("Item B Y should be %.2f (centered), got %.2f", expectedY, childB.Dimensions.Content.Y)
	}
}

// TestFlexGrowDistribution verifies that flex-grow distributes remaining space.
func TestFlexGrowDistribution(t *testing.T) {
	containerNode := &html.Node{TagName: "div", Type: html.ElementNode}
	childANode := &html.Node{TagName: "div", Type: html.ElementNode}
	childBNode := &html.Node{TagName: "div", Type: html.ElementNode}
	containerNode.Children = []*html.Node{childANode, childBNode}

	containerStyled := &style.StyledNode{
		Node: containerNode,
		SpecifiedValues: map[string]string{
			"display": "flex",
			"width":   "300px",
		},
		Children: []*style.StyledNode{
			{Node: childANode, SpecifiedValues: map[string]string{
				"display":    "block",
				"flex-basis": "50px",
				"flex-grow":  "1",
				"height":     "40px",
			}},
			{Node: childBNode, SpecifiedValues: map[string]string{
				"display":    "block",
				"flex-basis": "50px",
				"flex-grow":  "2",
				"height":     "40px",
			}},
		},
	}

	layoutTree := BuildLayoutTree(containerStyled)
	ctx := NewLayoutContext(800, 600)
	containingBlock := Dimensions{Content: Rect{X: 0, Y: 0, Width: 800, Height: 600}}
	layoutTree.LayoutWithContext(containingBlock, ctx)

	childA := layoutTree.Children[0]
	childB := layoutTree.Children[1]

	// Free space = 300 - 50 - 50 = 200
	// A gets 200 * (1/3) ≈ 66.67 → total ≈ 116.67
	// B gets 200 * (2/3) ≈ 133.33 → total ≈ 183.33
	expectedA := 50.0 + 200.0*(1.0/3.0)
	expectedB := 50.0 + 200.0*(2.0/3.0)
	if math.Abs(childA.Dimensions.Content.Width-expectedA) > 0.5 {
		t.Errorf("Child A width should be ~%.2f, got %.2f", expectedA, childA.Dimensions.Content.Width)
	}
	if math.Abs(childB.Dimensions.Content.Width-expectedB) > 0.5 {
		t.Errorf("Child B width should be ~%.2f, got %.2f", expectedB, childB.Dimensions.Content.Width)
	}
}

// --- Plan 02-03: Overflow and Scroll Tests ---

// TestClampScrollOffset verifies the scroll offset clamping helper.
func TestClampScrollOffset(t *testing.T) {
	// Negative offset clamped to 0
	if ClampScrollOffset(-10, 400, 200) != 0 {
		t.Errorf("ClampScrollOffset(-10, 400, 200) should be 0, got %.2f", ClampScrollOffset(-10, 400, 200))
	}

	// Offset exceeding max (400-200=200) clamped to 200
	if ClampScrollOffset(300, 400, 200) != 200 {
		t.Errorf("ClampScrollOffset(300, 400, 200) should be 200, got %.2f", ClampScrollOffset(300, 400, 200))
	}

	// Valid offset unchanged
	if ClampScrollOffset(100, 400, 200) != 100 {
		t.Errorf("ClampScrollOffset(100, 400, 200) should be 100, got %.2f", ClampScrollOffset(100, 400, 200))
	}

	// Content smaller than visible: max clamped to 0
	if ClampScrollOffset(50, 100, 200) != 0 {
		t.Errorf("ClampScrollOffset(50, 100, 200) should be 0, got %.2f", ClampScrollOffset(50, 100, 200))
	}
}

// TestIsScrollable verifies the IsScrollable helper on LayoutBox.
func TestIsScrollable(t *testing.T) {
	box := &LayoutBox{}
	if box.IsScrollable() {
		t.Error("Empty overflow should not be scrollable")
	}

	box.Overflow = "visible"
	if box.IsScrollable() {
		t.Error("overflow:visible should not be scrollable")
	}

	box.Overflow = "hidden"
	if box.IsScrollable() {
		t.Error("overflow:hidden should not be scrollable")
	}

	box.Overflow = "scroll"
	if !box.IsScrollable() {
		t.Error("overflow:scroll should be scrollable")
	}

	box.Overflow = "auto"
	if !box.IsScrollable() {
		t.Error("overflow:auto should be scrollable")
	}
}
