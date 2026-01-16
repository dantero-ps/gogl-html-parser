package css

import (
	"testing"
)

func TestTokenizer(t *testing.T) {
	input := "h1 { color: red; }"
	tokenizer := NewTokenizer(input)

	expected := []TokenType{
		TokenIdent, TokenWhitespace, TokenLeftBrace, TokenWhitespace,
		TokenIdent, TokenColon, TokenWhitespace, TokenIdent, TokenSemicolon,
		TokenWhitespace, TokenRightBrace,
	}

	for _, expType := range expected {
		tok := tokenizer.NextToken()
		if tok.Type != expType {
			t.Errorf("Beklenen tip %v, gelen %v (Değer: %q)", expType, tok.Type, tok.Value)
		}
	}
}
