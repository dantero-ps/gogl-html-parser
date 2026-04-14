package goglweb

import "os"

// config holds renderer configuration (internal).
type config struct {
	viewportWidth  float64
	viewportHeight float64
	title          string
	vertexShader   string // source code
	fragmentShader string // source code
	fontDirs       []string
}

// Option is a functional option for configuring the renderer.
type Option func(*config)

func defaultConfig() config {
	return config{
		viewportWidth:  1200,
		viewportHeight: 800,
		title:          "goglweb",
		vertexShader:   defaultVertexShader,
		fragmentShader: defaultFragmentShader,
	}
}

// WithViewport sets the viewport dimensions.
func WithViewport(width, height int) Option {
	return func(c *config) {
		c.viewportWidth = float64(width)
		c.viewportHeight = float64(height)
	}
}

// WithTitle sets the window title.
func WithTitle(title string) Option {
	return func(c *config) {
		c.title = title
	}
}

// WithShaders loads custom GLSL shaders from file paths.
func WithShaders(vertexPath, fragmentPath string) Option {
	return func(c *config) {
		vSrc, err := os.ReadFile(vertexPath)
		if err == nil {
			c.vertexShader = string(vSrc)
		}
		fSrc, err := os.ReadFile(fragmentPath)
		if err == nil {
			c.fragmentShader = string(fSrc)
		}
	}
}

// WithFonts adds additional font search directories.
func WithFonts(paths ...string) Option {
	return func(c *config) {
		c.fontDirs = append(c.fontDirs, paths...)
	}
}
