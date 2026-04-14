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
  <h1>Flexbox Layout</h1>

  <h2>Row — justify-content: space-between</h2>
  <div class="flex-row spaced">
    <div class="box a">A</div>
    <div class="box b">B</div>
    <div class="box c">C</div>
  </div>

  <h2>Row — justify-content: center, align-items: center</h2>
  <div class="flex-row centered">
    <div class="box tall a">Tall A</div>
    <div class="box b">B</div>
    <div class="box a">C</div>
  </div>

  <h2>Row — flex-grow</h2>
  <div class="flex-row grow-demo">
    <div class="box grow1 a">grow:1</div>
    <div class="box grow2 b">grow:2</div>
    <div class="box grow1 c">grow:1</div>
  </div>

  <h2>Column</h2>
  <div class="flex-col">
    <div class="box a">Row 1</div>
    <div class="box b">Row 2</div>
    <div class="box c">Row 3</div>
  </div>
</div>`

	css := `
* { margin: 0; padding: 0; }
.page {
  display: block;
  width: 760px;
  margin: 20px auto;
  padding: 20px;
  background-color: #ffffff;
  border: 2px solid #cccccc;
}
h1 {
  display: block;
  font-size: 24px;
  font-weight: bold;
  padding: 10px 14px;
  margin: 0 0 16px 0;
  background-color: #6c5ce7;
  color: #ffffff;
  border: 1px solid #4a3ab0;
}
h2 {
  display: block;
  font-size: 14px;
  font-weight: bold;
  color: #555555;
  margin: 18px 0 8px 0;
}
.flex-row {
  display: flex;
  flex-direction: row;
  height: 60px;
  margin: 0 0 8px 0;
  background-color: #f8f8f8;
  border: 1px solid #dddddd;
  padding: 6px;
}
.flex-col {
  display: flex;
  flex-direction: column;
  margin: 0 0 8px 0;
  background-color: #f8f8f8;
  border: 1px solid #dddddd;
  padding: 6px;
}
.spaced  { justify-content: space-between; }
.centered{ justify-content: center; align-items: center; }
.grow-demo { align-items: stretch; }
.box {
  display: block;
  padding: 8px 12px;
  font-size: 13px;
  font-weight: bold;
  color: #ffffff;
  border: 1px solid rgba(0,0,0,0.2);
  margin: 2px;
}
.tall { height: 50px; }
.a { background-color: #e17055; }
.b { background-color: #00b894; }
.c { background-color: #0984e3; }
.grow1 { flex-grow: 1; }
.grow2 { flex-grow: 2; }
`

	app, err := goglweb.New(html, css,
		goglweb.WithTitle("goglweb — Flexbox"),
		goglweb.WithViewport(820, 620),
	)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}
	if err := app.Run(); err != nil {
		log.Fatalf("App error: %v", err)
	}
}
