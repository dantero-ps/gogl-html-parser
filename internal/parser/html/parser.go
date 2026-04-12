package html

import "strings"

// maxNestingDepth limits how deeply HTML elements can be nested.
// Prevents stack overflow from malformed or adversarial input.
const maxNestingDepth = 256

type Parser struct {
	tokenizer *Tokenizer
}

func NewParser(input string) *Parser {
	return &Parser{tokenizer: NewTokenizer(input)}
}

func (p *Parser) Parse() *Node {
	var stack []*Node
	var root *Node

	for {
		tok := p.tokenizer.Next()
		if tok.Type == ErrorToken {
			break
		}

		switch tok.Type {
		case StartTag:
			node := NewElement(tok.Data)
			node.Attr = tok.Attr

			if root == nil {
				root = node
			}

			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, node)
			}

			if !tok.IsSelfClosing {
				if len(stack) >= maxNestingDepth {
					// Don't push — element stays at max depth, preventing deeper nesting
					continue
				}
				stack = append(stack, node)
			}

		case Text:
			// Filter whitespace-only text nodes (containing only spaces, tabs, newlines)
			// These nodes create unnecessary boxes in layout that appear as black rectangles
			trimmed := strings.TrimSpace(tok.Data)
			if trimmed == "" && tok.Data != "" {
				// Skip whitespace-only text node
				continue
			}

			if len(stack) > 0 {
				node := NewText(tok.Data)
				stack[len(stack)-1].Children = append(stack[len(stack)-1].Children, node)
			}

		case EndTag:
			if len(stack) > 0 && stack[len(stack)-1].TagName == tok.Data {
				stack = stack[:len(stack)-1]
			}
		}
	}

	return root
}
