package gpu

import (
	"fmt"

	"github.com/furkandgn/goglweb/internal/layout"
	"github.com/furkandgn/goglweb/internal/render"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// GPUPainter implements the render.Painter interface.
type GPUPainter struct {
	shader       *Shader
	windowWidth  float64
	windowHeight float64
	fbWidth      float64
	fbHeight     float64

	// A pre-created mesh for performance.
	// We will update the data dynamically.
	rectMesh *Mesh

	// Default white texture - required for sampler2D uniform
	defaultTexture *Texture

	fontCache   map[string]*atlasCacheEntry
	accessClock uint64

	batchMesh    *DynamicMesh
	rectVerts    []Vertex
	glyphBatches map[uint32][]Vertex

	scrollStack []scrollState

	// extraFontPaths are additional directories to search for fonts before system paths.
	extraFontPaths []string
}

// scrollState holds the clip rect and offset for a scrollable container.
type scrollState struct {
	offsetX, offsetY float64
	clip             layout.Rect
}

type atlasCacheEntry struct {
	atlas      *FontAtlas
	lastAccess uint64
}

const maxAtlasCacheSize = 20

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
		fbWidth:        width,
		fbHeight:       height,
		rectMesh:       rectMesh,
		defaultTexture: defaultTexture,
		fontCache:      make(map[string]*atlasCacheEntry),
		batchMesh:      NewDynamicMesh(),
		rectVerts:      make([]Vertex, 0, 2048),
		glyphBatches:   make(map[uint32][]Vertex),
	}

	p.updateProjection()
	return p, nil
}

// NewGPUPainterFromSource creates a new painter from shader source strings instead of file paths.
func NewGPUPainterFromSource(width, height float64, vertexSrc, fragmentSrc string) (*GPUPainter, error) {
	shader, err := NewShaderFromSource(vertexSrc, fragmentSrc)
	if err != nil {
		return nil, err
	}

	rectMesh, _ := NewRectMesh(0, 0, 1, 1, 1, 1, 1, 1)

	whitePixel := []byte{255, 255, 255, 255}
	defaultTexture, err := NewTextureFromData(whitePixel, 1, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to create default texture: %w", err)
	}

	gl.BindTexture(gl.TEXTURE_2D, defaultTexture.ID)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	p := &GPUPainter{
		shader:         shader,
		windowWidth:    width,
		windowHeight:   height,
		fbWidth:        width,
		fbHeight:       height,
		rectMesh:       rectMesh,
		defaultTexture: defaultTexture,
		fontCache:      make(map[string]*atlasCacheEntry),
		batchMesh:      NewDynamicMesh(),
		rectVerts:      make([]Vertex, 0, 2048),
		glyphBatches:   make(map[uint32][]Vertex),
	}

	p.updateProjection()
	return p, nil
}

// SetExtraFontPaths sets additional directories to search for fonts.
func (p *GPUPainter) SetExtraFontPaths(paths []string) {
	p.extraFontPaths = paths
}

// Font atlases are loaded lazily when first measured.
type TextMeasurerAdapter struct {
	painter *GPUPainter
}

// NewTextMeasurerAdapter returns a layout.TextMeasurer backed by real font glyph metrics.
func NewTextMeasurerAdapter(p *GPUPainter) layout.TextMeasurer {
	return &TextMeasurerAdapter{painter: p}
}

func (a *TextMeasurerAdapter) MeasureText(text string, fontFamily string, fontSize float64, fontWeight string, fontStyle string) layout.TextMetrics {
	atlas := a.painter.getOrLoadAtlas(fontFamily, fontSize, fontWeight, fontStyle)
	if atlas == nil {
		fb := &layout.FallbackMeasurer{}
		return fb.MeasureText(text, fontFamily, fontSize, fontWeight, fontStyle)
	}
	return atlas.MeasureText(text, fontFamily, fontSize)
}

func (a *TextMeasurerAdapter) MeasureWord(word string, fontFamily string, fontSize float64, fontWeight string, fontStyle string) float64 {
	return a.MeasureText(word, fontFamily, fontSize, fontWeight, fontStyle).Width
}

// pixelRatio returns the display pixel density ratio (physical / logical).
// On Retina/HiDPI displays this is typically 2.0; on standard displays 1.0.
func (p *GPUPainter) pixelRatio() float64 {
	if p.windowWidth <= 0 {
		return 1.0
	}
	r := p.fbWidth / p.windowWidth
	if r < 1.0 {
		return 1.0
	}
	return r
}

// getOrLoadAtlas returns (or lazily loads) a FontAtlas for the given family and size.
func (p *GPUPainter) getOrLoadAtlas(fontFamily string, fontSize float64, fontWeight string, fontStyle string) *FontAtlas {
	if fontSize <= 0 {
		fontSize = 16.0
	}
	pr := p.pixelRatio()
	cacheKey := fmt.Sprintf("%s-%.1f@%.2f-%s-%s", fontFamily, fontSize, pr, fontWeight, fontStyle)
	p.accessClock++

	if entry, ok := p.fontCache[cacheKey]; ok {
		entry.lastAccess = p.accessClock
		return entry.atlas
	}

	fontPath := findStyledFont(fontFamily, fontWeight, fontStyle, p.extraFontPaths)
	if fontPath == "" {
		var err error
		fontPath, err = findSystemFont(fontFamily)
		if err != nil {
			fontPath = findFontInPaths(fontFamily, p.extraFontPaths)
			if fontPath == "" {
				fontPath, err = findSystemFont(getDefaultFont())
				if err != nil {
					return nil
				}
			}
		}
	}
	atlas, err := p.BuildFontAtlas(fontPath, fontSize, pr)
	if err != nil {
		return nil
	}

	if len(p.fontCache) >= maxAtlasCacheSize {
		p.evictOldest()
	}

	p.fontCache[cacheKey] = &atlasCacheEntry{atlas: atlas, lastAccess: p.accessClock}
	return atlas
}

func (p *GPUPainter) evictOldest() {
	var oldestKey string
	var oldestAccess uint64 = ^uint64(0)
	for k, e := range p.fontCache {
		if e.lastAccess < oldestAccess {
			oldestAccess = e.lastAccess
			oldestKey = k
		}
	}
	if oldestKey != "" {
		entry := p.fontCache[oldestKey]
		if entry.atlas != nil {
			entry.atlas.Texture.Delete()
		}
		delete(p.fontCache, oldestKey)
	}
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

// SetFramebufferSize stores the physical framebuffer dimensions used for scissor test coordinates.
// Must be called after init and whenever the framebuffer is resized (e.g. window resize on Retina).
func (p *GPUPainter) SetFramebufferSize(width, height float64) {
	p.fbWidth = width
	p.fbHeight = height
}

// FillRect fills a rectangle.
func (p *GPUPainter) FillRect(rect layout.Rect, color render.Color) {
	dx, dy := p.scrollOffset()
	x := float32(rect.X + dx)
	y := float32(rect.Y + dy)
	w := float32(rect.Width)
	h := float32(rect.Height)
	r, g, b, a := normalizeColor(color)
	p.rectVerts = append(p.rectVerts,
		Vertex{[2]float32{x, y}, [4]float32{r, g, b, a}, [2]float32{0, 0}},
		Vertex{[2]float32{x + w, y}, [4]float32{r, g, b, a}, [2]float32{1, 0}},
		Vertex{[2]float32{x + w, y + h}, [4]float32{r, g, b, a}, [2]float32{1, 1}},
		Vertex{[2]float32{x, y + h}, [4]float32{r, g, b, a}, [2]float32{0, 1}},
	)
}

// DrawBorder draws borders.
func (p *GPUPainter) DrawBorder(rect layout.Rect, borders layout.EdgeSizes, color render.Color) {
	// Use FillRect directly — it applies scrollOffset() internally, so don't add it here.
	if borders.Top > 0 {
		p.FillRect(layout.Rect{X: rect.X, Y: rect.Y, Width: rect.Width, Height: borders.Top}, color)
	}
	if borders.Bottom > 0 {
		p.FillRect(layout.Rect{X: rect.X, Y: rect.Y + rect.Height - borders.Bottom, Width: rect.Width, Height: borders.Bottom}, color)
	}
	if borders.Left > 0 {
		p.FillRect(layout.Rect{X: rect.X, Y: rect.Y + borders.Top, Width: borders.Left, Height: rect.Height - borders.Top - borders.Bottom}, color)
	}
	if borders.Right > 0 {
		p.FillRect(layout.Rect{X: rect.X + rect.Width - borders.Right, Y: rect.Y + borders.Top, Width: borders.Right, Height: rect.Height - borders.Top - borders.Bottom}, color)
	}
}

// DrawText draws text to the screen. Automatically handles UTF-8 characters.
func (p *GPUPainter) DrawText(text string, x, y float64, fontSize float64, color render.Color, fontFamily string, textAlign string, containerWidth float64, fontWeight string, fontStyle string) {
	dx, dy := p.scrollOffset()
	x += dx
	y += dy

	atlas := p.getOrLoadAtlas(fontFamily, fontSize, fontWeight, fontStyle)
	if atlas == nil {
		return
	}

	r, g, b, a := normalizeColor(color)
	texID := atlas.Texture.ID

	// Glyphs were rasterized at PixelRatio× the CSS font size.
	// Divide all physical pixel dimensions by PixelRatio to get logical CSS pixel sizes
	// so the quads match the layout geometry while sampling full-resolution glyph texels.
	pr := atlas.PixelRatio
	if pr <= 0 {
		pr = 1.0
	}

	faceMetrics := atlas.Face.Metrics()
	ascent := float64(faceMetrics.Ascent.Round()) / pr
	descent := float64(faceMetrics.Descent.Round()) / pr
	lineHeight := (ascent + descent) * 1.2

	var lines []string
	if containerWidth > 0 {
		measurer := &TextMeasurerAdapter{painter: p}
		lines = layout.WordWrap(text, fontFamily, fontSize, containerWidth, measurer, fontWeight, fontStyle)
	}
	if len(lines) == 0 {
		lines = []string{text}
	}

	currentY := y
	for _, line := range lines {
		var lineX float64 = x
		if (textAlign == "center" || textAlign == "right") && containerWidth > 0 {
			var textWidth float64
			for _, char := range line {
				if glyph, ok := atlas.GetGlyph(char); ok {
					textWidth += float64(glyph.AdvanceX) / pr
				}
			}
			switch textAlign {
			case "center":
				lineX += (containerWidth - textWidth) / 2
			case "right":
				lineX += containerWidth - textWidth
			}
		}

		currentX := lineX
		for _, char := range line {
			glyph, ok := atlas.GetGlyph(char)
			if !ok {
				continue
			}

			posX := float32(currentX + float64(glyph.BearingX)/pr)
			posY := float32(currentY + float64(glyph.BearingY)/pr)
			w := float32(float64(glyph.Width) / pr)
			h := float32(float64(glyph.Height) / pr)

			u0 := float32(glyph.X) / float32(atlas.AtlasWidth)
			v0 := float32(glyph.Y) / float32(atlas.AtlasHeight)
			u1 := float32(glyph.X+glyph.Width) / float32(atlas.AtlasWidth)
			v1 := float32(glyph.Y+glyph.Height) / float32(atlas.AtlasHeight)

			p.glyphBatches[texID] = append(p.glyphBatches[texID],
				Vertex{[2]float32{posX, posY}, [4]float32{r, g, b, a}, [2]float32{u0, v0}},
				Vertex{[2]float32{posX + w, posY}, [4]float32{r, g, b, a}, [2]float32{u1, v0}},
				Vertex{[2]float32{posX + w, posY + h}, [4]float32{r, g, b, a}, [2]float32{u1, v1}},
				Vertex{[2]float32{posX, posY + h}, [4]float32{r, g, b, a}, [2]float32{u0, v1}},
			)
			currentX += float64(glyph.AdvanceX) / pr
		}

		currentY += lineHeight
	}
}

func (p *GPUPainter) Flush() {
	p.shader.Use()

	if len(p.rectVerts) > 0 {
		p.shader.SetBool("uUseTexture", false)
		p.defaultTexture.Bind(0)
		p.shader.SetInt("uTexture", 0)
		p.batchMesh.Upload(p.rectVerts)
		p.batchMesh.DrawQuads(len(p.rectVerts) / 4)
		p.rectVerts = p.rectVerts[:0]
	}

	if len(p.glyphBatches) > 0 {
		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		p.shader.SetBool("uUseTexture", true)
		p.shader.SetInt("uTexture", 0)
		for texID, verts := range p.glyphBatches {
			if len(verts) == 0 {
				continue
			}
			gl.BindTexture(gl.TEXTURE_2D, texID)
			p.batchMesh.Upload(verts)
			p.batchMesh.DrawQuads(len(verts) / 4)
			delete(p.glyphBatches, texID)
		}
		gl.Disable(gl.BLEND)
	}
}

// SetClip sets the clipping area using scissor test.
func (p *GPUPainter) SetClip(rect layout.Rect) {
	gl.Enable(gl.SCISSOR_TEST)

	scaleX := p.fbWidth / p.windowWidth
	scaleY := p.fbHeight / p.windowHeight
	// OpenGL Scissor accepts bottom-left origin; convert from top-left window coords.
	gl.Scissor(
		int32(rect.X*scaleX),
		int32(p.fbHeight-(rect.Y+rect.Height)*scaleY),
		int32(rect.Width*scaleX),
		int32(rect.Height*scaleY),
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

// scrollOffset returns the accumulated scroll translation from the scroll stack.
func (p *GPUPainter) scrollOffset() (dx, dy float64) {
	for _, s := range p.scrollStack {
		dx -= s.offsetX
		dy -= s.offsetY
	}
	return
}

// BeginScroll activates scroll clipping and translation for a scrollable container.
func (p *GPUPainter) BeginScroll(clipRect layout.Rect, offsetX, offsetY float64) {
	p.Flush()
	p.scrollStack = append(p.scrollStack, scrollState{offsetX: offsetX, offsetY: offsetY, clip: clipRect})
	p.SetClip(clipRect)
}

// EndScroll deactivates scroll clipping for a scrollable container.
func (p *GPUPainter) EndScroll() {
	p.Flush()
	if len(p.scrollStack) > 0 {
		p.scrollStack = p.scrollStack[:len(p.scrollStack)-1]
	}
	if len(p.scrollStack) > 0 {
		p.SetClip(p.scrollStack[len(p.scrollStack)-1].clip)
	} else {
		p.ClearClip()
	}
}

func (p *GPUPainter) Delete() {
	p.shader.Delete()
	p.rectMesh.Delete()
	p.batchMesh.Delete()
	if p.defaultTexture != nil {
		p.defaultTexture.Delete()
	}
}
