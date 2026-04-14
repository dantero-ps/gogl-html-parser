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

	// Pure tag name (no . or # qualifiers)
	if !strings.ContainsAny(selector, ".#") {
		return node.TagName == selector
	}

	// Pure class selector: ".foo"
	if strings.HasPrefix(selector, ".") {
		classValue := node.Attr["class"]
		if classValue == "" {
			return false
		}
		classes := strings.Fields(classValue)
		targetClass := selector[1:]
		return slices.Contains(classes, targetClass)
	}

	// Pure id selector: "#foo"
	if strings.HasPrefix(selector, "#") {
		idValue := node.Attr["id"]
		if idValue == "" {
			return false
		}
		return idValue == selector[1:]
	}

	// Compound selector: tag + one or more .class and/or #id qualifiers
	// e.g. "p.lead", "div.foo.bar", "input#submit"
	// Extract tag name (everything before the first . or #)
	firstQual := strings.IndexAny(selector, ".#")
	tagPart := selector[:firstQual]
	qualPart := selector[firstQual:]

	if tagPart != "" && node.TagName != tagPart {
		return false
	}

	// Walk through each qualifier (.class or #id)
	for len(qualPart) > 0 {
		if qualPart[0] == '.' {
			// Find end of this class name
			end := strings.IndexAny(qualPart[1:], ".#")
			var className string
			if end == -1 {
				className = qualPart[1:]
				qualPart = ""
			} else {
				className = qualPart[1 : end+1]
				qualPart = qualPart[end+1:]
			}
			classValue := node.Attr["class"]
			if classValue == "" {
				return false
			}
			if !slices.Contains(strings.Fields(classValue), className) {
				return false
			}
		} else if qualPart[0] == '#' {
			end := strings.IndexAny(qualPart[1:], ".#")
			var idName string
			if end == -1 {
				idName = qualPart[1:]
				qualPart = ""
			} else {
				idName = qualPart[1 : end+1]
				qualPart = qualPart[end+1:]
			}
			if node.Attr["id"] != idName {
				return false
			}
		} else {
			break
		}
	}
	return true
}
