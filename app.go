package goglweb

import (
	"fmt"
	"runtime"
	"time"

	"github.com/furkandgn/goglweb/internal/dom"
	"github.com/furkandgn/goglweb/internal/gpu"
	"github.com/furkandgn/goglweb/internal/layout"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// ModifierKey represents keyboard modifier keys (shift, ctrl, alt, super).
type ModifierKey int

const (
	ModShift ModifierKey = 1 << iota
	ModControl
	ModAlt
	ModSuper
)

func glfwModsToModifierKey(mods glfw.ModifierKey) ModifierKey {
	var result ModifierKey
	if mods&glfw.ModShift != 0 {
		result |= ModShift
	}
	if mods&glfw.ModControl != 0 {
		result |= ModControl
	}
	if mods&glfw.ModAlt != 0 {
		result |= ModAlt
	}
	if mods&glfw.ModSuper != 0 {
		result |= ModSuper
	}
	return result
}

// App is the high-level rendering application. It owns the GLFW window,
// OpenGL context, and event loop. Most users only need App.
type App struct {
	*Renderer // Embedded: DOM manipulation and pipeline control inherited

	window   *glfw.Window
	painter  *gpu.GPUPainter
	eventMgr *dom.EventManager
	cfg      config

	// User-registered callbacks
	onClick  func(x, y float64, target NodeRef)
	onKey    func(key string, mods ModifierKey)
	onScroll func(xoff, yoff float64)
	onHover  func(x, y float64, target NodeRef)

	// Internal hover tracking
	lastHoverNodeRef NodeRef
}

// New creates a high-level App with its own GLFW window and OpenGL context.
// This is the primary entry point for most users.
// Must be called from the main goroutine (OpenGL requirement).
func New(htmlSource, cssSource string, opts ...Option) (*App, error) {
	runtime.LockOSThread()

	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize GLFW: %w", err)
	}

	// OpenGL 4.1 Core Profile
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	window, err := glfw.CreateWindow(
		int(cfg.viewportWidth), int(cfg.viewportHeight), cfg.title, nil, nil,
	)
	if err != nil {
		glfw.Terminate()
		return nil, fmt.Errorf("failed to create window: %w", err)
	}
	window.MakeContextCurrent()
	glfw.SwapInterval(1) // Vsync

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		window.Destroy()
		glfw.Terminate()
		return nil, fmt.Errorf("failed to initialize OpenGL: %w", err)
	}

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(0.9, 0.9, 0.9, 1.0)

	// Get actual sizes (for Retina)
	actualWinWidth, actualWinHeight := window.GetSize()
	fbWidth, fbHeight := window.GetFramebufferSize()

	// Create GPU Painter with embedded shaders
	painter, err := gpu.NewGPUPainterFromSource(
		float64(actualWinWidth), float64(actualWinHeight),
		cfg.vertexShader, cfg.fragmentShader,
	)
	if err != nil {
		window.Destroy()
		glfw.Terminate()
		return nil, fmt.Errorf("failed to create GPU painter: %w", err)
	}

	// Set extra font paths if provided
	if len(cfg.fontDirs) > 0 {
		painter.SetExtraFontPaths(cfg.fontDirs)
	}

	gl.Viewport(0, 0, int32(fbWidth), int32(fbHeight))
	painter.SetFramebufferSize(float64(fbWidth), float64(fbHeight))

	// Create the low-level renderer
	renderer, err := NewRenderer(htmlSource, cssSource, opts...)
	if err != nil {
		painter.Delete()
		window.Destroy()
		glfw.Terminate()
		return nil, err
	}

	// Wire real font metrics and force re-layout so inline positions use real glyph widths.
	// NewRenderer builds an initial layout with FallbackMeasurer before this point.
	renderer.renderer.LayoutCtx.TextMeasurer = gpu.NewTextMeasurerAdapter(painter)
	renderer.renderer.MarkDirty()

	// Create event manager
	eventMgr := dom.NewEventManager()

	app := &App{
		Renderer: renderer,
		window:   window,
		painter:  painter,
		eventMgr: eventMgr,
		cfg:      cfg,
	}

	// Wire up GLFW callbacks
	app.setupCallbacks()

	return app, nil
}

// renderFrame clears, repaints, and swaps buffers. Safe to call from callbacks.
func (a *App) renderFrame() {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	dl := a.Renderer.renderer.GetDisplayList()
	dl.Execute(a.painter)
	a.painter.Flush()
	a.window.SwapBuffers()
	a.ClearDirty()
}

// Run starts the blocking event loop. Handles window events, rendering,
// and vsync. Returns when the window is closed or Close() is called.
func (a *App) Run() error {
	defer a.Close()

	frameCount := 0
	lastFPSTime := time.Now()

	for !a.window.ShouldClose() {
		glfw.PollEvents()
		if a.IsDirty() {
			a.renderFrame()

			frameCount++
			now := time.Now()
			if now.Sub(lastFPSTime) >= time.Second {
				frameCount = 0
				lastFPSTime = now
			}
		}
	}
	return nil
}

// Close releases all GPU and GLFW resources. Called automatically when Run() exits.
// Safe to call multiple times.
func (a *App) Close() {
	if a.painter != nil {
		a.painter.Delete()
		a.painter = nil
	}
	if a.window != nil {
		a.window.Destroy()
		a.window = nil
	}
	glfw.Terminate()
}

// OnClick registers a callback for mouse click events.
func (a *App) OnClick(cb func(x, y float64, target NodeRef)) { a.onClick = cb }

// OnKey registers a callback for keyboard events.
func (a *App) OnKey(cb func(key string, mods ModifierKey)) { a.onKey = cb }

// OnScroll registers a callback for mouse scroll events.
func (a *App) OnScroll(cb func(xoff, yoff float64)) { a.onScroll = cb }

// OnHover registers a callback for mouse hover (enter/leave) events.
func (a *App) OnHover(cb func(x, y float64, target NodeRef)) { a.onHover = cb }

// setupCallbacks wires GLFW window callbacks to the App's internal handling
// and user-registered callbacks.
func (a *App) setupCallbacks() {
	// Framebuffer size callback — always fires after SetSizeCallback and has the
	// correct physical pixel dimensions. We do the single authoritative render here
	// so we never render twice per resize step (which causes flicker).
	a.window.SetFramebufferSizeCallback(func(w *glfw.Window, fbWidth, fbHeight int) {
		gl.Viewport(0, 0, int32(fbWidth), int32(fbHeight))
		a.painter.SetFramebufferSize(float64(fbWidth), float64(fbHeight))
		winW, winH := w.GetSize()
		a.painter.SetViewport(float64(winW), float64(winH))
		a.Renderer.UpdateViewport(float64(winW), float64(winH))
		a.renderFrame()
	})

	// Window size callback — just update state; rendering is handled by
	// SetFramebufferSizeCallback which always follows this on every platform.
	a.window.SetSizeCallback(func(w *glfw.Window, width, height int) {
		a.painter.SetViewport(float64(width), float64(height))
		a.Renderer.UpdateViewport(float64(width), float64(height))
	})

	// Window refresh callback — called by the OS when the window needs repainting
	// (e.g. uncovered after being hidden). Mark dirty; the main loop or the
	// FramebufferSizeCallback (which fires after this during resize) will render.
	// Do NOT call renderFrame() here: during resize macOS fires both callbacks on
	// every step, causing double renders and visible flicker.
	a.window.SetRefreshCallback(func(w *glfw.Window) {
		a.MarkDirty()
	})

	// Mouse button callback
	a.window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		if button == glfw.MouseButtonLeft && action == glfw.Press {
			x, y := w.GetCursorPos()

			layoutTree := a.Renderer.renderer.GetLayoutTree()
			if layoutTree == nil {
				return
			}

			hitResult := dom.HitTest(layoutTree, x, y)
			var targetNodeRef NodeRef

			if hitResult != nil && hitResult.HTMLNode != nil {
				targetNodeRef = NodeRef{node: hitResult.HTMLNode}

				// Dispatch to internal event manager for bubbling
				a.eventMgr.DispatchEvent(dom.Event{
					Type:   dom.EventClick,
					Target: hitResult.HTMLNode,
					X:      x,
					Y:      y,
				}, a.Renderer.htmlRoot)
			}

			// Fire user callback
			if a.onClick != nil {
				a.onClick(x, y, targetNodeRef)
			}
		}
	})

	// Cursor position callback (hover tracking)
	a.window.SetCursorPosCallback(func(w *glfw.Window, x, y float64) {
		layoutTree := a.Renderer.renderer.GetLayoutTree()
		if layoutTree == nil {
			return
		}

		var currentHoverNodeRef NodeRef
		hitResult := dom.HitTest(layoutTree, x, y)
		if hitResult != nil && hitResult.HTMLNode != nil {
			currentHoverNodeRef = NodeRef{node: hitResult.HTMLNode}
		}

		// Check if hover changed
		if currentHoverNodeRef != a.lastHoverNodeRef {
			// Exit old hover
			if a.lastHoverNodeRef.node != nil {
				dom.RemoveClass(a.lastHoverNodeRef.node, "hover")
				a.eventMgr.DispatchEvent(dom.Event{
					Type:   dom.EventHover,
					Target: a.lastHoverNodeRef.node,
					X:      x,
					Y:      y,
				}, a.Renderer.htmlRoot)
				a.MarkDirty()
			}
			// Enter new hover
			if currentHoverNodeRef.node != nil {
				dom.AddClass(currentHoverNodeRef.node, "hover")
				a.eventMgr.DispatchEvent(dom.Event{
					Type:   dom.EventHover,
					Target: currentHoverNodeRef.node,
					X:      x,
					Y:      y,
				}, a.Renderer.htmlRoot)
				a.MarkDirty()
			}
			a.lastHoverNodeRef = currentHoverNodeRef

			// Fire user callback
			if a.onHover != nil {
				a.onHover(x, y, currentHoverNodeRef)
			}
		}
	})

	// Scroll callback
	a.window.SetScrollCallback(func(w *glfw.Window, xoff, yoff float64) {
		x, y := w.GetCursorPos()
		layoutTree := a.Renderer.renderer.GetLayoutTree()
		if layoutTree == nil {
			return
		}

		target := findInnermostScrollable(layoutTree, x, y)
		if target != nil {
			const scrollSpeed = 20.0
			target.ScrollOffsetY -= yoff * scrollSpeed
			target.ScrollOffsetY = layout.ClampScrollOffset(target.ScrollOffsetY, target.ChildrenHeight, target.Dimensions.Content.Height)
			a.MarkDirty()
		}

		if a.onScroll != nil {
			a.onScroll(xoff, yoff)
		}
	})

	// Keyboard callback
	a.window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press || action == glfw.Repeat {
			if a.onKey != nil {
				a.onKey(glfwKeyName(key), glfwModsToModifierKey(mods))
			}
		}
	})
}

// findInnermostScrollable returns the innermost scrollable LayoutBox under the cursor.
func findInnermostScrollable(root *layout.LayoutBox, x, y float64) *layout.LayoutBox {
	if root == nil {
		return nil
	}
	bb := root.Dimensions.BorderBox()
	if x < bb.X || x > bb.X+bb.Width || y < bb.Y || y > bb.Y+bb.Height {
		return nil
	}
	// Depth-first: prefer innermost child.
	for _, child := range root.Children {
		if found := findInnermostScrollable(child, x, y); found != nil {
			return found
		}
	}
	if root.IsScrollable() {
		return root
	}
	return nil
}

// glfwKeyName converts a glfw.Key to a human-readable string.
func glfwKeyName(key glfw.Key) string {
	switch key {
	case glfw.KeyUnknown:
		return "unknown"
	case glfw.KeySpace:
		return "space"
	case glfw.KeyEscape:
		return "escape"
	case glfw.KeyEnter:
		return "enter"
	case glfw.KeyTab:
		return "tab"
	case glfw.KeyBackspace:
		return "backspace"
	case glfw.KeyInsert:
		return "insert"
	case glfw.KeyDelete:
		return "delete"
	case glfw.KeyRight:
		return "right"
	case glfw.KeyLeft:
		return "left"
	case glfw.KeyDown:
		return "down"
	case glfw.KeyUp:
		return "up"
	case glfw.KeyPageUp:
		return "page_up"
	case glfw.KeyPageDown:
		return "page_down"
	case glfw.KeyHome:
		return "home"
	case glfw.KeyEnd:
		return "end"
	case glfw.KeyCapsLock:
		return "caps_lock"
	case glfw.KeyScrollLock:
		return "scroll_lock"
	case glfw.KeyNumLock:
		return "num_lock"
	case glfw.KeyPrintScreen:
		return "print_screen"
	case glfw.KeyPause:
		return "pause"
	case glfw.KeyLeftShift:
		return "left_shift"
	case glfw.KeyLeftControl:
		return "left_control"
	case glfw.KeyLeftAlt:
		return "left_alt"
	case glfw.KeyLeftSuper:
		return "left_super"
	case glfw.KeyRightShift:
		return "right_shift"
	case glfw.KeyRightControl:
		return "right_control"
	case glfw.KeyRightAlt:
		return "right_alt"
	case glfw.KeyRightSuper:
		return "right_super"
	case glfw.KeyF1:
		return "f1"
	case glfw.KeyF2:
		return "f2"
	case glfw.KeyF3:
		return "f3"
	case glfw.KeyF4:
		return "f4"
	case glfw.KeyF5:
		return "f5"
	case glfw.KeyF6:
		return "f6"
	case glfw.KeyF7:
		return "f7"
	case glfw.KeyF8:
		return "f8"
	case glfw.KeyF9:
		return "f9"
	case glfw.KeyF10:
		return "f10"
	case glfw.KeyF11:
		return "f11"
	case glfw.KeyF12:
		return "f12"
	case glfw.KeyKP0:
		return "numpad_0"
	case glfw.KeyKP1:
		return "numpad_1"
	case glfw.KeyKP2:
		return "numpad_2"
	case glfw.KeyKP3:
		return "numpad_3"
	case glfw.KeyKP4:
		return "numpad_4"
	case glfw.KeyKP5:
		return "numpad_5"
	case glfw.KeyKP6:
		return "numpad_6"
	case glfw.KeyKP7:
		return "numpad_7"
	case glfw.KeyKP8:
		return "numpad_8"
	case glfw.KeyKP9:
		return "numpad_9"
	default:
		// Printable ASCII range: A-Z, 0-9
		if key >= glfw.KeyA && key <= glfw.KeyZ {
			return string(rune('a' + (key - glfw.KeyA)))
		}
		if key >= glfw.Key0 && key <= glfw.Key9 {
			return string(rune('0' + (key - glfw.Key0)))
		}
		return fmt.Sprintf("key_%d", key)
	}
}
