package ui

import (
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"strings"
)

func renderMultilineText(renderer *sdl.Renderer, text string, font *ttf.Font, maxWidth int32, centerX, startY int32, color sdl.Color) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return
	}

	var lines []string
	currentLine := words[0]

	for _, word := range words[1:] {
		// Check if adding this word exceeds the max width
		testLine := currentLine + " " + word
		testSurface, err := font.RenderUTF8Solid(testLine, color)
		if err != nil {
			continue
		}

		if testSurface.W <= maxWidth {
			currentLine = testLine
			testSurface.Free()
		} else {
			// Add current line to lines and start a new line
			lines = append(lines, currentLine)
			currentLine = word
		}
	}

	// Add the last line
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	// Render each line
	lineHeight := int32(font.Height())
	totalHeight := lineHeight * int32(len(lines))
	currentY := startY - totalHeight/2 // Center vertically around startY

	for _, line := range lines {
		surface, err := font.RenderUTF8Solid(line, color)
		if err != nil {
			continue
		}

		texture, err := renderer.CreateTextureFromSurface(surface)
		if err == nil {
			rect := &sdl.Rect{
				X: centerX - surface.W/2, // Center horizontally
				Y: currentY,
				W: surface.W,
				H: surface.H,
			}
			renderer.Copy(texture, nil, rect)
			texture.Destroy()
		}

		surface.Free()
		currentY += lineHeight + 5 // Add a small gap between lines
	}
}

func drawRoundedRect(renderer *sdl.Renderer, rect *sdl.Rect, radius int32) {
	if radius <= 0 {
		renderer.FillRect(rect)
		return
	}
	// Get current draw color
	r, g, b, a, _ := renderer.GetDrawColor()
	color := sdl.Color{R: r, G: g, B: b, A: a}
	// Draw the main rectangle (center)
	gfx.BoxColor(
		renderer,
		rect.X+radius,
		rect.Y,
		rect.X+rect.W-radius,
		rect.Y+rect.H,
		color,
	)
	// Draw the left and right rectangles
	gfx.BoxColor(
		renderer,
		rect.X,
		rect.Y+radius,
		rect.X+radius,
		rect.Y+rect.H-radius,
		color,
	)
	gfx.BoxColor(
		renderer,
		rect.X+rect.W-radius,
		rect.Y+radius,
		rect.X+rect.W,
		rect.Y+rect.H-radius,
		color,
	)
	// Draw the four corner circles
	gfx.FilledCircleColor(
		renderer,
		rect.X+radius,
		rect.Y+radius,
		radius,
		color,
	)
	gfx.FilledCircleColor(
		renderer,
		rect.X+rect.W-radius,
		rect.Y+radius,
		radius,
		color,
	)
	gfx.FilledCircleColor(
		renderer,
		rect.X+radius,
		rect.Y+rect.H-radius,
		radius,
		color,
	)
	gfx.FilledCircleColor(
		renderer,
		rect.X+rect.W-radius,
		rect.Y+rect.H-radius,
		radius,
		color,
	)
}
