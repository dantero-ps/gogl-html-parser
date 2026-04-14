package layout

import (
	"fmt"
	"github.com/furkandgn/goglweb/internal/parser/css"
	"github.com/furkandgn/goglweb/internal/parser/html"
	"github.com/furkandgn/goglweb/internal/style"
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
	metrics := m.MeasureText("Hello", "", 16.0, "", "")
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
	lines := WordWrap("The quick brown fox", "", 16.0, 50.0, m, "", "")
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

// --- Margin Positioning Tests ---

func TestMarginTopNotDoubled(t *testing.T) {
	// Single child with margin-top: 20px inside a parent.
	// The child's Content.Y should be parent_content_Y + marginTop (once), not twice.
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
			}},
			{Selector: "p", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "height", Value: "50px"},
				{Property: "margin-top", Value: "20px"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}

	layoutTree.Layout(containingBlock)

	parentContentY := layoutTree.Dimensions.Content.Y
	child := layoutTree.Children[0]

	expectedY := parentContentY + 20
	if child.Dimensions.Content.Y != expectedY {
		t.Errorf("Child Content.Y should be %f (parent content Y + marginTop once), got %f",
			expectedY, child.Dimensions.Content.Y)
	}
}

func TestSiblingMarginCollapse(t *testing.T) {
	// Two siblings: first has margin-bottom: 30px, second has margin-top: 20px.
	// Collapsed margin = max(30, 20) = 30px.
	// Gap between content bottom of first and content top of second should be 30px.
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
				{Property: "margin-bottom", Value: "30px"},
				{Property: "margin-top", Value: "20px"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}

	layoutTree.Layout(containingBlock)

	child0 := layoutTree.Children[0]
	child1 := layoutTree.Children[1]

	child0Bottom := child0.Dimensions.Content.Y + child0.Dimensions.Content.Height
	gap := child1.Dimensions.Content.Y - child0Bottom

	// Gap should be collapsed margin (30px), not 30+20=50px
	// and not 30+20+20=70px (doubled marginTop)
	expectedCollapsedMargin := 30.0
	if math.Abs(gap-expectedCollapsedMargin) > 0.01 {
		t.Errorf("Gap between siblings should be %f (collapsed margin), got %f. "+
			"child0 bottom=%f, child1 top=%f",
			expectedCollapsedMargin, gap, child0Bottom, child1.Dimensions.Content.Y)
	}
}

func TestSiblingMarginCollapseWithPadding(t *testing.T) {
	// Two siblings with margin + padding + border
	// First: margin-bottom=10, padding-bottom=5, border-bottom=2
	// Second: margin-top=20, padding-top=5, border-top=2
	// Collapsed margin = max(10, 20) = 20
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
				{Property: "margin-bottom", Value: "10px"},
				{Property: "margin-top", Value: "20px"},
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

	child0 := layoutTree.Children[0]
	child1 := layoutTree.Children[1]

	// Bottom of child0's border box
	child0BorderBottom := child0.Dimensions.Content.Y + child0.Dimensions.Content.Height + child0.Dimensions.Padding.Bottom + child0.Dimensions.Border.Bottom
	// Top of child1's border box
	child1BorderTop := child1.Dimensions.Content.Y - child1.Dimensions.Padding.Top - child1.Dimensions.Border.Top

	marginGap := child1BorderTop - child0BorderBottom

	// Gap between border boxes should be collapsed margin = max(10, 20) = 20
	expectedCollapsedMargin := 20.0
	if math.Abs(marginGap-expectedCollapsedMargin) > 0.01 {
		t.Errorf("Margin gap between border boxes should be %f, got %f. "+
			"child0 borderBottom=%f, child1 borderTop=%f",
			expectedCollapsedMargin, marginGap, child0BorderBottom, child1BorderTop)
	}
}

func TestNoMarginCollapseFirstChild(t *testing.T) {
	// First child with margin-top: 40px should be positioned 40px below parent content top.
	// No collapsing happens for the first child (no previous sibling).
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
				{Property: "padding-top", Value: "10px"},
			}},
			{Selector: "p", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "height", Value: "50px"},
				{Property: "margin-top", Value: "40px"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}

	layoutTree.Layout(containingBlock)

	parentContentY := layoutTree.Dimensions.Content.Y
	child := layoutTree.Children[0]

	// Parent content Y + child's marginTop (once) = child's content Y
	expectedY := parentContentY + 40.0
	if child.Dimensions.Content.Y != expectedY {
		t.Errorf("Child Content.Y should be %f (parent content Y + marginTop once), got %f",
			expectedY, child.Dimensions.Content.Y)
	}
}

// --- Inline Span Layout Tests ---

// mockMeasurer simulates a real font measurer where "Hello World" fills ~80% of a 200px line.
// This tests behavior that FallbackMeasurer cannot reproduce.
type mockMeasurer struct {
	charWidth float64
	fontSize  float64
	lineH     float64
}

func (m *mockMeasurer) MeasureText(text string, fontFamily string, fontSize float64, fontWeight string, fontStyle string) TextMetrics {
	return TextMetrics{
		Width:      float64(len(text)) * m.charWidth * (fontSize / m.fontSize),
		Height:     m.lineH * (fontSize / m.fontSize),
		Ascent:     m.lineH * 0.8 * (fontSize / m.fontSize),
		LineHeight: m.lineH * 1.2 * (fontSize / m.fontSize),
	}
}
func (m *mockMeasurer) MeasureWord(word string, fontFamily string, fontSize float64, fontWeight string, fontStyle string) float64 {
	return m.MeasureText(word, fontFamily, fontSize, fontWeight, fontStyle).Width
}

func TestInlineSpanWidthMatchesTextContent(t *testing.T) {
	// <p><span>Hello</span></p>
	// The span's width should be proportional to "Hello" text length, NOT the container width.
	root := &html.Node{
		TagName: "p",
		Type:    html.ElementNode,
		Children: []*html.Node{
			{
				TagName:  "span",
				Type:     html.ElementNode,
				Children: []*html.Node{{Type: html.TextNode, Content: "Hello"}},
			},
		},
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "p", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "800px"},
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

	// Layout tree: p(BlockBox) > anonymous(AnonymousBox) > span(InlineBox)
	if len(layoutTree.Children) != 1 {
		t.Fatalf("Expected 1 child (anonymous box), got %d", len(layoutTree.Children))
	}
	anonBox := layoutTree.Children[0]
	if anonBox.BoxType != AnonymousBox {
		t.Fatalf("Expected AnonymousBox, got %v", anonBox.BoxType)
	}
	if len(anonBox.Children) != 1 {
		t.Fatalf("Expected 1 child in anonymous box, got %d", len(anonBox.Children))
	}
	spanBox := anonBox.Children[0]

	// With FallbackMeasurer: "Hello" = 5 chars * 16 * 0.6 = 48px
	// The span width should be ~48px, NOT close to 800px (container width)
	expectedWidth := 5.0 * 16.0 * 0.6
	if spanBox.Dimensions.Content.Width > expectedWidth*1.5 {
		t.Errorf("Span width should be ~%.1f (text measured width), got %.1f — "+
			"inline box is taking container width instead of text content width",
			expectedWidth, spanBox.Dimensions.Content.Width)
	}
	if spanBox.Dimensions.Content.Width == 0 {
		t.Errorf("Span width should not be 0")
	}
}

func TestTwoInlineSpansSideBySide(t *testing.T) {
	// <p><span>Hello</span><span>World</span></p>
	// Both spans should sit side by side, NOT on separate lines.
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
				Children: []*html.Node{{Type: html.TextNode, Content: "World"}},
			},
		},
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "p", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "800px"},
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

	anonBox := layoutTree.Children[0]
	if len(anonBox.Children) != 2 {
		t.Fatalf("Expected 2 spans in anonymous box, got %d", len(anonBox.Children))
	}

	span1 := anonBox.Children[0]
	span2 := anonBox.Children[1]

	// Both spans should be on the SAME line (same Y)
	if span1.Dimensions.Content.Y != span2.Dimensions.Content.Y {
		t.Errorf("Spans should be on the same line. span1.Y=%.1f, span2.Y=%.1f",
			span1.Dimensions.Content.Y, span2.Dimensions.Content.Y)
	}

	// span2 should be to the RIGHT of span1
	if span2.Dimensions.Content.X <= span1.Dimensions.Content.X {
		t.Errorf("span2 should be to the right of span1. span1.X=%.1f, span2.X=%.1f",
			span1.Dimensions.Content.X, span2.Dimensions.Content.X)
	}

	// span2.X should equal span1.X + span1.Width (they should be adjacent)
	expectedGap := span1.Dimensions.Content.X + span1.Dimensions.Content.Width
	gap := math.Abs(span2.Dimensions.Content.X - expectedGap)
	if gap > 1.0 {
		t.Errorf("span2.X (%.1f) should be adjacent to span1.X + span1.Width (%.1f), gap=%.1f",
			span2.Dimensions.Content.X, expectedGap, gap)
	}
}

func TestInlineSpanDoesNotOverflowToNewLine(t *testing.T) {
	// <p><span>Short</span><span>Text</span></p>
	// With a wide container, both spans should fit on one line.
	root := &html.Node{
		TagName: "p",
		Type:    html.ElementNode,
		Children: []*html.Node{
			{
				TagName:  "span",
				Type:     html.ElementNode,
				Children: []*html.Node{{Type: html.TextNode, Content: "Short"}},
			},
			{
				TagName:  "span",
				Type:     html.ElementNode,
				Children: []*html.Node{{Type: html.TextNode, Content: "Text"}},
			},
		},
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "p", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "800px"},
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

	anonBox := layoutTree.Children[0]
	span1 := anonBox.Children[0]
	span2 := anonBox.Children[1]

	// Total width of both spans should NOT exceed container width (800)
	// FallbackMeasurer: "Short" = 5*16*0.6=48, "Text" = 4*16*0.6=38.4 → total ~86.4
	totalWidth := span2.Dimensions.Content.X + span2.Dimensions.Content.Width - span1.Dimensions.Content.X
	if totalWidth > 800 {
		t.Errorf("Both spans should fit on one line. Total width=%.1f, container=800", totalWidth)
	}

	// Anonymous box height should be one line, not two
	// FallbackMeasurer: single line height = fontSize * 1.2 = 19.2
	if anonBox.Dimensions.Content.Height > 16.0*1.2*2.0 {
		t.Errorf("Anonymous box height should be ~1 line (%.1f), got %.1f — "+
			"spans may be wrapping to multiple lines",
			16.0*1.2, anonBox.Dimensions.Content.Height)
	}
}

func TestInlineSpanBetweenTextNodes(t *testing.T) {
	// <p>Hello <span>inline</span> world</p>
	// "Hello " (text) + span("inline") + " world" (text) should all flow inline.
	root := &html.Node{
		TagName: "p",
		Type:    html.ElementNode,
		Children: []*html.Node{
			{Type: html.TextNode, Content: "Hello "},
			{
				TagName:  "span",
				Type:     html.ElementNode,
				Children: []*html.Node{{Type: html.TextNode, Content: "inline"}},
			},
			{Type: html.TextNode, Content: " world"},
		},
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "p", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "800px"},
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

	anonBox := layoutTree.Children[0]

	// All children of the anonymous box should be on the same Y
	for i, child := range anonBox.Children {
		if i > 0 && child.Dimensions.Content.Y != anonBox.Children[0].Dimensions.Content.Y {
			t.Errorf("Child %d Y (%.1f) differs from child 0 Y (%.1f) — "+
				"inline elements not on the same line", i, child.Dimensions.Content.Y,
				anonBox.Children[0].Dimensions.Content.Y)
		}
	}

	// Check X ordering: each child should start after the previous one ends
	for i := 1; i < len(anonBox.Children); i++ {
		prev := anonBox.Children[i-1]
		curr := anonBox.Children[i]
		if curr.Dimensions.Content.X < prev.Dimensions.Content.X+prev.Dimensions.Content.Width {
			t.Errorf("Child %d X (%.1f) overlaps with child %d end (%.1f) — "+
				"inline elements not flowing left-to-right",
				i, curr.Dimensions.Content.X, i-1, prev.Dimensions.Content.X+prev.Dimensions.Content.Width)
		}
	}
}

func TestInlineSpanWidthNarrowNotContainer(t *testing.T) {
	// Regression: an inline span with short text should NOT have width ≈ container width.
	// This is the core bug: layoutInline gives spans the container width via WordWrap.
	root := &html.Node{
		TagName: "p",
		Type:    html.ElementNode,
		Children: []*html.Node{
			{Type: html.TextNode, Content: "Before "},
			{
				TagName:  "span",
				Type:     html.ElementNode,
				Children: []*html.Node{{Type: html.TextNode, Content: "X"}},
			},
			{Type: html.TextNode, Content: " After"},
		},
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "p", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "800px"},
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

	anonBox := layoutTree.Children[0]
	// Find the span box
	var spanBox *LayoutBox
	for _, child := range anonBox.Children {
		if child.BoxType == InlineBox && child.StyledNode != nil && child.StyledNode.Node.Type == html.ElementNode {
			spanBox = child
			break
		}
	}
	if spanBox == nil {
		t.Fatal("Could not find span InlineBox in anonymous box")
	}

	// "X" is 1 char: 1 * 16 * 0.6 = 9.6px
	// Span width should be ~9.6px, NOT 800px
	measuredWidth := 1.0 * 16.0 * 0.6
	if spanBox.Dimensions.Content.Width > measuredWidth*2.0 {
		t.Errorf("Span with single char 'X' should have width ~%.1f, got %.1f — "+
			"inline box width is inflated (likely taking container width from WordWrap)",
			measuredWidth, spanBox.Dimensions.Content.Width)
	}
}

// --- parseShorthand Tests ---

// TestInlineSpanWidthWithRealisticMeasurer uses a mock measurer with wider characters
// (like a real font) to detect if layoutInline inflates inline box width.
// With FallbackMeasurer the bug doesn't show because char widths are too narrow.
func TestInlineSpanWidthWithRealisticMeasurer(t *testing.T) {
	// <p><span>inline spans</span></p>
	// With a measurer that gives ~10px per char, "inline spans" = 12 chars * 10 = 120px.
	// If layoutInline gives it container width (~800px), that's a bug.
	root := &html.Node{
		TagName: "p",
		Type:    html.ElementNode,
		Children: []*html.Node{
			{
				TagName:  "span",
				Type:     html.ElementNode,
				Children: []*html.Node{{Type: html.TextNode, Content: "inline spans"}},
			},
		},
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "p", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "800px"},
			}},
			{Selector: "span", Declarations: []css.Declaration{
				{Property: "display", Value: "inline"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	// Mock measurer: 10px per character at fontSize 16
	measurer := &mockMeasurer{charWidth: 10.0, fontSize: 16.0, lineH: 16.0}
	ctx := NewLayoutContext(800, 600)
	ctx.TextMeasurer = measurer

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}

	layoutTree.LayoutWithContext(containingBlock, ctx)

	anonBox := layoutTree.Children[0]
	spanBox := anonBox.Children[0]

	// "inline spans" = 12 chars, with spaces = 12 (splitWords splits to ["inline", "spans"])
	// WordWrap produces one line: "inline spans" = 12 chars * 10px = 120px
	expectedTextWidth := 12.0 * 10.0
	if spanBox.Dimensions.Content.Width > expectedTextWidth*1.5 {
		t.Errorf("Span width should be ~%.1f (text measured width), got %.1f — "+
			"inline box is taking container width instead of text content width",
			expectedTextWidth, spanBox.Dimensions.Content.Width)
	}
	if spanBox.Dimensions.Content.Width == 0 {
		t.Fatal("Span width should not be 0")
	}
}

func TestTwoSpansAdjacentWithRealisticMeasurer(t *testing.T) {
	// <p><span>AAA</span><span>BBB</span></p>
	// With 10px/char: "AAA"=30px, "BBB"=30px. Total = 60px.
	// If span1 takes container width, span2 is pushed to a new line or far right.
	root := &html.Node{
		TagName: "p",
		Type:    html.ElementNode,
		Children: []*html.Node{
			{
				TagName:  "span",
				Type:     html.ElementNode,
				Children: []*html.Node{{Type: html.TextNode, Content: "AAA"}},
			},
			{
				TagName:  "span",
				Type:     html.ElementNode,
				Children: []*html.Node{{Type: html.TextNode, Content: "BBB"}},
			},
		},
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "p", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "800px"},
			}},
			{Selector: "span", Declarations: []css.Declaration{
				{Property: "display", Value: "inline"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	measurer := &mockMeasurer{charWidth: 10.0, fontSize: 16.0, lineH: 16.0}
	ctx := NewLayoutContext(800, 600)
	ctx.TextMeasurer = measurer

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}

	layoutTree.LayoutWithContext(containingBlock, ctx)

	anonBox := layoutTree.Children[0]
	span1 := anonBox.Children[0]
	span2 := anonBox.Children[1]

	// Both should be on the same line
	if span1.Dimensions.Content.Y != span2.Dimensions.Content.Y {
		t.Errorf("Spans should be on same line. span1.Y=%.1f, span2.Y=%.1f",
			span1.Dimensions.Content.Y, span2.Dimensions.Content.Y)
	}

	// span1 width should be ~30px (3 chars * 10px), NOT ~800px
	span1ExpectedWidth := 3.0 * 10.0
	if span1.Dimensions.Content.Width > span1ExpectedWidth*1.5 {
		t.Errorf("span1 width should be ~%.1f, got %.1f — inline box inflated to container width",
			span1ExpectedWidth, span1.Dimensions.Content.Width)
	}

	// span2 should start right after span1
	expectedSpan2X := span1.Dimensions.Content.X + span1.Dimensions.Content.Width
	if math.Abs(span2.Dimensions.Content.X-expectedSpan2X) > 1.0 {
		t.Errorf("span2.X (%.1f) should be adjacent to span1 end (%.1f)",
			span2.Dimensions.Content.X, expectedSpan2X)
	}
}

func TestLongInlineTextWrapsAndDoesNotInflate(t *testing.T) {
	// <p><span>This is a longer piece of text that should wrap</span></p>
	// Container = 200px, charWidth=10px → ~20 chars per line.
	// The span should wrap but its width should be ~200px (container), not more.
	// And crucially, the text AFTER the span should start at the correct position.
	root := &html.Node{
		TagName: "p",
		Type:    html.ElementNode,
		Children: []*html.Node{
			{
				TagName:  "span",
				Type:     html.ElementNode,
				Children: []*html.Node{{Type: html.TextNode, Content: "This is a longer piece of text that should wrap"}},
			},
		},
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "p", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "200px"},
			}},
			{Selector: "span", Declarations: []css.Declaration{
				{Property: "display", Value: "inline"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	measurer := &mockMeasurer{charWidth: 10.0, fontSize: 16.0, lineH: 16.0}
	ctx := NewLayoutContext(200, 600)
	ctx.TextMeasurer = measurer

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 200, Height: 600},
	}

	layoutTree.LayoutWithContext(containingBlock, ctx)

	anonBox := layoutTree.Children[0]
	spanBox := anonBox.Children[0]

	// Span should not exceed container width
	if spanBox.Dimensions.Content.Width > 200+1 {
		t.Errorf("Span width (%.1f) should not exceed container width (200)", spanBox.Dimensions.Content.Width)
	}

	// Span should be multi-line: "This is a longer piece of text that should wrap"
	// = 50 chars, at 10px/char and 200px/line → at least 2 lines
	// Height should be > single line height (16*1.2=19.2)
	singleLineHeight := 16.0 * 1.2
	if spanBox.Dimensions.Content.Height <= singleLineHeight {
		t.Errorf("Span height (%.1f) should be > single line height (%.1f) for wrapped text",
			spanBox.Dimensions.Content.Height, singleLineHeight)
	}
}

func TestInlineSpanBetweenTextWithRealisticMeasurer(t *testing.T) {
	// <p>Before <span>MID</span> After</p>
	// With 10px/char: "Before "=70px, "MID"=30px, " After"=60px. Total ~160px in 800px container.
	// All should be on same line, flowing left-to-right.
	root := &html.Node{
		TagName: "p",
		Type:    html.ElementNode,
		Children: []*html.Node{
			{Type: html.TextNode, Content: "Before "},
			{
				TagName:  "span",
				Type:     html.ElementNode,
				Children: []*html.Node{{Type: html.TextNode, Content: "MID"}},
			},
			{Type: html.TextNode, Content: " After"},
		},
	}

	stylesheet := &css.Stylesheet{
		Rules: []css.Rule{
			{Selector: "p", Declarations: []css.Declaration{
				{Property: "display", Value: "block"},
				{Property: "width", Value: "800px"},
			}},
			{Selector: "span", Declarations: []css.Declaration{
				{Property: "display", Value: "inline"},
			}},
		},
	}

	styledTree := style.BuildStyledTree(root, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	measurer := &mockMeasurer{charWidth: 10.0, fontSize: 16.0, lineH: 16.0}
	ctx := NewLayoutContext(800, 600)
	ctx.TextMeasurer = measurer

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}

	layoutTree.LayoutWithContext(containingBlock, ctx)

	anonBox := layoutTree.Children[0]
	if len(anonBox.Children) != 3 {
		t.Fatalf("Expected 3 children in anonymous box, got %d", len(anonBox.Children))
	}

	textBefore := anonBox.Children[0]
	spanBox := anonBox.Children[1]
	textAfter := anonBox.Children[2]

	// All on same line
	firstY := textBefore.Dimensions.Content.Y
	for i, child := range anonBox.Children {
		if child.Dimensions.Content.Y != firstY {
			t.Errorf("Child %d Y (%.1f) differs from first child Y (%.1f)", i, child.Dimensions.Content.Y, firstY)
		}
	}

	// Span width should be ~30px (3 chars * 10px)
	spanExpectedWidth := 3.0 * 10.0
	if spanBox.Dimensions.Content.Width > spanExpectedWidth*1.5 {
		t.Errorf("Span width should be ~%.1f, got %.1f", spanExpectedWidth, spanBox.Dimensions.Content.Width)
	}

	// Span should start after "Before " (~70px)
	textBeforeEnd := textBefore.Dimensions.Content.X + textBefore.Dimensions.Content.Width
	if math.Abs(spanBox.Dimensions.Content.X-textBeforeEnd) > 1.0 {
		t.Errorf("Span X (%.1f) should be adjacent to textBefore end (%.1f)",
			spanBox.Dimensions.Content.X, textBeforeEnd)
	}

	// textAfter should start after span
	spanEnd := spanBox.Dimensions.Content.X + spanBox.Dimensions.Content.Width
	if math.Abs(textAfter.Dimensions.Content.X-spanEnd) > 1.0 {
		t.Errorf("textAfter X (%.1f) should be adjacent to span end (%.1f)",
			textAfter.Dimensions.Content.X, spanEnd)
	}
}

// --- parseShorthand Tests ---

func TestParseShorthandEmpty(t *testing.T) {
	result := parseShorthand("", 0)
	if result != "" {
		t.Errorf("parseShorthand('', 0) = %q, want empty", result)
	}
}

func TestParseShorthandSingleValue(t *testing.T) {
	for _, idx := range []int{0, 1, 2, 3} {
		result := parseShorthand("10px", idx)
		if result != "10px" {
			t.Errorf("parseShorthand('10px', %d) = %q, want '10px'", idx, result)
		}
	}
}

func TestParseShorthandTwoValues(t *testing.T) {
	// CSS: [top/bottom] [left/right] → "20px 40px"
	// index 0=left → 40px, 1=right → 40px, 2=top → 20px, 3=bottom → 20px
	cases := []struct {
		index    int
		expected string
	}{
		{0, "40px"},
		{1, "40px"},
		{2, "20px"},
		{3, "20px"},
	}
	for _, tc := range cases {
		result := parseShorthand("20px 40px", tc.index)
		if result != tc.expected {
			t.Errorf("parseShorthand('20px 40px', %d) = %q, want %q", tc.index, result, tc.expected)
		}
	}
}

func TestParseShorthandThreeValues(t *testing.T) {
	// CSS: [top] [left/right] [bottom] → "10px 20px 30px"
	// index 0=left → 20px, 1=right → 20px, 2=top → 10px, 3=bottom → 30px
	cases := []struct {
		index    int
		expected string
	}{
		{0, "20px"},
		{1, "20px"},
		{2, "10px"},
		{3, "30px"},
	}
	for _, tc := range cases {
		result := parseShorthand("10px 20px 30px", tc.index)
		if result != tc.expected {
			t.Errorf("parseShorthand('10px 20px 30px', %d) = %q, want %q", tc.index, result, tc.expected)
		}
	}
}

func TestParseShorthandFourValues(t *testing.T) {
	// CSS: [top] [right] [bottom] [left] → "1px 2px 3px 4px"
	// index 0=left → 4px, 1=right → 2px, 2=top → 1px, 3=bottom → 3px
	cases := []struct {
		index    int
		expected string
	}{
		{0, "4px"},
		{1, "2px"},
		{2, "1px"},
		{3, "3px"},
	}
	for _, tc := range cases {
		result := parseShorthand("1px 2px 3px 4px", tc.index)
		if result != tc.expected {
			t.Errorf("parseShorthand('1px 2px 3px 4px', %d) = %q, want %q", tc.index, result, tc.expected)
		}
	}
}

func TestParseShorthandFourValuesMargin(t *testing.T) {
	// Real example: margin: 20px 0 10px 0
	// CSS: top=20px right=0 bottom=10px left=0
	// index 0=left → 0, 1=right → 0, 2=top → 20px, 3=bottom → 10px
	cases := []struct {
		index    int
		expected string
	}{
		{0, "0"},
		{1, "0"},
		{2, "20px"},
		{3, "10px"},
	}
	for _, tc := range cases {
		result := parseShorthand("20px 0 10px 0", tc.index)
		if result != tc.expected {
			t.Errorf("parseShorthand('20px 0 10px 0', %d) = %q, want %q", tc.index, result, tc.expected)
		}
	}
}

func TestParseShorthandFourValuesAllSame(t *testing.T) {
	// margin: 10px 10px 10px 10px → every side should be 10px
	for _, idx := range []int{0, 1, 2, 3} {
		result := parseShorthand("10px 10px 10px 10px", idx)
		if result != "10px" {
			t.Errorf("parseShorthand('10px 10px 10px 10px', %d) = %q, want '10px'", idx, result)
		}
	}
}

func TestParseShorthandMixedUnits(t *testing.T) {
	// padding: 1em 2em 3em 4em — should still split and return correctly
	cases := []struct {
		index    int
		expected string
	}{
		{0, "4em"},
		{1, "2em"},
		{2, "1em"},
		{3, "3em"},
	}
	for _, tc := range cases {
		result := parseShorthand("1em 2em 3em 4em", tc.index)
		if result != tc.expected {
			t.Errorf("parseShorthand('1em 2em 3em 4em', %d) = %q, want %q", tc.index, result, tc.expected)
		}
	}
}

func TestParseShorthandTwoValuesZero(t *testing.T) {
	// margin: 0 auto → top/bottom=0, left/right=auto
	cases := []struct {
		index    int
		expected string
	}{
		{0, "auto"},
		{1, "auto"},
		{2, "0"},
		{3, "0"},
	}
	for _, tc := range cases {
		result := parseShorthand("0 auto", tc.index)
		if result != tc.expected {
			t.Errorf("parseShorthand('0 auto', %d) = %q, want %q", tc.index, result, tc.expected)
		}
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

// --- Debug: Real HTML inline spans layout ---

func printTree(box *LayoutBox, depth int) {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}
	boxType := ""
	switch box.BoxType {
	case BlockBox:
		boxType = "Block"
	case InlineBox:
		boxType = "Inline"
	case AnonymousBox:
		boxType = "Anonymous"
	}
	tag := ""
	content := ""
	if box.StyledNode != nil {
		tag = box.StyledNode.Node.TagName
		if box.StyledNode.Node.Content != "" {
			content = fmt.Sprintf(" content=%q", box.StyledNode.Node.Content)
		}
	}
	fmt.Printf("%s[%s] tag=%s X=%.1f Y=%.1f W=%.1f H=%.1f%s\n",
		indent, boxType, tag,
		box.Dimensions.Content.X, box.Dimensions.Content.Y,
		box.Dimensions.Content.Width, box.Dimensions.Content.Height,
		content)
	for _, child := range box.Children {
		printTree(child, depth+1)
	}
}

func TestDebugInlineSpansLayout(t *testing.T) {
	htmlStr := `<p class="lead">This paragraph uses a larger font size and bold weight to demonstrate <span class="highlight">inline spans</span> inside block elements.</p>`

	cssStr := `
* { margin: 0; padding: 0; }
p.lead {
  display: block;
  font-size: 16px;
  font-weight: bold;
  color: #222222;
  padding: 10px;
  background-color: #f0f4ff;
  border: 1px solid #b0c4f0;
}
.highlight {
  color: #c0392b;
  font-style: italic;
}
`

	doc := html.NewParser(htmlStr).Parse()

	stylesheet := css.NewParser(cssStr).Parse()
	styledTree := style.BuildStyledTree(doc, stylesheet)
	layoutTree := BuildLayoutTree(styledTree)

	containingBlock := Dimensions{
		Content: Rect{X: 0, Y: 0, Width: 800, Height: 600},
	}

	layoutTree.Layout(containingBlock)

	t.Log("=== Layout Tree ===")
	printTree(layoutTree, 0)

	// Verify structure
	if len(layoutTree.Children) != 1 {
		t.Fatalf("Expected 1 child (anonymous box), got %d", len(layoutTree.Children))
	}
	anonBox := layoutTree.Children[0]
	if anonBox.BoxType != AnonymousBox {
		t.Fatalf("Expected AnonymousBox, got %v", anonBox.BoxType)
	}

	// There should be 3 inline children: text("This...demonstrate "), span("inline spans"), text(" inside...")
	t.Logf("Anonymous box has %d children", len(anonBox.Children))
	for i, child := range anonBox.Children {
		tag := ""
		cont := ""
		if child.StyledNode != nil {
			tag = child.StyledNode.Node.TagName
			cont = child.StyledNode.Node.Content
		}
		bt := "Inline"
		if child.BoxType == AnonymousBox {
			bt = "Anonymous"
		}
		t.Logf("  child[%d]: %s tag=%s X=%.1f Y=%.1f W=%.1f H=%.1f content=%q",
			i, bt, tag,
			child.Dimensions.Content.X, child.Dimensions.Content.Y,
			child.Dimensions.Content.Width, child.Dimensions.Content.Height,
			cont)
	}

	// First two children (text + span "inline spans") should be on the same line.
	// The third text node (" inside block elements.") may wrap to a new line depending
	// on container width, which is correct behavior.
	if len(anonBox.Children) >= 2 {
		if anonBox.Children[1].Dimensions.Content.Y != anonBox.Children[0].Dimensions.Content.Y {
			t.Errorf("span Y (%.1f) should be on same line as first text Y (%.1f)",
				anonBox.Children[1].Dimensions.Content.Y, anonBox.Children[0].Dimensions.Content.Y)
		}
	}

	// Span should NOT have container width
	for _, child := range anonBox.Children {
		if child.StyledNode != nil && child.StyledNode.Node.TagName == "span" {
			if child.Dimensions.Content.Width > 800*0.5 {
				t.Errorf("Span width (%.1f) is suspiciously large — may be taking container width",
					child.Dimensions.Content.Width)
			}
		}
	}
}

// TestMultiLineInlineSpanFlow verifies that when a text node wraps to multiple lines,
// the next sibling (span) continues on the LAST line of the previous text rather than
// starting on a new line below it.
func TestMultiLineInlineSpanFlow(t *testing.T) {
	// Build: <p>AAA BBB CCC DDD <span>EEE</span> FFF</p>
	// charWidth=10, container=120px → "AAA BBB CCC DDD " wraps to:
	//   Line 1: "AAA BBB CCC"  (110px)
	//   Line 2: "DDD"          (30px)
	// LastLineWidth=30, NumLines=2
	// Span "EEE" (30px) should start at X=30 on line 2 (Y = firstText.Y + lineHeight)
	// " FFF" should follow span on the same line.

	root := &html.Node{
		TagName: "p",
		Type:    html.ElementNode,
		Children: []*html.Node{
			{Type: html.TextNode, Content: "AAA BBB CCC DDD "},
			{
				TagName: "span",
				Type:    html.ElementNode,
				Children: []*html.Node{
					{Type: html.TextNode, Content: "EEE"},
				},
			},
			{Type: html.TextNode, Content: " FFF"},
		},
	}

	styledRoot := &style.StyledNode{
		Node:            root,
		SpecifiedValues: style.PropertyMap{"display": "block"},
		Children: []*style.StyledNode{
			{
				Node:            root.Children[0],
				SpecifiedValues: style.PropertyMap{},
			},
			{
				Node:            root.Children[1],
				SpecifiedValues: style.PropertyMap{},
				Children: []*style.StyledNode{
					{
						Node:            root.Children[1].Children[0],
						SpecifiedValues: style.PropertyMap{},
					},
				},
			},
			{
				Node:            root.Children[2],
				SpecifiedValues: style.PropertyMap{},
			},
		},
	}

	layoutTree := BuildLayoutTree(styledRoot)
	measurer := &mockMeasurer{charWidth: 10.0, fontSize: 16.0, lineH: 16.0}
	ctx := NewLayoutContext(120, 600)
	ctx.TextMeasurer = measurer
	containingBlock := Dimensions{Content: Rect{X: 0, Y: 0, Width: 120, Height: 600}}
	layoutTree.LayoutWithContext(containingBlock, ctx)

	// Log the full tree for debugging.
	t.Logf("root box: X=%.1f Y=%.1f W=%.1f H=%.1f",
		layoutTree.Dimensions.Content.X,
		layoutTree.Dimensions.Content.Y,
		layoutTree.Dimensions.Content.Width,
		layoutTree.Dimensions.Content.Height)

	if len(layoutTree.Children) == 0 {
		t.Fatal("expected at least one child (anonymous box) under <p>")
	}

	anonBox := layoutTree.Children[0]
	t.Logf("anonBox: X=%.1f Y=%.1f W=%.1f H=%.1f children=%d",
		anonBox.Dimensions.Content.X,
		anonBox.Dimensions.Content.Y,
		anonBox.Dimensions.Content.Width,
		anonBox.Dimensions.Content.Height,
		len(anonBox.Children))

	if len(anonBox.Children) < 3 {
		t.Fatalf("expected 3 inline children in anonymous box, got %d", len(anonBox.Children))
	}

	firstText := anonBox.Children[0]
	spanBox := anonBox.Children[1]
	lastText := anonBox.Children[2]

	t.Logf("firstText: X=%.1f Y=%.1f W=%.1f H=%.1f NumLines=%d LastLineWidth=%.1f",
		firstText.Dimensions.Content.X,
		firstText.Dimensions.Content.Y,
		firstText.Dimensions.Content.Width,
		firstText.Dimensions.Content.Height,
		firstText.NumLines,
		firstText.LastLineWidth)
	t.Logf("span:      X=%.1f Y=%.1f W=%.1f H=%.1f",
		spanBox.Dimensions.Content.X,
		spanBox.Dimensions.Content.Y,
		spanBox.Dimensions.Content.Width,
		spanBox.Dimensions.Content.Height)
	t.Logf("lastText:  X=%.1f Y=%.1f W=%.1f H=%.1f",
		lastText.Dimensions.Content.X,
		lastText.Dimensions.Content.Y,
		lastText.Dimensions.Content.Width,
		lastText.Dimensions.Content.Height)

	// First text must wrap to exactly 2 lines.
	if firstText.NumLines != 2 {
		t.Errorf("firstText.NumLines = %d, want 2", firstText.NumLines)
	}

	// LastLineWidth of "DDD" = 3 chars * 10 = 30px.
	if math.Abs(firstText.LastLineWidth-30.0) > 0.1 {
		t.Errorf("firstText.LastLineWidth = %.2f, want 30.0", firstText.LastLineWidth)
	}

	// lineHeight = lineH * 1.2 = 16 * 1.2 = 19.2
	lineHeight := 16.0 * 1.2
	expectedSpanY := firstText.Dimensions.Content.Y + lineHeight
	if math.Abs(spanBox.Dimensions.Content.Y-expectedSpanY) > 0.1 {
		t.Errorf("span Y = %.2f, want %.2f (firstText.Y + lineHeight)", spanBox.Dimensions.Content.Y, expectedSpanY)
	}

	// Span X must continue after the last line of firstText (X = container left + LastLineWidth).
	expectedSpanX := containingBlock.Content.X + firstText.LastLineWidth
	if math.Abs(spanBox.Dimensions.Content.X-expectedSpanX) > 0.1 {
		t.Errorf("span X = %.2f, want %.2f (container.X + lastLineWidth)", spanBox.Dimensions.Content.X, expectedSpanX)
	}

	// " FFF" must be on the same line as the span.
	if math.Abs(lastText.Dimensions.Content.Y-spanBox.Dimensions.Content.Y) > 0.1 {
		t.Errorf("lastText Y = %.2f, want %.2f (same line as span)", lastText.Dimensions.Content.Y, spanBox.Dimensions.Content.Y)
	}

	// " FFF" must start immediately after the span.
	expectedLastTextX := spanBox.Dimensions.Content.X + spanBox.Dimensions.Content.Width
	if math.Abs(lastText.Dimensions.Content.X-expectedLastTextX) > 0.1 {
		t.Errorf("lastText X = %.2f, want %.2f (span.X + span.Width)", lastText.Dimensions.Content.X, expectedLastTextX)
	}
}
