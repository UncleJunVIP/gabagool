package ui

import (
	"github.com/UncleJunVIP/gabagool/internal"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type FooterHelpItem struct {
	HelpText   string
	ButtonName string
}

func RenderFooter(
	renderer *sdl.Renderer,
	font *ttf.Font,
	footerTextColor sdl.Color,
	footerHelpItems []FooterHelpItem,
	margins int32,
) {
	if len(footerHelpItems) == 0 {
		return
	}

	window := internal.GetWindow()
	_, height := window.Window.GetSize()

	y := height - margins - 30

	if len(footerHelpItems) > 0 {
		x := margins
		for _, item := range footerHelpItems {

			keySurface, err := font.RenderUTF8Blended(item.ButtonName, sdl.Color{R: 255, G: 255, B: 255, A: 255})
			if err == nil {
				keyTexture, err := renderer.CreateTextureFromSurface(keySurface)
				if err == nil {
					keyRect := &sdl.Rect{
						X: x,
						Y: y,
						W: keySurface.W,
						H: keySurface.H,
					}
					renderer.Copy(keyTexture, nil, keyRect)
					keyTexture.Destroy()
					x += keySurface.W + 10
				}
				keySurface.Free()
			}

			textSurface, err := font.RenderUTF8Blended(item.HelpText, footerTextColor)
			if err == nil {
				textTexture, err := renderer.CreateTextureFromSurface(textSurface)
				if err == nil {
					textRect := &sdl.Rect{
						X: x,
						Y: y,
						W: textSurface.W,
						H: textSurface.H,
					}
					renderer.Copy(textTexture, nil, textRect)
					textTexture.Destroy()
					x += textSurface.W + 30
				}
				textSurface.Free()
			}
		}
	}
}
