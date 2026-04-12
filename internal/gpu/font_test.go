package gpu

import "testing"

func TestSanitizeFontFamily(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"Arial"`, "Arial"},
		{`'Times New Roman'`, "Times New Roman"},
		{"../etc/passwd", "etcpasswd"},
		{"Arial; rm -rf /", "Arial rm -rf"},
		{"Arial\x00Helvetica", "ArialHelvetica"},
		{"Helvetica Neue", "Helvetica Neue"},
		{"", ""},
		{"Arial", "Arial"},
		{"../../secret", "secret"},
		{"/usr/share/fonts/arial", "usrsharefontsarial"},
	}

	for _, tt := range tests {
		got := SanitizeFontFamily(tt.input)
		if got != tt.expected {
			t.Errorf("SanitizeFontFamily(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}
