package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/furkandgn/goglweb"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	html := `
<div class="page">
  <h1>Interaction Demo</h1>

  <h2>Click to toggle</h2>
  <div class="zone">
    <div class="btn" id="btn1">Click me — toggles .active</div>
    <div class="btn" id="btn2">Click me too</div>
  </div>

  <h2>Hover feedback (automatic)</h2>
  <div class="zone">
    <div class="hoverable" id="hov1">Hover over me</div>
    <div class="hoverable" id="hov2">And me</div>
  </div>

  <h2>Key press — press Space to add a row</h2>
  <div class="zone">
    <div class="key-hint">Focus the window and press Space</div>
    <div id="list-container"></div>
  </div>

  <div class="status" id="status">No interaction yet</div>
</div>`

	css := `
* { margin: 0; padding: 0; }
.page {
  display: block;
  width: 700px;
  margin: 24px auto;
  padding: 20px;
  background-color: #ffffff;
  border: 2px solid #cccccc;
}
h1 {
  display: block;
  font-size: 22px;
  font-weight: bold;
  padding: 10px 14px;
  margin: 0 0 14px 0;
  background-color: #e17055;
  color: #ffffff;
  border: 1px solid #c0392b;
}
h2 {
  display: block;
  font-size: 14px;
  font-weight: bold;
  color: #555555;
  margin: 14px 0 6px 0;
}
.zone {
  display: block;
  padding: 10px;
  margin: 0 0 12px 0;
  background-color: #f8f8f8;
  border: 1px solid #e0e0e0;
}
.btn {
  display: block;
  width: 280px;
  padding: 10px 14px;
  margin: 6px 0;
  background-color: #74b9ff;
  color: #ffffff;
  font-size: 13px;
  font-weight: bold;
  border: 2px solid #0984e3;
  cursor: pointer;
}
.btn.hover {
  background-color: #0984e3;
}
.btn.active {
  background-color: #6c5ce7;
  border: 2px solid #4a3ab0;
}
.hoverable {
  display: block;
  width: 260px;
  padding: 10px 14px;
  margin: 6px 0;
  background-color: #55efc4;
  color: #2d3436;
  font-size: 13px;
  border: 2px solid #00b894;
  cursor: pointer;
}
.hoverable.hover {
  background-color: #00b894;
  color: #ffffff;
}
.key-hint {
  display: block;
  padding: 8px 12px;
  margin: 0 0 6px 0;
  background-color: #ffeaa7;
  color: #2d3436;
  font-size: 12px;
  border: 1px solid #fdcb6e;
}
.list-item {
  display: block;
  padding: 6px 12px;
  margin: 3px 0;
  background-color: #dfe6e9;
  color: #2d3436;
  font-size: 12px;
  border: 1px solid #b2bec3;
}
.status {
  display: block;
  margin: 12px 0 0 0;
  padding: 8px 12px;
  background-color: #2d3436;
  color: #dfe6e9;
  font-size: 12px;
  border: 1px solid #636e72;
}
`

	app, err := goglweb.New(html, css,
		goglweb.WithTitle("goglweb — Interaction"),
		goglweb.WithViewport(760, 640),
	)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	statusNode := app.FindNodeByID("status")
	listContainer := app.FindNodeByID("list-container")
	itemCount := 0

	// OnClick: toggle .active class on clicked buttons
	app.OnClick(func(x, y float64, target goglweb.NodeRef) {
		id := app.GetAttribute(target, "id")
		if id == "btn1" || id == "btn2" {
			app.ToggleClass(target, "active")
			app.MarkDirty()
			msg := fmt.Sprintf("Clicked %s — toggled .active", id)
			app.SetTextContent(statusNode, msg)
			app.MarkDirty()
		}
	})

	// OnKey: press Space to append a new list item
	app.OnKey(func(key string, mods goglweb.ModifierKey) {
		if key == "space" {
			itemCount++
			item := app.NewElement("div")
			app.AddClass(item, "list-item")
			text := app.NewText(fmt.Sprintf("Item %d added via AppendChild", itemCount))
			app.AppendChild(item, text)
			app.AppendChild(listContainer, item)
			app.MarkDirty()
			app.SetTextContent(statusNode, fmt.Sprintf("Space pressed — appended item %d", itemCount))
			app.MarkDirty()
		}
	})

	if err := app.Run(); err != nil {
		log.Fatalf("App error: %v", err)
	}
}
