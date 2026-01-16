package html

import "testing"

func TestTokenizer(t *testing.T) {
	input := `<div class="container"><img src="a.jpg" /> <br></div>`
	tokenizer := NewTokenizer(input)

	tests := []struct {
		expectedType TokenType
		expectedData string
		isSelf       bool
	}{
		{StartTag, "div", false},
		{StartTag, "img", true},
		{Text, " ", false},
		{StartTag, "br", true},
		{EndTag, "div", false},
	}

	for _, tc := range tests {
		tok := tokenizer.Next()
		if tok.Type != tc.expectedType || tok.Data != tc.expectedData || tok.IsSelfClosing != tc.isSelf {
			t.Errorf("Hata! Beklenen: %v %s (Self:%v), Gelen: %v %s (Self:%v)",
				tc.expectedType, tc.expectedData, tc.isSelf, tok.Type, tok.Data, tok.IsSelfClosing)
		}
	}
}
