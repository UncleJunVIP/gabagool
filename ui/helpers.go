package ui

import (
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"strings"
)

func renderMultilineText(renderer *sdl.Renderer, text string, font *ttf.Font, maxWidth int32, centerX, startY int32, color sdl.Color) int32 {
	if text == "" {
		return 0
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return 0
	}

	lines := []string{}
	currentLine := words[0]

	wordSurface, err := font.RenderUTF8Solid(words[0], color)
	if err != nil {
		return 0
	}
	wordWidth := wordSurface.W
	wordSurface.Free()

	if wordWidth > maxWidth && len(words[0]) > 1 {

		currentLine = ""

		for _, char := range words[0] {
			testLine := currentLine + string(char)
			charSurface, err := font.RenderUTF8Solid(testLine, color)
			if err != nil {
				continue
			}

			if charSurface.W > maxWidth {

				if currentLine != "" {
					lines = append(lines, currentLine)
				}
				currentLine = string(char)
			} else {
				currentLine = testLine
			}
			charSurface.Free()
		}
		lines = append(lines, currentLine)

		words = words[1:]
		currentLine = ""
	}

	for i := 1; i < len(words); i++ {
		testLine := currentLine + " " + words[i]
		lineSurface, err := font.RenderUTF8Solid(testLine, color)
		if err != nil {
			continue
		}

		if lineSurface.W <= maxWidth {
			currentLine = testLine
		} else {
			lines = append(lines, currentLine)
			currentLine = words[i]
		}
		lineSurface.Free()
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	lineHeight := int32(font.Height())
	totalHeight := int32(0)

	for i, line := range lines {
		lineSurface, err := font.RenderUTF8Solid(line, color)
		if err != nil {
			continue
		}

		lineTexture, err := renderer.CreateTextureFromSurface(lineSurface)
		if err != nil {
			lineSurface.Free()
			continue
		}

		lineY := startY + int32(i)*lineHeight

		lineX := centerX - (lineSurface.W / 2)

		lineRect := &sdl.Rect{
			X: lineX,
			Y: lineY,
			W: lineSurface.W,
			H: lineSurface.H,
		}

		renderer.Copy(lineTexture, nil, lineRect)
		lineTexture.Destroy()
		lineSurface.Free()

		totalHeight = lineY + lineSurface.H - startY
	}

	return totalHeight
}

func DrawRoundedRect(renderer *sdl.Renderer, rect *sdl.Rect, radius int32) {
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
