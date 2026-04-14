# interaction

Demonstrates the event system and DOM manipulation: `OnClick`, hover styling, `OnKey`, and `AppendChild`.

## Run

```bash
go run ./examples/interaction/main.go
```

## Features shown

- `OnClick` callback — click the blue buttons to toggle the `.active` CSS class
- Automatic hover styling — App adds/removes the `"hover"` class on the hovered element; CSS handles visuals via `.element.hover { }`
- `OnKey` callback — press **Space** to dynamically append a new `div` via `AppendChild` + `NewElement` + `NewText`
- `app.MarkDirty()` called after every DOM mutation to trigger repaint

## Screenshot

![Screenshot](screenshot.png)
