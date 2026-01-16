package css

type TokenType int

const (
	TokenError TokenType = iota
	TokenEOF
	TokenIdent
	TokenWhitespace
	TokenLeftBrace
	TokenRightBrace
	TokenColon
	TokenSemicolon
)

type Token struct {
	Type  TokenType
	Value string
}

type Tokenizer struct {
	input string
	pos   int
}

func NewTokenizer(input string) *Tokenizer {
	return &Tokenizer{input: input}
}

func (t *Tokenizer) NextToken() Token {
	if t.pos >= len(t.input) {
		return Token{Type: TokenEOF}
	}

	char := t.input[t.pos]

	// Return whitespace token
	if char == ' ' || char == '\n' || char == '\t' || char == '\r' {
		start := t.pos
		for t.pos < len(t.input) && (t.input[t.pos] == ' ' || t.input[t.pos] == '\n' || t.input[t.pos] == '\t' || t.input[t.pos] == '\r') {
			t.pos++
		}
		return Token{Type: TokenWhitespace, Value: t.input[start:t.pos]}
	}

	switch char {
	case '{':
		t.pos++
		return Token{Type: TokenLeftBrace, Value: "{"}
	case '}':
		t.pos++
		return Token{Type: TokenRightBrace, Value: "}"}
	case ':':
		t.pos++
		return Token{Type: TokenColon, Value: ":"}
	case ';':
		t.pos++
		return Token{Type: TokenSemicolon, Value: ";"}
	}

	// Hex color (#rrggbb or #rgb)
	if char == '#' {
		start := t.pos
		t.pos++
		// Read hex characters
		for t.pos < len(t.input) && ((t.input[t.pos] >= '0' && t.input[t.pos] <= '9') ||
			(t.input[t.pos] >= 'a' && t.input[t.pos] <= 'f') ||
			(t.input[t.pos] >= 'A' && t.input[t.pos] <= 'F')) {
			t.pos++
		}
		return Token{Type: TokenIdent, Value: t.input[start:t.pos]}
	}

	// Number (like 200px, 3px, 10px)
	if (char >= '0' && char <= '9') || char == '.' || char == '+' || char == '-' {
		start := t.pos
		// Sign
		if char == '+' || char == '-' {
			t.pos++
		}
		// Number part
		for t.pos < len(t.input) && ((t.input[t.pos] >= '0' && t.input[t.pos] <= '9') || t.input[t.pos] == '.') {
			t.pos++
		}
		// Unit part (like px, em, rem, %)
		for t.pos < len(t.input) && isAlpha(t.input[t.pos]) {
			t.pos++
		}
		return Token{Type: TokenIdent, Value: t.input[start:t.pos]}
	}

	// Ident (letters, numbers, -, _)
	start := t.pos
	for t.pos < len(t.input) && isIdentChar(t.input[t.pos]) {
		t.pos++
	}
	if start < t.pos {
		return Token{Type: TokenIdent, Value: t.input[start:t.pos]}
	}

	// Unknown character
	t.pos++
	return Token{Type: TokenError, Value: string(char)}
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-'
}

func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.'
}
