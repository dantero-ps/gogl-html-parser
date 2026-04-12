package layout

import "strings"

// TextMetrics holds measured dimensions for a string of text.
type TextMetrics struct {
	Width      float64
	Height     float64
	Ascent     float64 // distance from top of box to baseline
	LineHeight float64
}

// TextMeasurer is the interface for measuring text using real font metrics.
type TextMeasurer interface {
	MeasureText(text string, fontFamily string, fontSize float64) TextMetrics
	MeasureWord(word string, fontFamily string, fontSize float64) float64
}

// FallbackMeasurer implements TextMeasurer using the byte-count x 0.6 approximation.
// Used when no real font measurer is available.
type FallbackMeasurer struct{}

func (f *FallbackMeasurer) MeasureText(text string, fontFamily string, fontSize float64) TextMetrics {
	width := float64(len(text)) * fontSize * 0.6
	return TextMetrics{
		Width:      width,
		Height:     fontSize,
		Ascent:     fontSize * 0.8,
		LineHeight: fontSize * 1.2,
	}
}

func (f *FallbackMeasurer) MeasureWord(word string, fontFamily string, fontSize float64) float64 {
	return f.MeasureText(word, fontFamily, fontSize).Width
}

// WordWrap breaks text into lines that fit within maxWidth using the given measurer.
// Returns a slice of line strings.
func WordWrap(text string, fontFamily string, fontSize float64, maxWidth float64, measurer TextMeasurer) []string {
	if measurer == nil {
		measurer = &FallbackMeasurer{}
	}
	words := splitWords(text)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	currentLine := ""

	for _, word := range words {
		candidate := word
		if currentLine != "" {
			candidate = currentLine + " " + word
		}
		w := measurer.MeasureText(candidate, fontFamily, fontSize).Width
		if w > maxWidth && currentLine != "" {
			lines = append(lines, currentLine)
			currentLine = word
		} else {
			currentLine = candidate
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines
}

// splitWords splits text into words on whitespace boundaries.
func splitWords(text string) []string {
	return strings.Fields(text)
}
