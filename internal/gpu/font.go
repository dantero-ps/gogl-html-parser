package gpu

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/furkandgn/goglweb/internal/layout"

	"github.com/go-gl/gl/v4.1-core/gl"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

// GlyphInfo holds a character's position and dimensions on the atlas.
type GlyphInfo struct {
	X, Y          int // Pixel coordinates within atlas
	Width, Height int // Character's pixel dimensions
	AdvanceX      int // Distance to next character
	BearingX      int // Left spacing (offset)
	BearingY      int // Vertical offset relative to baseline
}

// FontAtlas is a dynamically expanding character table.
type FontAtlas struct {
	Texture      *Texture
	Characters   map[rune]GlyphInfo
	Face         font.Face
	NextX, NextY int
	LineHeight   int
	AtlasWidth   int
	AtlasHeight  int
	FontSize     float64
	PixelRatio   float64 // display pixel ratio (e.g. 2.0 for Retina); glyphs rasterized at FontSize*PixelRatio
}

// GetGlyph returns the character if it exists in atlas, otherwise adds it dynamically.
func (a *FontAtlas) GetGlyph(r rune) (GlyphInfo, bool) {
	if info, ok := a.Characters[r]; ok {
		return info, true
	}

	// Create character from font face
	dot := fixed.P(0, 0)
	dr, mask, maskp, adv, ok := a.Face.Glyph(dot, r)
	if !ok {
		return a.GetGlyph('?')
	}

	// Check space on atlas
	if a.NextX+dr.Dx() >= a.AtlasWidth {
		a.NextX = 5
		a.NextY += a.LineHeight + 2
	}

	// If atlas fills vertically, expand atlas
	if a.NextY+dr.Dy() >= a.AtlasHeight {
		// New size: double the height
		newHeight := a.AtlasHeight * 2
		newData := make([]byte, a.AtlasWidth*newHeight*4)
		// Copy old data (if Texture.Data exists, otherwise keep CPU copy)
		// Note: Add Data []byte to Texture struct and update it.
		// Here simply create new texture and copy old one (copy on GPU with GL_READ_FRAMEBUFFER etc.)
		// For simplicity: Read old texture and write to new (gl.GetTexImage)
		var oldData []byte = make([]byte, a.AtlasWidth*a.AtlasHeight*4)
		gl.BindTexture(gl.TEXTURE_2D, a.Texture.ID)
		gl.GetTexImage(gl.TEXTURE_2D, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(oldData))
		copy(newData, oldData)

		a.Texture.Delete()
		var err error
		a.Texture, err = NewTextureFromData(newData, a.AtlasWidth, newHeight)
		if err != nil {
			fmt.Printf("Atlas expansion error: %v\n", err)
			return GlyphInfo{}, false
		}
		a.AtlasHeight = newHeight
	}

	// Draw glyph to temporary RGBA image (white color with alpha mask)
	rgba := image.NewRGBA(dr.Bounds())
	draw.DrawMask(rgba, rgba.Bounds(), &image.Uniform{color.White}, image.Point{}, mask, maskp, draw.Src)

	// Send only relevant region to GPU texture (critical for performance)
	// Use advance even for empty glyphs, but don't upload if empty
	if dr.Dx() > 0 && dr.Dy() > 0 {
		gl.BindTexture(gl.TEXTURE_2D, a.Texture.ID)
		gl.TexSubImage2D(
			gl.TEXTURE_2D, 0,
			int32(a.NextX), int32(a.NextY),
			int32(dr.Dx()), int32(dr.Dy()),
			gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix),
		)
	}

	// Use dr.Min for bearing — it is exactly where face.Glyph placed the pixels,
	// so it matches the atlas image without the ±1px rounding error that
	// bounds.Min.X/Y.Round() can introduce.
	info := GlyphInfo{
		X:        a.NextX,
		Y:        a.NextY,
		Width:    dr.Dx(),
		Height:   dr.Dy(),
		AdvanceX: adv.Round(),
		BearingX: dr.Min.X,
		BearingY: dr.Min.Y,
	}

	a.Characters[r] = info
	a.NextX += dr.Dx() + 2
	if dr.Dy() > a.LineHeight {
		a.LineHeight = dr.Dy()
	}

	return info, true
}

// SanitizeFontFamily cleans a CSS font-family value for safe filesystem lookup.
// Removes quotes, path traversal attempts, null bytes, and shell metacharacters.
func SanitizeFontFamily(fontFamily string) string {
	if fontFamily == "" {
		return ""
	}

	// Strip surrounding quotes (single or double)
	fontFamily = strings.Trim(fontFamily, `"'`)

	// Remove null bytes
	fontFamily = strings.ReplaceAll(fontFamily, "\x00", "")

	// Remove path traversal characters
	fontFamily = strings.ReplaceAll(fontFamily, "..", "")
	fontFamily = strings.ReplaceAll(fontFamily, "/", "")
	fontFamily = strings.ReplaceAll(fontFamily, "\\", "")

	// Remove shell metacharacters
	dangerous := []string{";", "|", "&", "$", "`", "(", ")", "{", "}", "<", ">"}
	for _, ch := range dangerous {
		fontFamily = strings.ReplaceAll(fontFamily, ch, "")
	}

	// Trim whitespace
	fontFamily = strings.TrimSpace(fontFamily)

	return fontFamily
}

func findStyledFont(fontFamily string, fontWeight string, fontStyle string, extraPaths []string) string {
	fontFamily = SanitizeFontFamily(fontFamily)
	if fontFamily == "" {
		return ""
	}

	isBold, isItalic := resolveFontOptions(fontWeight, fontStyle)
	if !isBold && !isItalic {
		return ""
	}

	result := searchStyledFontFiles(fontFamily, isBold, isItalic, extraPaths)
	if result != "" {
		return result
	}

	if strings.ToLower(fontFamily) != "arial" {
		result = searchStyledFontFiles("arial", isBold, isItalic, extraPaths)
		if result != "" {
			return result
		}
	}

	return ""
}

func searchStyledFontFiles(fontFamily string, isBold bool, isItalic bool, extraPaths []string) string {
	fontFamily = SanitizeFontFamily(fontFamily)
	if fontFamily == "" {
		return ""
	}

	var searchPaths []string
	switch runtime.GOOS {
	case "windows":
		searchPaths = []string{os.Getenv("WINDIR") + "\\Fonts"}
	case "darwin":
		searchPaths = []string{"/System/Library/Fonts/Supplemental", "/System/Library/Fonts", "/Library/Fonts"}
	case "linux":
		searchPaths = []string{"/usr/share/fonts", "/usr/local/share/fonts"}
	}
	searchPaths = append(searchPaths, extraPaths...)

	lowerFamily := strings.ToLower(fontFamily)
	extensions := []string{".ttf", ".otf"}

	for _, root := range searchPaths {
		var foundPath string
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			name := strings.ToLower(info.Name())
			for _, ext := range extensions {
				if !strings.HasSuffix(name, ext) {
					continue
				}
				if !strings.Contains(name, lowerFamily) {
					continue
				}
				if isBold && isItalic {
					if strings.Contains(name, "bold") && (strings.Contains(name, "italic") || strings.Contains(name, "oblique")) {
						foundPath = path
						return filepath.SkipDir
					}
				} else if isBold {
					if strings.Contains(name, "bold") && !strings.Contains(name, "italic") {
						foundPath = path
						return filepath.SkipDir
					}
				} else if isItalic {
					if (strings.Contains(name, "italic") || strings.Contains(name, "oblique")) && !strings.Contains(name, "bold") {
						foundPath = path
						return filepath.SkipDir
					}
				}
			}
			return nil
		})
		if err == nil && foundPath != "" {
			return foundPath
		}
	}

	return ""
}

func resolveFontOptions(fontWeight string, fontStyle string) (isBold bool, isItalic bool) {
	isBold = fontWeight == "bold" || fontWeight == "bolder" || fontWeight == "700" || fontWeight == "800" || fontWeight == "900"
	isItalic = fontStyle == "italic" || fontStyle == "oblique"
	return
}

// findSystemFont finds the file path in the system for the given font family name.
func findSystemFont(fontFamily string) (string, error) {
	fontFamily = SanitizeFontFamily(fontFamily)
	if fontFamily == "" {
		return "", fmt.Errorf("empty font family after sanitization")
	}
	var searchPaths []string
	switch runtime.GOOS {
	case "windows":
		searchPaths = []string{os.Getenv("WINDIR") + "\\Fonts"}
	case "darwin":
		searchPaths = []string{"/System/Library/Fonts", "/Library/Fonts"}
	case "linux":
		searchPaths = []string{"/usr/share/fonts", "/usr/local/share/fonts"}
	}

	extensions := []string{".ttf", ".otf", ".ttc"}
	for _, root := range searchPaths {
		var foundPath string
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			name := strings.ToLower(info.Name())
			lowerFamily := strings.ToLower(fontFamily)
			for _, ext := range extensions {
				if strings.Contains(name, lowerFamily) && strings.HasSuffix(name, ext) {
					foundPath = path
					return filepath.SkipDir
				}
			}
			return nil
		})
		if err != nil {
			return "", err
		}
		if foundPath != "" {
			return foundPath, nil
		}
	}
	return "", fmt.Errorf("font not found: %s", fontFamily)
}

// findFontInPaths searches extra font directories for the given font family.
// Returns the first matching file path, or empty string if not found.
func findFontInPaths(fontFamily string, paths []string) string {
	fontFamily = SanitizeFontFamily(fontFamily)
	if fontFamily == "" {
		return ""
	}
	extensions := []string{".ttf", ".otf", ".ttc"}
	lowerFamily := strings.ToLower(fontFamily)
	for _, root := range paths {
		var foundPath string
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			name := strings.ToLower(info.Name())
			for _, ext := range extensions {
				if strings.Contains(name, lowerFamily) && strings.HasSuffix(name, ext) {
					foundPath = path
					return filepath.SkipDir
				}
			}
			return nil
		})
		if err == nil && foundPath != "" {
			return foundPath
		}
	}
	return ""
}

// getDefaultFont returns the default font name according to platform.
func getDefaultFont() string {
	switch runtime.GOOS {
	case "windows":
		return "arial"
	case "darwin":
		return "helvetica"
	default:
		return "dejavu"
	}
}

// MeasureText measures the width of text using real glyph advance widths.
// The fontFamily parameter is ignored (uses the atlas's loaded face).
// fontSize scales relative to the atlas's base font size.
func (a *FontAtlas) MeasureText(text string, fontFamily string, fontSize float64) layout.TextMetrics {
	// Face was rasterized at FontSize*PixelRatio, so raw advance values are in
	// physical pixels. Divide by PixelRatio to get logical CSS pixels, then
	// apply the requested fontSize scale relative to our logical FontSize.
	pr := a.PixelRatio
	if pr <= 0 {
		pr = 1.0
	}
	scale := fontSize / a.FontSize
	// Sum per-character rounded advances to match DrawText's rendering exactly.
	// DrawText does: currentX += float64(glyph.AdvanceX) / pr
	// where AdvanceX = adv.Round() (per character). We must use the same
	// approach here; accumulating fixed.Int26_6 and rounding at the end
	// introduces per-character rounding drift that shifts inline siblings.
	var totalAdvancePx float64
	for _, r := range text {
		glyph, ok := a.GetGlyph(r)
		var advPx int
		if ok {
			advPx = glyph.AdvanceX
		} else {
			advPx = int(a.FontSize)
		}
		totalAdvancePx += float64(advPx) / pr
	}
	metrics := a.Face.Metrics()
	lineHeight := float64((metrics.Ascent + metrics.Descent).Round()) / pr
	ascent := float64(metrics.Ascent.Round()) / pr
	return layout.TextMetrics{
		Width:      totalAdvancePx * scale,
		Height:     lineHeight * scale,
		Ascent:     ascent * scale,
		LineHeight: lineHeight * scale * 1.2,
	}
}

// MeasureWord measures a single word using real glyph advance widths.
func (a *FontAtlas) MeasureWord(word string, fontFamily string, fontSize float64) float64 {
	return a.MeasureText(word, fontFamily, fontSize).Width
}

// BuildFontAtlas initializes a new dynamic atlas from a font file.
// pixelRatio is the display pixel density multiplier (e.g. 2.0 for Retina/HiDPI).
// Glyphs are rasterized at fontSize*pixelRatio for crisp rendering on high-DPI displays.
func (p *GPUPainter) BuildFontAtlas(fontPath string, fontSize float64, pixelRatio float64) (*FontAtlas, error) {
	if pixelRatio <= 0 {
		pixelRatio = 1.0
	}
	fontBytes, err := os.ReadFile(fontPath)
	if err != nil {
		return nil, err
	}

	var f *sfnt.Font
	// Distinguish TTC and TTF
	if collection, err := opentype.ParseCollection(fontBytes); err == nil {
		f, _ = collection.Font(0)
	} else {
		f, err = opentype.Parse(fontBytes)
		if err != nil {
			return nil, err
		}
	}

	// Rasterize at physical pixel size: fontSize * pixelRatio.
	// This gives full-resolution glyph bitmaps on HiDPI/Retina displays.
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    fontSize * pixelRatio,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}

	// Create an empty texture initially (e.g. 1024x1024)
	atlasWidth, atlasHeight := 1024, 1024
	emptyData := make([]byte, atlasWidth*atlasHeight*4)
	tex, err := NewTextureFromData(emptyData, atlasWidth, atlasHeight)
	if err != nil {
		return nil, err
	}

	// Set texture parameters (for filtering)
	gl.BindTexture(gl.TEXTURE_2D, tex.ID)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	return &FontAtlas{
		Texture:     tex,
		Characters:  make(map[rune]GlyphInfo),
		Face:        face,
		AtlasWidth:  atlasWidth,
		AtlasHeight: atlasHeight,
		FontSize:    fontSize,
		PixelRatio:  pixelRatio,
		NextX:       5,
		NextY:       5,
	}, nil
}
