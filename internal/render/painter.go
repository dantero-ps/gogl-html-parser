package render

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"goglweb/internal/layout"
)

// Color represents an RGBA color.
type Color struct {
	R, G, B, A uint8
}

// Painter defines the interface for platform-agnostic drawing.
type Painter interface {
	FillRect(rect layout.Rect, color Color)
	DrawBorder(rect layout.Rect, borders layout.EdgeSizes, color Color)
	DrawText(text string, x, y float64, fontSize float64, color Color, fontFamily string)
	SetClip(rect layout.Rect)
	ClearClip()
	BeginScroll(clipRect layout.Rect, offsetX, offsetY float64)
	EndScroll()
}

// MockPainter records operations for testing purposes.
type MockPainter struct {
	Operations       []string
	BeginScrollCalls int
	EndScrollCalls   int
}

func (p *MockPainter) FillRect(rect layout.Rect, color Color) {
	p.Operations = append(p.Operations, fmt.Sprintf("FillRect: %+v Color: %+v", rect, color))
}

func (p *MockPainter) DrawBorder(rect layout.Rect, borders layout.EdgeSizes, color Color) {
	p.Operations = append(p.Operations, fmt.Sprintf("DrawBorder: %+v Edges: %+v Color: %+v", rect, borders, color))
}

func (p *MockPainter) DrawText(text string, x, y float64, fontSize float64, color Color, fontFamily string) {
	p.Operations = append(p.Operations, fmt.Sprintf("DrawText: '%s' at (%.1f, %.1f) Size: %.1f Color: %+v", text, x, y, fontSize, color))
}

func (p *MockPainter) SetClip(rect layout.Rect) {
	p.Operations = append(p.Operations, fmt.Sprintf("SetClip: %+v", rect))
}

func (p *MockPainter) ClearClip() {
	p.Operations = append(p.Operations, "ClearClip")
}

func (p *MockPainter) BeginScroll(clipRect layout.Rect, offsetX, offsetY float64) {
	p.BeginScrollCalls++
	p.Operations = append(p.Operations, fmt.Sprintf("BeginScroll: ClipRect: %+v OffsetX: %.2f OffsetY: %.2f", clipRect, offsetX, offsetY))
}

func (p *MockPainter) EndScroll() {
	p.EndScrollCalls++
	p.Operations = append(p.Operations, "EndScroll")
}

var namedColors = map[string]Color{
	"black":  {0, 0, 0, 255},
	"white":  {255, 255, 255, 255},
	"red":    {255, 0, 0, 255},
	"green":  {0, 255, 0, 255},
	"blue":   {0, 0, 255, 255},
	"yellow": {255, 255, 0, 255},
	"cyan":   {0, 255, 255, 255},
	"gray":   {128, 128, 128, 255},
	"silver": {192, 192, 192, 255},
	"maroon": {128, 0, 0, 255},
	"navy":   {0, 0, 128, 255},
	"orange": {255, 165, 0, 255},
	"purple": {128, 0, 128, 255},
}

// ParseColor converts CSS color strings to Render Color.
func ParseColor(colorStr string) Color {
	colorStr = strings.ToLower(strings.TrimSpace(colorStr))
	if colorStr == "" || colorStr == "transparent" {
		return Color{0, 0, 0, 0}
	}

	// Named Colors
	if c, ok := namedColors[colorStr]; ok {
		return c
	}

	// Hex #RGB or #RRGGBB
	if strings.HasPrefix(colorStr, "#") {
		return parseHex(colorStr)
	}

	// RGB and RGBA
	if strings.HasPrefix(colorStr, "rgb") {
		return parseRGB(colorStr)
	}

	return Color{0, 0, 0, 255} // Default black
}

func parseHex(hex string) Color {
	hex = strings.TrimPrefix(hex, "#")
	var r, g, b, a uint8 = 0, 0, 0, 255

	if len(hex) == 3 {
		fmt.Sscanf(hex, "%1x%1x%1x", &r, &g, &b)
		r, g, b = r*17, g*17, b*17
	} else if len(hex) == 6 {
		fmt.Sscanf(hex, "%2x%2x%2x", &r, &g, &b)
	} else if len(hex) == 8 {
		fmt.Sscanf(hex, "%2x%2x%2x%2x", &r, &g, &b, &a)
	}
	return Color{r, g, b, a}
}

func parseRGB(rgbStr string) Color {
	re := regexp.MustCompile(`rgba?\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*(?:,\s*([\d.]+)\s*)?\)`)
	matches := re.FindStringSubmatch(rgbStr)
	if len(matches) < 4 {
		return Color{0, 0, 0, 255}
	}

	r, _ := strconv.Atoi(matches[1])
	g, _ := strconv.Atoi(matches[2])
	b, _ := strconv.Atoi(matches[3])
	alpha := 1.0
	if matches[4] != "" {
		alpha, _ = strconv.ParseFloat(matches[4], 64)
	}

	return Color{uint8(r), uint8(g), uint8(b), uint8(alpha * 255)}
}
