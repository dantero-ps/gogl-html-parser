package render

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/furkandgn/goglweb/internal/layout"
)

// Color represents an RGBA color.
type Color struct {
	R, G, B, A uint8
}

// Painter defines the interface for platform-agnostic drawing.
type Painter interface {
	FillRect(rect layout.Rect, color Color)
	DrawBorder(rect layout.Rect, borders layout.EdgeSizes, color Color)
	DrawText(text string, x, y float64, fontSize float64, color Color, fontFamily string, textAlign string, containerWidth float64, fontWeight string, fontStyle string)
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

func (p *MockPainter) DrawText(text string, x, y float64, fontSize float64, color Color, fontFamily string, textAlign string, containerWidth float64, fontWeight string, fontStyle string) {
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
	"black":            {0, 0, 0, 255},
	"white":            {255, 255, 255, 255},
	"red":              {255, 0, 0, 255},
	"green":            {0, 128, 0, 255},
	"lime":             {0, 255, 0, 255},
	"blue":             {0, 0, 255, 255},
	"yellow":           {255, 255, 0, 255},
	"cyan":             {0, 255, 255, 255},
	"aqua":             {0, 255, 255, 255},
	"magenta":          {255, 0, 255, 255},
	"fuchsia":          {255, 0, 255, 255},
	"gray":             {128, 128, 128, 255},
	"grey":             {128, 128, 128, 255},
	"silver":           {192, 192, 192, 255},
	"maroon":           {128, 0, 0, 255},
	"navy":             {0, 0, 128, 255},
	"orange":           {255, 165, 0, 255},
	"purple":           {128, 0, 128, 255},
	"teal":             {0, 128, 128, 255},
	"olive":            {128, 128, 0, 255},
	"pink":             {255, 192, 203, 255},
	"brown":            {165, 42, 42, 255},
	"coral":            {255, 127, 80, 255},
	"salmon":           {250, 128, 114, 255},
	"gold":             {255, 215, 0, 255},
	"khaki":            {240, 230, 140, 255},
	"indigo":           {75, 0, 130, 255},
	"violet":           {238, 130, 238, 255},
	"plum":             {221, 160, 221, 255},
	"orchid":           {218, 112, 214, 255},
	"turquoise":        {64, 224, 208, 255},
	"skyblue":          {135, 206, 235, 255},
	"steelblue":        {70, 130, 180, 255},
	"cornflowerblue":   {100, 149, 237, 255},
	"royalblue":        {65, 105, 225, 255},
	"dodgerblue":       {30, 144, 255, 255},
	"deepskyblue":      {0, 191, 255, 255},
	"lightblue":        {173, 216, 230, 255},
	"powderblue":       {176, 224, 230, 255},
	"lightgreen":       {144, 238, 144, 255},
	"palegreen":        {152, 251, 152, 255},
	"darkgreen":        {0, 100, 0, 255},
	"forestgreen":      {34, 139, 34, 255},
	"seagreen":         {46, 139, 87, 255},
	"mediumseagreen":   {60, 179, 113, 255},
	"springgreen":      {0, 255, 127, 255},
	"chartreuse":       {127, 255, 0, 255},
	"yellowgreen":      {154, 205, 50, 255},
	"greenyellow":      {173, 255, 47, 255},
	"lightyellow":      {255, 255, 224, 255},
	"lemonchiffon":     {255, 250, 205, 255},
	"peachpuff":        {255, 218, 185, 255},
	"bisque":           {255, 228, 196, 255},
	"wheat":            {245, 222, 179, 255},
	"tan":              {210, 180, 140, 255},
	"sienna":           {160, 82, 45, 255},
	"chocolate":        {210, 105, 30, 255},
	"peru":             {205, 133, 63, 255},
	"burlywood":        {222, 184, 135, 255},
	"sandybrown":       {244, 164, 96, 255},
	"tomato":           {255, 99, 71, 255},
	"orangered":        {255, 69, 0, 255},
	"crimson":          {220, 20, 60, 255},
	"firebrick":        {178, 34, 34, 255},
	"darkred":          {139, 0, 0, 255},
	"hotpink":          {255, 105, 180, 255},
	"deeppink":         {255, 20, 147, 255},
	"lightpink":        {255, 182, 193, 255},
	"palevioletred":    {219, 112, 147, 255},
	"mediumpurple":     {147, 112, 219, 255},
	"darkorchid":       {153, 50, 204, 255},
	"darkviolet":       {148, 0, 211, 255},
	"blueviolet":       {138, 43, 226, 255},
	"mediumblue":       {0, 0, 205, 255},
	"darkblue":         {0, 0, 139, 255},
	"midnightblue":     {25, 25, 112, 255},
	"slateblue":        {106, 90, 205, 255},
	"mediumslateblue":  {123, 104, 238, 255},
	"lightslategray":   {119, 136, 153, 255},
	"slategray":        {112, 128, 144, 255},
	"darkslategray":    {47, 79, 79, 255},
	"lightgray":        {211, 211, 211, 255},
	"lightgrey":        {211, 211, 211, 255},
	"gainsboro":        {220, 220, 220, 255},
	"whitesmoke":       {245, 245, 245, 255},
	"ghostwhite":       {248, 248, 255, 255},
	"aliceblue":        {240, 248, 255, 255},
	"azure":            {240, 255, 255, 255},
	"honeydew":         {240, 255, 240, 255},
	"mintcream":        {245, 255, 250, 255},
	"ivory":            {255, 255, 240, 255},
	"floralwhite":      {255, 250, 240, 255},
	"seashell":         {255, 245, 238, 255},
	"linen":            {250, 240, 230, 255},
	"lavender":         {230, 230, 250, 255},
	"lavenderblush":    {255, 240, 245, 255},
	"mistyrose":        {255, 228, 225, 255},
	"snow":             {255, 250, 250, 255},
	"oldlace":          {253, 245, 230, 255},
	"antiquewhite":     {250, 235, 215, 255},
	"moccasin":         {255, 228, 181, 255},
	"navajowhite":      {255, 222, 173, 255},
	"cornsilk":         {255, 248, 220, 255},
	"blanchedalmond":   {255, 235, 205, 255},
	"papayawhip":       {255, 239, 213, 255},
	"darkkhaki":        {189, 183, 107, 255},
	"goldenrod":        {218, 165, 32, 255},
	"darkgoldenrod":    {184, 134, 11, 255},
	"palegoldenrod":    {238, 232, 170, 255},
	"cadetblue":        {95, 158, 160, 255},
	"aquamarine":       {127, 255, 212, 255},
	"mediumaquamarine": {102, 205, 170, 255},
	"mediumturquoise":  {72, 209, 204, 255},
	"darkturquoise":    {0, 206, 209, 255},
	"lightseagreen":    {32, 178, 170, 255},
	"darkcyan":         {0, 139, 139, 255},
	"darkslateblue":    {72, 61, 139, 255},
	"rosybrown":        {188, 143, 143, 255},
	"indianred":        {205, 92, 92, 255},
	"dimgray":          {105, 105, 105, 255},
	"dimgrey":          {105, 105, 105, 255},
	"darkgray":         {169, 169, 169, 255},
	"darkgrey":         {169, 169, 169, 255},
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
