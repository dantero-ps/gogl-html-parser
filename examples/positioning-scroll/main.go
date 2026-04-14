package main

import (
	"log"
	"runtime"

	"github.com/furkandgn/goglweb"
)

func init() {
	runtime.LockOSThread()
}

func main() {
	html := `
<div class="fixed-bar">Fixed Header — stays in place while scrolling</div>
<div class="page">
  <h1>Positioning &amp; Scroll</h1>

  <h2>Relative + Absolute positioning</h2>
  <div class="rel-parent">
    <p>Relatively positioned parent (outlined)</p>
    <div class="abs-child">Absolute child — top:10px left:180px</div>
  </div>

  <h2>Scrollable container (overflow: auto)</h2>
  <div class="scroll-box">
    <div class="scroll-content">
      <p>Line 1 — scroll down to see more</p>
      <p>Line 2</p>
      <p>Line 3</p>
      <p>Line 4</p>
      <p>Line 5</p>
      <p>Line 6</p>
      <p>Line 7</p>
      <p>Line 8</p>
      <p>Line 9</p>
      <p>Line 10 — end of scrollable content</p>
    </div>
  </div>
</div>`

	css := `
* { margin: 0; padding: 0; }
.fixed-bar {
  position: fixed;
  top: 0;
  left: 0;
  width: 800px;
  padding: 10px 16px;
  background-color: #2d3436;
  color: #ffffff;
  font-size: 13px;
  font-weight: bold;
  border-width: 0 0 2px 0;
  border-style: solid;
  border-color: #636e72;
}
.page {
  display: block;
  width: 740px;
  margin: 50px auto 20px auto;
  padding: 20px;
  background-color: #ffffff;
  border: 2px solid #cccccc;
}
h1 {
  display: block;
  font-size: 22px;
  font-weight: bold;
  padding: 10px 14px;
  margin: 0 0 16px 0;
  background-color: #00cec9;
  color: #ffffff;
  border: 1px solid #00a8a3;
}
h2 {
  display: block;
  font-size: 14px;
  font-weight: bold;
  color: #555555;
  margin: 16px 0 8px 0;
}
p { display: block; margin: 0 0 6px 0; font-size: 13px; color: #333333; }
.rel-parent {
  position: relative;
  height: 80px;
  margin: 0 0 20px 0;
  padding: 10px;
  background-color: #ffeaa7;
  border: 2px dashed #fdcb6e;
}
.abs-child {
  position: absolute;
  top: 10px;
  left: 180px;
  padding: 6px 10px;
  background-color: #fd79a8;
  color: #ffffff;
  font-size: 12px;
  border: 1px solid #e84393;
}
.scroll-box {
  display: block;
  width: 680px;
  height: 140px;
  overflow: auto;
  border: 2px solid #b2bec3;
  background-color: #f8f9fa;
  padding: 8px;
  margin: 0 0 16px 0;
}
.scroll-content {
  display: block;
  padding: 4px;
}
`

	app, err := goglweb.New(html, css,
		goglweb.WithTitle("goglweb — Positioning & Scroll"),
		goglweb.WithViewport(820, 600),
	)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}
	if err := app.Run(); err != nil {
		log.Fatalf("App error: %v", err)
	}
}
