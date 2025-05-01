package ui

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type Button struct {
	text      string
	x, y      int32
	w, h      int32
	selected  bool
	highlight bool
}

func DrawButton(renderer *sdl.Renderer, font *ttf.Font, button Button) {
	// Draw button background
	renderer.SetDrawColor(50, 50, 50, 255)
	buttonRect := sdl.Rect{X: button.x, Y: button.y, W: button.w, H: button.h}
	renderer.FillRect(&buttonRect)

	// Draw button border
	renderer.SetDrawColor(100, 100, 100, 255)
	renderer.DrawRect(&buttonRect)

	// Split text for POWER SLEEP button
	if button.text == "POWER SLEEP" {
		// Draw "POWER" part
		powerText := "POWER"
		var powerColor sdl.Color
		if button.highlight {
			powerColor = sdl.Color{R: 0, G: 0, B: 0, A: 255} // Black text for highlighted
		} else {
			powerColor = sdl.Color{R: 180, G: 180, B: 180, A: 255}
		}

		powerSurface, _ := font.RenderUTF8Blended(powerText, powerColor)
		powerTexture, _ := renderer.CreateTextureFromSurface(powerSurface)
		powerWidth := powerSurface.W
		powerHeight := powerSurface.H
		powerSurface.Free()

		// Draw power background if highlighted
		if button.highlight {
			// Use rounded rectangle for POWER highlight
			powerBgRect := sdl.Rect{
				X: button.x + 5,
				Y: button.y + 5,
				W: powerWidth + 10,
				H: powerHeight + 2,
			}
			drawRoundedRect(renderer, &powerBgRect, 8, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		}

		powerRect := sdl.Rect{X: button.x + 10, Y: button.y + 6, W: powerWidth, H: powerHeight}
		renderer.Copy(powerTexture, nil, &powerRect)
		powerTexture.Destroy()

		// Draw "SLEEP" part
		sleepText := "SLEEP"
		sleepColor := sdl.Color{R: 180, G: 180, B: 180, A: 255}
		sleepSurface, _ := font.RenderUTF8Blended(sleepText, sleepColor)
		sleepTexture, _ := renderer.CreateTextureFromSurface(sleepSurface)
		sleepWidth := sleepSurface.W
		sleepHeight := sleepSurface.H
		sleepSurface.Free()

		sleepRect := sdl.Rect{X: button.x + powerWidth + 20, Y: button.y + 6, W: sleepWidth, H: sleepHeight}
		renderer.Copy(sleepTexture, nil, &sleepRect)
		sleepTexture.Destroy()
	} else if button.text == "A OPEN" {
		// Draw "A" part
		aText := "A"
		aColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
		aSurface, _ := font.RenderUTF8Blended(aText, aColor)
		aTexture, _ := renderer.CreateTextureFromSurface(aSurface)
		aWidth := aSurface.W
		aHeight := aSurface.H
		aSurface.Free()

		// Draw circle around A
		renderer.SetDrawColor(255, 255, 255, 255)
		aCircleRect := sdl.Rect{X: button.x + 10, Y: button.y + 6, W: aWidth + 10, H: aHeight + 6}
		renderer.DrawRect(&aCircleRect)

		aRect := sdl.Rect{X: button.x + 15, Y: button.y + 9, W: aWidth, H: aHeight}
		renderer.Copy(aTexture, nil, &aRect)
		aTexture.Destroy()

		// Draw "OPEN" part
		openText := "OPEN"
		openColor := sdl.Color{R: 180, G: 180, B: 180, A: 255}
		openSurface, _ := font.RenderUTF8Blended(openText, openColor)
		openTexture, _ := renderer.CreateTextureFromSurface(openSurface)
		openWidth := openSurface.W
		openHeight := openSurface.H
		openSurface.Free()

		openRect := sdl.Rect{X: button.x + aWidth + 30, Y: button.y + 9, W: openWidth, H: openHeight}
		renderer.Copy(openTexture, nil, &openRect)
		openTexture.Destroy()
	}
}
