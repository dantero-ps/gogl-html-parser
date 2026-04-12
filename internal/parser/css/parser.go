package css

import "strings"

type Parser struct {
	tokenizer *Tokenizer
	curToken  Token
}

func NewParser(input string) *Parser {
	p := &Parser{tokenizer: NewTokenizer(input)}
	p.nextToken() // Load first token
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.tokenizer.NextToken()
	// Skip whitespace (usually insignificant in CSS)
	if p.curToken.Type == TokenWhitespace {
		p.nextToken()
	}
}

func (p *Parser) Parse() *Stylesheet {
	s := &Stylesheet{}
	for p.curToken.Type != TokenEOF {
		if p.curToken.Type == TokenIdent {
			s.Rules = append(s.Rules, p.parseRule())
		} else {
			p.nextToken()
		}
	}
	return s
}

func (p *Parser) parseRule() Rule {
	rule := Rule{Selector: p.curToken.Value}
	p.nextToken() // { expected

	if p.curToken.Type == TokenLeftBrace {
		p.nextToken()
		for p.curToken.Type != TokenRightBrace && p.curToken.Type != TokenEOF {
			rule.Declarations = append(rule.Declarations, p.parseDeclaration())
		}
	}
	return rule
}

func (p *Parser) parseDeclaration() Declaration {
	prop := p.curToken.Value
	p.nextToken() // : expected

	// Read value part (can be multiple tokens: "3px solid #000")
	var valueParts []string
	p.nextToken() // First value token

	// Collect all values until semicolon or right brace
	for p.curToken.Type != TokenSemicolon && p.curToken.Type != TokenRightBrace && p.curToken.Type != TokenEOF {
		if p.curToken.Type == TokenIdent {
			valueParts = append(valueParts, p.curToken.Value)
		}
		p.nextToken()
	}

	// Combine values
	var val strings.Builder
	for i, part := range valueParts {
		if i > 0 {
			val.WriteString(" ")
		}
		val.WriteString(part)
	}

	// Skip semicolon if present
	if p.curToken.Type == TokenSemicolon {
		p.nextToken()
	}

	return Declaration{Property: prop, Value: val.String()}
}
