package css

import "testing"

func TestParser(t *testing.T) {
	input := "h1 { color: red; }"
	p := NewParser(input)
	stylesheet := p.Parse()

	if len(stylesheet.Rules) != 1 {
		t.Fatalf("1 kural bekleniyordu, %d bulundu", len(stylesheet.Rules))
	}

	rule := stylesheet.Rules[0]
	if rule.Selector != "h1" {
		t.Errorf("Beklenen selector 'h1', gelen '%s'", rule.Selector)
	}

	if rule.Declarations[0].Property != "color" || rule.Declarations[0].Value != "red" {
		t.Errorf("Declaration hatalı: %+v", rule.Declarations[0])
	}
}
