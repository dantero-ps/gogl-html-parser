package style

import (
	"github.com/furkandgn/goglweb/internal/parser/html"
	"slices"
	"strings"
)

func matches(node *html.Node, selector string) bool {
	if selector == "*" {
		return true
	}

	// Pure tag name (no qualifiers).
	if !strings.ContainsAny(selector, ".#:") {
		return node.TagName == selector
	}

	// Compound selector: optional tag + any number of .class, #id, :pseudo qualifiers.
	// e.g. ".foo", "#bar", "p.lead", "div.foo.bar", "input#submit", ".btn:hover"
	firstQual := strings.IndexAny(selector, ".#:")
	if tag := selector[:firstQual]; tag != "" && node.TagName != tag {
		return false
	}
	qualPart := selector[firstQual:]

	for len(qualPart) > 0 {
		kind := qualPart[0]
		end := strings.IndexAny(qualPart[1:], ".#:")
		var value string
		if end == -1 {
			value = qualPart[1:]
			qualPart = ""
		} else {
			value = qualPart[1 : end+1]
			qualPart = qualPart[end+1:]
		}
		switch kind {
		case '.':
			if !slices.Contains(strings.Fields(node.Attr["class"]), value) {
				return false
			}
		case '#':
			if node.Attr["id"] != value {
				return false
			}
		case ':':
			switch value {
			case "hover":
				if !node.Hovered {
					return false
				}
			case "active":
				if !node.Active {
					return false
				}
			}
		}
	}
	return true
}
