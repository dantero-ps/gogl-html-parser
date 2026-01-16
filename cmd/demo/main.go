package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"goglweb/internal/dom"
	"goglweb/internal/gpu"
	"goglweb/internal/layout"
	"goglweb/internal/parser/css"
	"goglweb/internal/parser/html"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

func init() {
	runtime.LockOSThread()
}

// isPointInLayoutBox checks if a point is inside a layout box
func isPointInLayoutBox(box *layout.LayoutBox, x, y float64) bool {
	if box == nil {
		return false
	}
	borderBox := box.Dimensions.BorderBox()
	return x >= borderBox.X &&
		x <= borderBox.X+borderBox.Width &&
		y >= borderBox.Y &&
		y <= borderBox.Y+borderBox.Height
}

func renderScene(window *glfw.Window, renderer *dom.Renderer, painter *gpu.GPUPainter) {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	displayList := renderer.GetDisplayList()
	displayList.Execute(painter)
	window.SwapBuffers()
}

func main() {
	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		log.Fatalf("Failed to initialize GLFW: %v", err)
	}
	defer glfw.Terminate()

	// OpenGL 4.1 Core Profile
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// Create window
	winWidth, winHeight := 1200, 800
	window, err := glfw.CreateWindow(winWidth, winHeight, "GoGL HTML Demo", nil, nil)
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}
	defer window.Destroy()

	window.MakeContextCurrent()

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		log.Fatalf("Failed to initialize OpenGL: %v", err)
	}

	// OpenGL ayarları
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(0.9, 0.9, 0.9, 1.0) // Light gray background

	// Get window size and framebuffer size (for retina display support)
	actualWinWidth, actualWinHeight := window.GetSize()
	fbWidth, fbHeight := window.GetFramebufferSize()

	// Create GPU Painter
	painter, err := gpu.NewGPUPainter(
		float64(actualWinWidth),
		float64(actualWinHeight),
		"assets/shaders/vertex.glsl",
		"assets/shaders/fragment.glsl",
	)
	if err != nil {
		log.Fatalf("Failed to create GPU Painter: %v", err)
	}
	defer painter.Delete()

	// Set initial viewport
	gl.Viewport(0, 0, int32(fbWidth), int32(fbHeight))

	// HTML ve CSS parse et
	htmlSource := `
       <div class="container">
          <div class="box red" id="box1">Red Box</div>
          <div class="box blue" id="box2">Blue Box</div>
          <div class="box green" id="box3">Green Box</div>
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
          display: block;
          margin: 0 0 20px 0;
          padding: 10px;
          background-color: #4a90e2;
          color: white;
          border: 1px solid #2a6bb0;
       }
       p {
          display: block;
          margin: 0 0 40px 0;
          padding: 10px;
          background-color: #f0f0f0;
          border: 1px solid #ccc;
       }
       .box {
          display: block;
          width: 200px;
          height: 100px;
          margin: 10px;
          padding: 15px;
          border: 3px solid #000;
          cursor: pointer;
       }
       .red {
          background-color: #ff6b6b;
       }
       .blue {
          background-color: #4ecdc4;
       }
       .green {
          background-color: #95e1d3;
       }
       .box:hover {
          opacity: 0.8;
       }
       .box.hover {
          opacity: 0.8;
       }
       .box.clicked {
          border: 5px solid #ff0000;
       }
    `

	// Parse operations
	htmlParser := html.NewParser(htmlSource)
	htmlRoot := htmlParser.Parse()

	cssParser := css.NewParser(cssSource)
	stylesheet := cssParser.Parse()

	// Create renderer
	renderer := dom.NewRenderer(htmlRoot, stylesheet, float64(actualWinWidth), float64(actualWinHeight))

	// Create event manager
	eventManager := dom.NewEventManager()

	// Example: Add click event listener to boxes
	var findNodeByID func(root *html.Node, id string) *html.Node
	findNodeByID = func(root *html.Node, id string) *html.Node {
		if root == nil {
			return nil
		}
		if dom.GetAttribute(root, "id") == id {
			return root
		}
		for _, child := range root.Children {
			if result := findNodeByID(child, id); result != nil {
				return result
			}
		}
		return nil
	}

	// Add click listener to boxes - use e.Target to avoid closure problem
	eventManager.AddEventListener(htmlRoot, dom.EventClick, func(e dom.Event) {
		if e.Target != nil {
			boxID := dom.GetAttribute(e.Target, "id")
			if boxID != "" && len(boxID) >= 3 && boxID[:3] == "box" {
				fmt.Printf("Box %s clicked!\n", boxID)
				// Toggle 'clicked' class on the clicked box
				dom.ToggleClass(e.Target, "clicked")
				// DOM changed, re-render needed
				renderer.MarkDirty()
			}
		}
	})

	// Framebuffer size callback
	window.SetFramebufferSizeCallback(func(w *glfw.Window, fbWidth, fbHeight int) {
		gl.Viewport(0, 0, int32(fbWidth), int32(fbHeight))
		winW, winH := w.GetSize()
		painter.SetViewport(float64(winW), float64(winH))
		renderer.UpdateViewport(float64(winW), float64(winH))
	})

	// Window size callback
	window.SetSizeCallback(func(w *glfw.Window, width, height int) {
		painter.SetViewport(float64(width), float64(height))
		renderer.UpdateViewport(float64(width), float64(height))
	})

	// Refresh callback
	window.SetRefreshCallback(func(w *glfw.Window) {
		renderScene(w, renderer, painter)
	})

	// Mouse click callback
	window.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		if button == glfw.MouseButtonLeft && action == glfw.Press {
			x, y := w.GetCursorPos()
			// Invert Y coordinate (GLFW is top-down, our system is bottom-up)
			// _, winH := w.GetSize()
			// y = float64(winH) - y  // This line was removed

			// Perform hit test
			layoutTree := renderer.GetLayoutTree()
			if layoutTree != nil {
				hitResult := dom.HitTest(layoutTree, x, y)
				if hitResult != nil && hitResult.HTMLNode != nil {
					targetNode := hitResult.HTMLNode

					// Debug: which node was clicked
					nodeID := dom.GetAttribute(targetNode, "id")
					fmt.Printf("Hit test: tag=%s, id=%s, pos=(%.0f, %.0f)\n", targetNode.TagName, nodeID, x, y)

					// Dispatch event
					eventManager.DispatchEvent(dom.Event{
						Type:   dom.EventClick,
						Target: targetNode,
						X:      x,
						Y:      y,
					}, htmlRoot)
				} else {
					// If text node was clicked, find parent element
					// Check boxes containing text nodes in layout tree
					var findBoxForTextNode func(root *html.Node) *html.Node
					findBoxForTextNode = func(root *html.Node) *html.Node {
						if root == nil {
							return nil
						}
						for _, child := range root.Children {
							if child.Type == html.ElementNode {
								classes := dom.GetClasses(child)
								hasBox := false
								for _, c := range classes {
									if c == "box" {
										hasBox = true
										break
									}
								}
								if hasBox {
									// Check if this box contains a text node
									for _, textChild := range child.Children {
										if textChild.Type == html.TextNode {
											// Check this text node's layout box
											layoutBox := dom.FindLayoutBoxByNode(layoutTree, textChild)
											if layoutBox != nil && isPointInLayoutBox(layoutBox, x, y) {
												return child
											}
										}
									}
								}
							}
							if result := findBoxForTextNode(child); result != nil {
								return result
							}
						}
						return nil
					}

					boxNode := findBoxForTextNode(htmlRoot)
					if boxNode != nil {
						nodeID := dom.GetAttribute(boxNode, "id")
						fmt.Printf("Hit test (text node parent): tag=%s, id=%s, pos=(%.0f, %.0f)\n", boxNode.TagName, nodeID, x, y)
						eventManager.DispatchEvent(dom.Event{
							Type:   dom.EventClick,
							Target: boxNode,
							X:      x,
							Y:      y,
						}, htmlRoot)
					} else {
						fmt.Printf("Hit test: nothing found, pos=(%.0f, %.0f)\n", x, y)
					}
				}
			}
		}
	})

	// Mouse move callback (for hover)
	var lastHoverNode *html.Node
	window.SetCursorPosCallback(func(w *glfw.Window, x, y float64) {
		// Invert Y coordinate
		// _, winH := w.GetSize()
		// y = float64(winH) - y  // This line was removed

		// Perform hit test
		var currentHoverNode *html.Node
		layoutTree := renderer.GetLayoutTree()
		if layoutTree != nil {
			hitResult := dom.HitTest(layoutTree, x, y)
			if hitResult != nil && hitResult.HTMLNode != nil {
				currentHoverNode = hitResult.HTMLNode
			}
		}

		// Check if hover changed
		if currentHoverNode != lastHoverNode {
			// Exit old hover
			if lastHoverNode != nil {
				dom.RemoveClass(lastHoverNode, "hover")
				eventManager.DispatchEvent(dom.Event{
					Type:   dom.EventHover,
					Target: lastHoverNode,
					X:      x,
					Y:      y,
				}, htmlRoot)
				renderer.MarkDirty()
			}
			// Enter new hover
			if currentHoverNode != nil {
				boxID := dom.GetAttribute(currentHoverNode, "id")
				if boxID != "" && len(boxID) >= 3 && boxID[:3] == "box" {
					dom.AddClass(currentHoverNode, "hover")
					renderer.MarkDirty()
				}
				eventManager.DispatchEvent(dom.Event{
					Type:   dom.EventHover,
					Target: currentHoverNode,
					X:      x,
					Y:      y,
				}, htmlRoot)
			}
			lastHoverNode = currentHoverNode
		}
	})

	// Keyboard callback
	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press || action == glfw.Repeat {
			// Example: Add a new box when 'A' key is pressed
			if key == glfw.KeyA {
				container := findNodeByID(htmlRoot, "")
				// Find container (first div)
				for _, child := range htmlRoot.Children {
					if child.TagName == "div" && dom.GetAttribute(child, "class") == "container" {
						container = child
						break
					}
				}
				if container != nil {
					newBox := html.NewElement("div")
					boxID := fmt.Sprintf("box%d", len(container.Children)+1)
					dom.SetAttribute(newBox, "class", "box red")
					dom.SetAttribute(newBox, "id", boxID)
					textNode := html.NewText("New Box")
					dom.AppendChild(newBox, textNode)
					dom.AppendChild(container, newBox)
					// No need to add event listener to new box because we added it to root
					renderer.MarkDirty()
					fmt.Printf("New box added! (ID: %s)\n", boxID)
				}
			}
			// Delete last box when 'D' key is pressed
			if key == glfw.KeyD {
				container := findNodeByID(htmlRoot, "")
				for _, child := range htmlRoot.Children {
					if child.TagName == "div" && dom.GetAttribute(child, "class") == "container" {
						container = child
						break
					}
				}
				if container != nil && len(container.Children) > 0 {
					lastChild := container.Children[len(container.Children)-1]
					dom.RemoveChild(container, lastChild)
					renderer.MarkDirty()
					fmt.Println("Last box deleted!")
				}
			}
		}
	})

	// FPS sayacı
	frameCount := 0
	lastFPSTime := time.Now()

	// Main loop
	for !window.ShouldClose() {
		renderScene(window, renderer, painter)
		glfw.PollEvents()

		// FPS sayacı
		frameCount++
		now := time.Now()
		if now.Sub(lastFPSTime) >= time.Second {
			fps := float64(frameCount) / now.Sub(lastFPSTime).Seconds()
			fmt.Printf("FPS: %.2f\n", fps)
			frameCount = 0
			lastFPSTime = now
		}
	}

	fmt.Println("Program terminated.")
}
