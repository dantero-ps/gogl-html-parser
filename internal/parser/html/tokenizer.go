package html

import (
	gohtml "html"
	"strings"
	"unicode"
)

type TokenType int

const (
	ErrorToken TokenType = iota
	StartTag
	EndTag
	Text
)

var voidElements = map[string]bool{
	"img": true, "br": true, "hr": true, "input": true, "meta": true, "link": true,
}

type Token struct {
	Type          TokenType
	Data          string
	Attr          map[string]string
	IsSelfClosing bool
}

type Tokenizer struct {
	input string
	pos   int
}

func NewTokenizer(input string) *Tokenizer {
	return &Tokenizer{input: input}
}

func (t *Tokenizer) Next() Token {
	if t.pos >= len(t.input) {
		return Token{Type: ErrorToken}
	}

	if t.input[t.pos] == '<' {
		return t.parseTag()
	}

	return t.parseText()
}

func (t *Tokenizer) parseTag() Token {
	t.pos++ // Skip '<' character
	if t.pos < len(t.input) && t.input[t.pos] == '/' {
		// End Tag: </div>
		t.pos++
		end := strings.Index(t.input[t.pos:], ">")
		tagName := t.input[t.pos : t.pos+end]
		t.pos += end + 1
		return Token{Type: EndTag, Data: strings.TrimSpace(tagName)}
	}

	// Start Tag or Self-closing: <div class="foo"> or <img />
	end := strings.Index(t.input[t.pos:], ">")
	raw := t.input[t.pos : t.pos+end]
	t.pos += end + 1

	isSelfClosing := strings.HasSuffix(raw, "/")
	raw = strings.TrimSuffix(raw, "/")

	parts := parseAttributes(raw)
	tagName := parts[0]
	attrs := make(map[string]string)

	for _, attrStr := range parts[1:] {
		if strings.Contains(attrStr, "=") {
			kv := strings.SplitN(attrStr, "=", 2)
			key := kv[0]
			val := strings.Trim(kv[1], `"'`)
			attrs[key] = val
		}
	}

	return Token{
		Type:          StartTag,
		Data:          tagName,
		Attr:          attrs,
		IsSelfClosing: isSelfClosing || voidElements[tagName],
	}
}

func (t *Tokenizer) parseText() Token {
	end := strings.Index(t.input[t.pos:], "<")
	if end == -1 {
		end = len(t.input) - t.pos
	}
	content := t.input[t.pos : t.pos+end]
	t.pos += end
	return Token{Type: Text, Data: gohtml.UnescapeString(content)}
}

func parseAttributes(s string) []string {
	var result []string
	var builder strings.Builder
	inQuotes := false

	for _, r := range s {
		if r == '"' || r == '\'' {
			inQuotes = !inQuotes
			builder.WriteRune(r)
		} else if unicode.IsSpace(r) && !inQuotes {
			if builder.Len() > 0 {
				result = append(result, builder.String())
				builder.Reset()
			}
		} else {
			builder.WriteRune(r)
		}
	}
	if builder.Len() > 0 {
		result = append(result, builder.String())
	}
	return result
}
