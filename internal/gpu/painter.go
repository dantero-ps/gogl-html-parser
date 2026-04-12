package gpu

import (
	"fmt"

	"goglweb/internal/layout"
	"goglweb/internal/render"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// GPUPainter implements the render.Painter interface.
type GPUPainter struct {
	shader       *Shader
	windowWidth  float64
	windowHeight float64

	// A pre-created mesh for performance.
	// We will update the data dynamically.
	rectMesh *Mesh

	// Default white texture - required for sampler2D uniform
	defaultTexture *Texture

	fontCache map[string]*FontAtlas
}

// NewGPUPainter creates a new painter object.
func NewGPUPainter(width, height float64, vertexPath, fragmentPath string) (*GPUPainter, error) {
	shader, err := NewShader(vertexPath, fragmentPath)
	if err != nil {
		return nil, err
	}

	// Create an empty initial mesh (UpdateRect will update the data)
	rectMesh, _ := NewRectMesh(0, 0, 1, 1, 1, 1, 1, 1)

	// Create default white texture (1x1 white pixel)
	whitePixel := []byte{255, 255, 255, 255}
	defaultTexture, err := NewTextureFromData(whitePixel, 1, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to create default texture: %w", err)
	}

	// Default texture parameters
	gl.BindTexture(gl.TEXTURE_2D, defaultTexture.ID)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	p := &GPUPainter{
		shader:         shader,
		windowWidth:    width,
		windowHeight:   height,
		rectMesh:       rectMesh,
		defaultTexture: defaultTexture,
		fontCache:      make(map[string]*FontAtlas),
	}

	p.updateProjection()
	return p, nil
}

// TextMeasurerAdapter wraps GPUPainter to implement layout.TextMeasurer.
// Font atlases are loaded lazily when first measured.
type TextMeasurerAdapter struct {
	painter *GPUPainter
}

// NewTextMeasurerAdapter returns a layout.TextMeasurer backed by real font glyph metrics.
func NewTextMeasurerAdapter(p *GPUPainter) layout.TextMeasurer {
	return &TextMeasurerAdapter{painter: p}
}

func (a *TextMeasurerAdapter) MeasureText(text string, fontFamily string, fontSize float64) layout.TextMetrics {
	atlas := a.painter.getOrLoadAtlas(fontFamily, fontSize)
	if atlas == nil {
		fb := &layout.FallbackMeasurer{}
		return fb.MeasureText(text, fontFamily, fontSize)
	}
	return atlas.MeasureText(text, fontFamily, fontSize)
}

func (a *TextMeasurerAdapter) MeasureWord(word string, fontFamily string, fontSize float64) float64 {
	return a.MeasureText(word, fontFamily, fontSize).Width
}

// getOrLoadAtlas returns (or lazily loads) a FontAtlas for the given family and size.
func (p *GPUPainter) getOrLoadAtlas(fontFamily string, fontSize float64) *FontAtlas {
	if fontSize <= 0 {
		fontSize = 16.0
	}
	cacheKey := fmt.Sprintf("%s-%.1f", fontFamily, fontSize)
	if atlas, ok := p.fontCache[cacheKey]; ok {
		return atlas
	}
	fontPath, err := findSystemFont(fontFamily)
	if err != nil {
		fontPath, err = findSystemFont(getDefaultFont())
		if err != nil {
			return nil
		}
	}
	atlas, err := p.BuildFontAtlas(fontPath, fontSize)
	if err != nil {
		return nil
	}
	p.fontCache[cacheKey] = atlas
	return atlas
}

// updateProjection sends the projection matrix to the shader.
func (p *GPUPainter) updateProjection() {
	p.shader.Use()

	// Transformation matrix from web coordinates (Y-down) to OpenGL NDC
	// 2/W,  0,    0, -1
	// 0,   -2/H,  0,  1
	// 0,    0,   -1,  0
	// 0,    0,    0,  1
	projection := [16]float32{
		float32(2.0 / p.windowWidth), 0, 0, 0,
		0, float32(-2.0 / p.windowHeight), 0, 0,
		0, 0, -1, 0,
		-1, 1, 0, 1,
	}
	p.shader.SetMat4("uProjection", &projection[0])
}

// SetViewport should be called when window size changes.
// Note: This function only updates the projection matrix, it doesn't set gl.Viewport.
// gl.Viewport should be set separately with framebuffer size (for Retina display support).
func (p *GPUPainter) SetViewport(width, height float64) {
	p.windowWidth = width
	p.windowHeight = height
	p.updateProjection()
}

// FillRect fills a rectangle.
func (p *GPUPainter) FillRect(rect layout.Rect, color render.Color) {
	p.shader.Use()
	p.shader.SetBool("uUseTexture", false)

	// Bind default texture for sampler uniform (OpenGL requirement)
	p.defaultTexture.Bind(0)
	p.shader.SetInt("uTexture", 0)

	r, g, b, a := normalizeColor(color)

	// Update mesh data and draw
	p.updateRectMesh(float32(rect.X), float32(rect.Y), float32(rect.Width), float32(rect.Height), r, g, b, a)
	p.rectMesh.Draw()
}

// DrawBorder draws borders.
func (p *GPUPainter) DrawBorder(rect layout.Rect, borders layout.EdgeSizes, color render.Color) {
	// Top edge
	if borders.Top > 0 {
		p.FillRect(layout.Rect{X: rect.X, Y: rect.Y, Width: rect.Width, Height: borders.Top}, color)
	}
	// Alt kenar
	if borders.Bottom > 0 {
		p.FillRect(layout.Rect{X: rect.X, Y: rect.Y + rect.Height - borders.Bottom, Width: rect.Width, Height: borders.Bottom}, color)
	}
	// Sol kenar
	if borders.Left > 0 {
		p.FillRect(layout.Rect{X: rect.X, Y: rect.Y + borders.Top, Width: borders.Left, Height: rect.Height - borders.Top - borders.Bottom}, color)
	}
	// Right edge
	if borders.Right > 0 {
		p.FillRect(layout.Rect{X: rect.X + rect.Width - borders.Right, Y: rect.Y + borders.Top, Width: borders.Right, Height: rect.Height - borders.Top - borders.Bottom}, color)
	}
}

// DrawText draws text to the screen. Automatically handles UTF-8 characters.
func (p *GPUPainter) DrawText(text string, x, y float64, fontSize float64, color render.Color, fontFamily string) {
	// 1. Get Font Atlas from cache or create it
	cacheKey := fmt.Sprintf("%s-%.1f", fontFamily, fontSize)
	atlas, ok := p.fontCache[cacheKey]

	if !ok {
		fontPath, err := findSystemFont(fontFamily)
		if err != nil {
			fmt.Printf("Font not found: %s, falling back to default\n", fontFamily)
			fontPath, err = findSystemFont(getDefaultFont())
			if err != nil {
				fmt.Printf("Default font not found: %v\n", err)
				return
			}
		}

		newAtlas, err := p.BuildFontAtlas(fontPath, fontSize)
		if err != nil {
			fmt.Printf("Failed to create font atlas: %v\n", err)
			return
		}
		p.fontCache[cacheKey] = newAtlas
		atlas = newAtlas
	}

	// 2. Shader Preparation
	p.shader.Use()
	p.shader.SetBool("uUseTexture", true)

	// Alpha Blending (critical for text transparency)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	atlas.Texture.Bind(0)
	p.shader.SetInt("uTexture", 0)

	r, g, b, a := normalizeColor(color)

	// 3. Draw Characters
	currentX := x

	// Convert text to runes (UTF-8)
	for _, char := range text {
		glyph, ok := atlas.GetGlyph(char)
		if !ok {
			fmt.Printf("Skipped character: %c\n", char)
			continue
		}

		// Position calculation: x + bearingX, y - bearingY (bearingY above baseline, usually positive?)
		// Note: BearingY is bounds.Min.Y, usually negative. Treat y as baseline, shift up with + BearingY.
		posX := float32(currentX + float64(glyph.BearingX))
		posY := float32(y - float64(atlas.Face.Metrics().Ascent.Round()) + float64(glyph.BearingY) + float64(atlas.Face.Metrics().Ascent.Round())) // Baseline adjustment
		// Simple: If text shifts down, use y + fontSize or ascent.
		// For testing: posY = float32(y) + float32(glyph.BearingY)  # If BearingY is negative, shifts up.

		w := float32(glyph.Width)
		h := float32(glyph.Height)

		// Calculate UV coordinates within atlas
		u0 := float32(glyph.X) / float32(atlas.AtlasWidth)
		v0 := float32(glyph.Y) / float32(atlas.AtlasHeight)
		u1 := float32(glyph.X+glyph.Width) / float32(atlas.AtlasWidth)
		v1 := float32(glyph.Y+glyph.Height) / float32(atlas.AtlasHeight)

		// Update mesh data with this character's coordinates
		p.updateRectMeshWithUV(posX, posY, w, h, u0, v0, u1, v1, r, g, b, a)

		p.rectMesh.Draw()

		// Advance to next character
		currentX += float64(glyph.AdvanceX)
	}

	gl.Disable(gl.BLEND)
}

// updateRectMeshWithUV updates the existing rectMesh with both coordinates and custom UV.
func (p *GPUPainter) updateRectMeshWithUV(x, y, w, h, u0, v0, u1, v1, r, g, b, a float32) {
	vertices := []Vertex{
		{[2]float32{x, y}, [4]float32{r, g, b, a}, [2]float32{u0, v0}},         // Top Left
		{[2]float32{x + w, y}, [4]float32{r, g, b, a}, [2]float32{u1, v0}},     // Top Right
		{[2]float32{x + w, y + h}, [4]float32{r, g, b, a}, [2]float32{u1, v1}}, // Bottom Right
		{[2]float32{x, y + h}, [4]float32{r, g, b, a}, [2]float32{u0, v1}},     // Bottom Left
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, p.rectMesh.VBO)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(vertices)*32, gl.Ptr(vertices))
}

// SetClip sets the clipping area using scissor test.
func (p *GPUPainter) SetClip(rect layout.Rect) {
	gl.Enable(gl.SCISSOR_TEST)
	// OpenGL Scissor accepts bottom-left (0,0), we convert.
	gl.Scissor(
		int32(rect.X),
		int32(p.windowHeight-(rect.Y+rect.Height)),
		int32(rect.Width),
		int32(rect.Height),
	)
}

// ClearClip disables scissor test.
func (p *GPUPainter) ClearClip() {
	gl.Disable(gl.SCISSOR_TEST)
}

// Helper Functions

func normalizeColor(c render.Color) (r, g, b, a float32) {
	return float32(c.R) / 255.0, float32(c.G) / 255.0, float32(c.B) / 255.0, float32(c.A) / 255.0
}

func (p *GPUPainter) updateRectMesh(x, y, w, h, r, g, b, a float32) {
	vertices := []Vertex{
		{[2]float32{x, y}, [4]float32{r, g, b, a}, [2]float32{0, 0}},
		{[2]float32{x + w, y}, [4]float32{r, g, b, a}, [2]float32{1, 0}},
		{[2]float32{x + w, y + h}, [4]float32{r, g, b, a}, [2]float32{1, 1}},
		{[2]float32{x, y + h}, [4]float32{r, g, b, a}, [2]float32{0, 1}},
	}

	gl.BindBuffer(gl.ARRAY_BUFFER, p.rectMesh.VBO)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(vertices)*32, gl.Ptr(vertices)) // 32 = vertex size in bytes
}

func (p *GPUPainter) Delete() {
	p.shader.Delete()
	p.rectMesh.Delete()
	if p.defaultTexture != nil {
		p.defaultTexture.Delete()
	}
}
