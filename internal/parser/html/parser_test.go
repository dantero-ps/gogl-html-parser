package html

import (
	"strings"
	"testing"
)

func TestNestingDepthLimit(t *testing.T) {
	// Create HTML with 300 nested divs — should not panic
	var sb strings.Builder
	for range 300 {
		sb.WriteString("<div>")
	}
	sb.WriteString("deep")
	for range 300 {
		sb.WriteString("</div>")
	}

	parser := NewParser(sb.String())
	root := parser.Parse()

	if root == nil {
		t.Fatal("Parser should return a valid tree, not nil")
	}

	// Walk to the deepest node
	depth := 0
	node := root
	for len(node.Children) > 0 {
		found := false
		for _, child := range node.Children {
			if child.Type == ElementNode {
				node = child
				depth++
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	if depth > maxNestingDepth {
		t.Errorf("Nesting depth %d exceeds limit %d", depth, maxNestingDepth)
	}

	// Verify the deep text "deep" exists somewhere in the tree
	if !containsText(root, "deep") {
		t.Error("Deep text content should still be in the tree")
	}
}

func TestNormalNestingWorks(t *testing.T) {
	// 10 levels of nesting should work fine
	html := `<div><div><div><div><div><div><div><div><div><div>hello</div></div></div></div></div></div></div></div></div></div>`
	parser := NewParser(html)
	root := parser.Parse()

	depth := 0
	node := root
	for len(node.Children) > 0 {
		found := false
		for _, child := range node.Children {
			if child.Type == ElementNode {
				node = child
				depth++
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	// root is div 1, then we walk 9 children to reach div 10 — depth = 9
	if depth != 9 {
		t.Errorf("Expected depth 9 (10 divs, root + 9 children walked), got %d", depth)
	}
}

func containsText(node *Node, text string) bool {
	if node.Type == TextNode && strings.Contains(node.Content, text) {
		return true
	}
	for _, child := range node.Children {
		if containsText(child, text) {
			return true
		}
	}
	return false
}

func TestFullParser(t *testing.T) {
	input := `<section id="main"><img src="logo.png"><p>Hello <b>World</b></p><br /></section>`
	p := NewParser(input)
	root := p.Parse()

	if root.TagName != "section" || root.Attr["id"] != "main" {
		t.Error("Root section is incorrect")
	}

	// Child count check: img, p, br (Total 3)
	if len(root.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(root.Children))
	}

	// img check (should be sibling of p since it's self-closing)
	img := root.Children[0]
	if img.TagName != "img" || img.Attr["src"] != "logo.png" {
		t.Error("img tag is incorrect")
	}

	// Nested check: section > p > b
	pTag := root.Children[1]
	if pTag.TagName != "p" || pTag.Children[1].TagName != "b" {
		t.Error("Nested tag structure is broken")
	}
}

// TestWhitespaceOnlyTextNodes - Tests whether whitespace-only text nodes are parsed
func TestWhitespaceOnlyTextNodes(t *testing.T) {
	// Single line - no whitespace
	input1 := `<div class="container"></div>`
	p1 := NewParser(input1)
	root1 := p1.Parse()

	// Single line should not have whitespace-only text nodes
	hasWhitespaceOnlyText := false
	for _, child := range root1.Children {
		if child.Type == TextNode && strings.TrimSpace(child.Content) == "" && child.Content != "" {
			hasWhitespaceOnlyText = true
		}
	}
	if hasWhitespaceOnlyText {
		t.Error("Single line HTML should not have whitespace-only text nodes")
	}

	// Multi-line - has whitespace
	input2 := `<div class="container">
		</div>`
	p2 := NewParser(input2)
	root2 := p2.Parse()

	// Multi-line may have whitespace-only text nodes (this was the bug source)
	whitespaceOnlyTextNodes := []*Node{}
	for _, child := range root2.Children {
		if child.Type == TextNode && strings.TrimSpace(child.Content) == "" && child.Content != "" {
			whitespaceOnlyTextNodes = append(whitespaceOnlyTextNodes, child)
			t.Logf("Found whitespace-only text node in multi-line: '%s' (len=%d)",
				strings.ReplaceAll(strings.ReplaceAll(child.Content, "\n", "\\n"), "\t", "\\t"), len(child.Content))
		}
	}

	if len(whitespaceOnlyTextNodes) == 0 {
		t.Log("No whitespace-only text nodes found in multi-line HTML (this is expected behavior)")
	} else {
		t.Logf("Found %d whitespace-only text nodes in multi-line HTML (this may be the bug source)",
			len(whitespaceOnlyTextNodes))
	}
}

// TestWhitespaceOnlyTextNodesWithContent - Whitespace test with div containing content
func TestWhitespaceOnlyTextNodesWithContent(t *testing.T) {
	input := `<div class="container">
		<h1>GoGL Web Browser</h1>
		<p>This is a demo page.</p>
	</div>`

	p := NewParser(input)
	root := p.Parse()

	// Find all text nodes
	textNodes := []*Node{}
	var collectTextNodes func(*Node)
	collectTextNodes = func(n *Node) {
		if n.Type == TextNode {
			textNodes = append(textNodes, n)
		}
		for _, child := range n.Children {
			collectTextNodes(child)
		}
	}
	collectTextNodes(root)

	whitespaceOnlyCount := 0
	for _, tn := range textNodes {
		if strings.TrimSpace(tn.Content) == "" && tn.Content != "" {
			whitespaceOnlyCount++
			t.Logf("Whitespace-only text node: '%s' (len=%d)",
				strings.ReplaceAll(strings.ReplaceAll(tn.Content, "\n", "\\n"), "\t", "\\t"), len(tn.Content))
		}
	}

	t.Logf("Total text node count: %d, Whitespace-only: %d", len(textNodes), whitespaceOnlyCount)

	if whitespaceOnlyCount > 0 {
		t.Logf("BUG DETECTED: %d whitespace-only text nodes were parsed. These nodes create boxes in layout that may appear as black rectangles.", whitespaceOnlyCount)
	}
}
