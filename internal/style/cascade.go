package style

import (
	"goglweb/internal/parser/html"
	"strings"
)

func matches(node *html.Node, selector string) bool {
	if selector == "*" {
		return true
	}
	if node.TagName == selector {
		return true
	}
	if strings.HasPrefix(selector, ".") {
		classValue := node.Attr["class"]
		if classValue == "" {
			return false
		}

		classes := strings.Fields(classValue)
		targetClass := selector[1:]
		for _, cls := range classes {
			if cls == targetClass {
				return true
			}
		}

		return false
	}
	if strings.HasPrefix(selector, "#") {
		idValue := node.Attr["id"]
		if idValue == "" {
			return false
		}
		return idValue == selector[1:]
	}
	return false
}
