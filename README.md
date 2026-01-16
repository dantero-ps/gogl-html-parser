# gogl-html-parser

A lightweight, from-scratch HTML and CSS parser + renderer that uses OpenGL (via modern OpenGL / GLSL) to draw web-like UI directly in a Go application.

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![OpenGL](https://img.shields.io/badge/OpenGL-4.1+-5586A4?style=flat-square&logo=opengl)](https://www.opengl.org/)

## Screenshot

![Demo](screenshots/demo.gif)

## Features

- **HTML5 Parser**: Parses a subset of HTML5 including common elements (`div`, `p`, `h1-h6`, `span`, `img`, `br`, etc.)
- **CSS Parser**: Supports basic CSS properties including:
  - Layout: `display`, `position`, `width`, `height`, `min-width`, `max-width`, `min-height`, `max-height`
  - Spacing: `margin`, `padding` (with shorthand and individual sides)
  - Borders: `border` (with shorthand and individual sides)
  - Colors: `color`, `background-color`
  - Typography: `font-size`, `font-family`, `font-weight`, `font-style`, `text-align`
  - Visual: `opacity`, `cursor`
- **OpenGL Rendering**: Renders UI using modern OpenGL (4.1+) with custom shaders
- **Zero Dependencies on Browser Engines**: No WebKit, Blink, or Servo – pure Go implementation
- **No JavaScript**: Pure static HTML+CSS rendering (no scripting support)
- **Custom Font Rendering**: Dynamic font atlas generation using TTF fonts via `golang.org/x/image/font`
- **Event Handling**: Basic DOM-like event system for mouse clicks and hover states
- **Minimal & Embeddable**: Designed to be integrated into games and tools without Electron/webview overhead

## Installation / Quick Start

```bash
git clone https://github.com/yourusername/gogl-html-parser.git
cd gogl-html-parser
go mod tidy
```

### Dependencies

- Go 1.25 or later
- OpenGL 4.1+ support
- GLFW 3.3+ (for window management)
- C compiler (for CGO dependencies)

On macOS:
```bash
brew install glfw
```

On Linux (Ubuntu/Debian):
```bash
sudo apt-get install libgl1-mesa-dev libglfw3-dev
```

On Windows:
- Install MinGW-w64 or use MSYS2
- Install GLFW from [glfw.org](https://www.glfw.org/download.html)

## Basic Usage

```go
package main

import (
    "goglweb/internal/dom"
    "goglweb/internal/gpu"
    "goglweb/internal/parser/css"
    "goglweb/internal/parser/html"
    
    "github.com/go-gl/gl/v4.1-core/gl"
    "github.com/go-gl/glfw/v3.3/glfw"
)

func main() {
    // Initialize GLFW and OpenGL (see main.go for full setup)
    glfw.Init()
    defer glfw.Terminate()
    
    window, _ := glfw.CreateWindow(1200, 800, "My App", nil, nil)
    window.MakeContextCurrent()
    gl.Init()
    
    // Create GPU painter
    painter, _ := gpu.NewGPUPainter(
        1200, 800,
        "assets/shaders/vertex.glsl",
        "assets/shaders/fragment.glsl",
    )
    defer painter.Delete()
    
    // Parse HTML and CSS
    htmlSource := `
        <div class="container">
            <h1>Hello, World!</h1>
            <p>This is rendered with OpenGL</p>
        </div>
    `
    
    cssSource := `
        .container {
            display: block;
            width: 800px;
            margin: 50px auto;
            padding: 20px;
            background-color: white;
            border: 2px solid #333;
        }
        h1 {
            color: #4a90e2;
            font-size: 32px;
        }
    `
    
    htmlParser := html.NewParser(htmlSource)
    htmlRoot := htmlParser.Parse()
    
    cssParser := css.NewParser(cssSource)
    stylesheet := cssParser.Parse()
    
    // Create renderer
    renderer := dom.NewRenderer(htmlRoot, stylesheet, 1200, 800)
    
    // Main loop
    for !window.ShouldClose() {
        gl.Clear(gl.COLOR_BUFFER_BIT)
        displayList := renderer.GetDisplayList()
        displayList.Execute(painter)
        window.SwapBuffers()
        glfw.PollEvents()
    }
}
```

## Supported HTML/CSS Subset

### HTML Tags

| Tag | Status | Notes |
|-----|--------|-------|
| `div` | ✅ Supported | Block container |
| `p` | ✅ Supported | Paragraph (block) |
| `h1` - `h6` | ✅ Supported | Headings (block) |
| `span` | ✅ Supported | Inline container |
| `img` | ✅ Supported | Self-closing, requires `src` attribute |
| `br` | ✅ Supported | Line break |
| `hr` | ✅ Supported | Horizontal rule |
| `input` | ⚠️ Partial | Parsed but not fully rendered |
| `meta`, `link` | ⚠️ Partial | Parsed but ignored in layout |

### CSS Properties

| Property | Status | Notes |
|----------|--------|-------|
| `display` | ✅ Supported | `block`, `inline`, `none` |
| `width`, `height` | ✅ Supported | Pixels (`px`), percentages (`%`), `auto` |
| `margin` | ✅ Supported | Shorthand and individual sides |
| `padding` | ✅ Supported | Shorthand and individual sides |
| `border` | ✅ Supported | Shorthand and individual sides (width, style, color) |
| `background-color` | ✅ Supported | Hex colors (`#rrggbb`, `#rgb`), named colors |
| `color` | ✅ Supported | Text color |
| `font-size` | ✅ Supported | Pixels, `em`, `rem` |
| `font-family` | ✅ Supported | Font name matching |
| `font-weight` | ✅ Supported | `normal`, `bold` |
| `font-style` | ✅ Supported | `normal`, `italic` |
| `text-align` | ✅ Supported | `left`, `center`, `right` |
| `opacity` | ✅ Supported | 0.0 to 1.0 |
| `cursor` | ✅ Supported | Parsed but not yet applied to system cursor |
| `min-width`, `max-width` | ✅ Supported | Constraint properties |
| `min-height`, `max-height` | ✅ Supported | Constraint properties |
| `position` | ⚠️ Partial | Parsed but only `static` positioning implemented |
| `flexbox` | ❌ Planned | Not yet implemented |
| `grid` | ❌ Planned | Not yet implemented |
| `transform` | ❌ Planned | Not yet implemented |
| `animation` | ❌ Planned | Not yet implemented |

### CSS Selectors

- ✅ Element selectors (`div`, `p`, etc.)
- ✅ Class selectors (`.class-name`)
- ✅ ID selectors (`#id-name`)
- ✅ Descendant selectors (`div p`)
- ✅ Pseudo-classes (`:hover`) - basic support
- ❌ Attribute selectors - planned
- ❌ Pseudo-elements (`::before`, `::after`) - planned

## Roadmap / Planned Features

- [ ] **Flexbox Layout**: Implement CSS flexbox for more flexible layouts
- [ ] **Positioning**: Full support for `position: absolute`, `relative`, `fixed`
- [ ] **More CSS Properties**: `box-shadow`, `border-radius`, `overflow`, `z-index`
- [ ] **Text Rendering Improvements**: Better line breaking, text wrapping, multi-line text
- [ ] **Image Loading**: Load and render images from files or URLs
- [ ] **CSS Transitions**: Basic animation support for property changes
- [ ] **More HTML Elements**: `button`, `input` (fully functional), `textarea`, `select`
- [ ] **CSS Grid**: Grid layout system
- [ ] **Media Queries**: Responsive design support
- [ ] **Performance Optimizations**: Dirty rectangle tracking, partial re-rendering

## Limitations

- **Early Development**: This is a proof-of-concept. Many CSS properties and layout features are missing or partially implemented.
- **No JavaScript**: This is intentional – the project focuses on static HTML+CSS rendering only.
- **Limited Text Layout**: Text wrapping and line breaking are basic. Complex text layouts may not render correctly.
- **No Network Requests**: Cannot load external resources (images, fonts) from URLs. All assets must be local.
- **Basic Event System**: Only mouse click and hover events are supported. Keyboard events are parsed but not fully integrated.
- **No CSS Preprocessing**: No support for CSS variables, `@import`, or `@media` queries yet.
- **Performance**: No optimization for large DOM trees or frequent re-renders. Suitable for simple UIs but may struggle with complex layouts.

## Building & Running

### Prerequisites

1. **Go 1.25+**: [Download Go](https://golang.org/dl/)
2. **OpenGL 4.1+**: Usually provided by your graphics drivers
3. **GLFW 3.3+**: Window and input management
4. **C Compiler**: Required for CGO (used by Go GL bindings)

### Build

```bash
go build -o renderer ./cmd/renderer
```

### Run

```bash
./renderer
```

Or directly with `go run`:

```bash
go run ./cmd/renderer/main.go
```

### Project Structure

```
goglweb/
├── cmd/
│   └── renderer/          # Main application entry point
├── internal/
│   ├── parser/
│   │   ├── html/          # HTML tokenizer and parser
│   │   └── css/           # CSS tokenizer and parser
│   ├── style/             # Style computation and cascade
│   ├── layout/            # Layout engine (block/inline)
│   ├── render/            # Display list generation
│   ├── gpu/               # OpenGL rendering (shaders, textures, fonts)
│   └── dom/               # DOM manipulation and events
└── assets/
    └── shaders/           # GLSL vertex and fragment shaders
```

## License

MIT License. See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! This project is in early development, so there's plenty of room for improvement.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Areas That Need Help

- CSS property implementations
- Layout algorithm improvements
- Text rendering enhancements
- Performance optimizations
- Documentation and examples
- Test coverage

## Author / Contact

Created for in-game/user interfaces in custom OpenGL applications, especially game engines or Minecraft-like voxel games.

For questions, issues, or contributions, please open an issue on GitHub.

---

**Note**: This project is a learning exercise and proof-of-concept. It's not intended to replace full-featured browser engines, but rather to provide a lightweight alternative for applications that need simple HTML/CSS-like UI rendering in OpenGL contexts.
