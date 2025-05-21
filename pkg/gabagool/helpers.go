package gabagool

import (
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"strings"
)

const (
	TextAlignLeft TextAlign = iota

	TextAlignCenter
)

func renderMultilineText(renderer *sdl.Renderer, text string, font *ttf.Font, maxWidth int32, x, startY int32, color sdl.Color, alignment ...TextAlign) {

	textAlign := TextAlignCenter
	if len(alignment) > 0 {
		textAlign = alignment[0]
	}

	paragraphs := strings.Split(text, "\n")
	var lines []string

	for _, paragraph := range paragraphs {

		if paragraph == "" {
			lines = append(lines, "")
			continue
		}

		words := strings.Fields(paragraph)
		if len(words) == 0 {
			continue
		}

		currentLine := words[0]

		for _, word := range words[1:] {

			testLine := currentLine + " " + word
			testSurface, err := font.RenderUTF8Blended(testLine, color)
			if err != nil {
				continue
			}

			if testSurface.W <= maxWidth {
				currentLine = testLine
				testSurface.Free()
			} else {

				lines = append(lines, currentLine)
				currentLine = word
			}
		}

		if currentLine != "" {
			lines = append(lines, currentLine)
		}
	}

	if len(lines) == 0 {
		return
	}

	lineHeight := int32(font.Height())
	totalHeight := lineHeight * int32(len(lines))

	var currentY int32
	if textAlign == TextAlignCenter {

		currentY = startY - totalHeight/2
	} else {

		currentY = startY
	}

	for _, line := range lines {

		if line == "" {
			currentY += lineHeight + 5
			continue
		}

		surface, err := font.RenderUTF8Blended(line, color)
		if err != nil {
			continue
		}

		texture, err := renderer.CreateTextureFromSurface(surface)
		if err == nil {
			rect := &sdl.Rect{
				Y: currentY,
				W: surface.W,
				H: surface.H,
			}

			if textAlign == TextAlignCenter {
				rect.X = x - surface.W/2
			} else {
				rect.X = x
			}

			renderer.Copy(texture, nil, rect)
			texture.Destroy()
		}

		surface.Free()
		currentY += lineHeight + 5
	}
}

func drawRoundedRect(renderer *sdl.Renderer, rect *sdl.Rect, radius int32) {
	if radius <= 0 {
		renderer.FillRect(rect)
		return
	}

	r, g, b, a, _ := renderer.GetDrawColor()
	color := sdl.Color{R: r, G: g, B: b, A: a}

	gfx.BoxColor(
		renderer,
		rect.X+radius,
		rect.Y,
		rect.X+rect.W-radius,
		rect.Y+rect.H,
		color,
	)

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

type textureCache struct {
	textures map[string]*sdl.Texture
}

func newTextureCache() *textureCache {
	return &textureCache{
		textures: make(map[string]*sdl.Texture),
	}
}

func (c *textureCache) get(key string) *sdl.Texture {
	if texture, exists := c.textures[key]; exists {
		return texture
	}
	return nil
}

func (c *textureCache) set(key string, texture *sdl.Texture) {
	c.textures[key] = texture
}

func (c *textureCache) destroy() {
	for _, texture := range c.textures {
		texture.Destroy()
	}
	c.textures = make(map[string]*sdl.Texture)
}
