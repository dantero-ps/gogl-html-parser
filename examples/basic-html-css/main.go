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
<div class="page">
  <h1>Basic HTML &amp; CSS</h1>
  <h2>Typography</h2>
  <p class="lead">This paragraph uses a larger font size and bold weight to demonstrate <span class="highlight">inline spans</span> inside block elements.</p>
  <p class="body-text">Normal body text sits here. Long lines wrap automatically when they exceed the container width, demonstrating inline text reflow across multiple lines without any extra configuration.</p>
  <h3>Colors &amp; Borders</h3>
  <div class="color-row">
    <div class="swatch hex">#4a90e2 hex</div>
    <div class="swatch rgb">rgb(80,200,120) rgb</div>
    <div class="swatch named">cornflowerblue named</div>
  </div>
  <h3>Text Alignment</h3>
  <p class="align-left">Left aligned text (default)</p>
  <p class="align-center">Center aligned text</p>
  <p class="align-right">Right aligned text</p>
</div>`

	css := `
* { margin: 0; padding: 0; }
.page {
  display: block;
  width: 760px;
  margin: 30px auto;
  padding: 24px;
  background-color: #ffffff;
  border: 2px solid #cccccc;
}
h1 {
  display: block;
  font-size: 28px;
  font-weight: bold;
  color: #1a1a2e;
  margin: 0 0 16px 0;
  padding: 12px 16px;
  background-color: #4a90e2;
  color: #ffffff;
  border: 1px solid #2a6bb0;
}
h2 {
  display: block;
  font-size: 20px;
  font-weight: bold;
  color: #333333;
  margin: 20px 0 10px 0;
}
h3 {
  display: block;
  font-size: 16px;
  font-weight: bold;
  color: #555555;
  margin: 18px 0 8px 0;
}
p { display: block; margin: 0 0 12px 0; }
.lead {
  font-size: 16px;
  font-weight: bold;
  color: #222222;
  padding: 10px;
  background-color: #f0f4ff;
  border: 1px solid #b0c4f0;
}
.highlight {
  color: #c0392b;
  font-style: italic;
}
.body-text {
  font-size: 14px;
  color: #444444;
  padding: 8px;
  background-color: #fafafa;
  border: 1px solid #e0e0e0;
}
.color-row {
  display: block;
  margin: 8px 0 16px 0;
}
.swatch {
  display: block;
  width: 240px;
  padding: 10px 14px;
  margin: 6px 0;
  font-size: 13px;
  color: #ffffff;
  border: 2px solid #333333;
}
.hex { background-color: #4a90e2; }
.rgb { background-color: rgb(80,200,120); color: #111111; }
.named { background-color: cornflowerblue; }
.align-left  { text-align: left;   padding: 6px; border: 1px solid #ddd; margin: 4px 0; }
.align-center{ text-align: center; padding: 6px; border: 1px solid #ddd; margin: 4px 0; }
.align-right { text-align: right;  padding: 6px; border: 1px solid #ddd; margin: 4px 0; }
`

	app, err := goglweb.New(html, css,
		goglweb.WithTitle("goglweb — Basic HTML & CSS"),
		goglweb.WithViewport(820, 700),
	)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}
	if err := app.Run(); err != nil {
		log.Fatalf("App error: %v", err)
	}
}
