package internal

import (
	"strings"
	"time"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type TextScrollData struct {
	NeedsScrolling      bool
	ScrollOffset        int32
	TextWidth           int32
	ContainerWidth      int32
	Direction           int
	LastDirectionChange *time.Time
}

func RenderMultilineText(renderer *sdl.Renderer, text string, font *ttf.Font, maxWidth int32, x, startY int32, color sdl.Color, alignment ...TextAlign) {

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

func RenderMultilineTextWithCache(
	renderer *sdl.Renderer,
	text string,
	font *ttf.Font,
	maxWidth int32,
	x, y int32,
	color sdl.Color,
	align TextAlign,
	cache *TextureCache) {

	if text == "" {
		return
	}

	_, fontHeight, err := font.SizeUTF8("Aj")
	if err != nil {
		fontHeight = 20
	}

	lineSpacing := int32(float32(fontHeight) * 0.3)
	lineY := y

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if line == "" {
			lineY += int32(fontHeight) + lineSpacing
			continue
		}

		remainingText := line
		for len(remainingText) > 0 {
			width, _, err := font.SizeUTF8(remainingText)
			if err != nil || int32(width) <= maxWidth {
				cacheKey := "line_" + remainingText + "_" + string(color.R) + string(color.G) + string(color.B)
				lineTexture := cache.Get(cacheKey)

				if lineTexture == nil {
					lineSurface, err := font.RenderUTF8Blended(remainingText, color)
					if err == nil {
						lineTexture, err = renderer.CreateTextureFromSurface(lineSurface)
						lineSurface.Free()

						if err == nil {
							cache.Set(cacheKey, lineTexture)
						}
					}
				}

				if lineTexture != nil {
					_, _, lineW, lineH, _ := lineTexture.Query()

					var lineX int32
					switch align {
					case TextAlignCenter:
						lineX = x + (maxWidth-lineW)/2
					case TextAlignRight:
						lineX = x + maxWidth - lineW
					default:
						lineX = x
					}

					lineRect := &sdl.Rect{
						X: lineX,
						Y: lineY,
						W: lineW,
						H: lineH,
					}

					renderer.Copy(lineTexture, nil, lineRect)
				}

				lineY += int32(fontHeight) + lineSpacing
				break
			}

			charsPerLine := int(float32(len(remainingText)) * float32(maxWidth) / float32(width))
			if charsPerLine <= 0 {
				charsPerLine = 1
			}

			if charsPerLine < len(remainingText) {
				for i := charsPerLine; i > 0; i-- {
					if i < len(remainingText) && remainingText[i] == ' ' {
						charsPerLine = i
						break
					}
				}
			}

			lineText := remainingText[:min(charsPerLine, len(remainingText))]
			cacheKey := "line_" + lineText + "_" + string(color.R) + string(color.G) + string(color.B)
			lineTexture := cache.Get(cacheKey)

			if lineTexture == nil {
				lineSurface, err := font.RenderUTF8Blended(lineText, color)
				if err == nil {
					lineTexture, err = renderer.CreateTextureFromSurface(lineSurface)
					lineSurface.Free()

					if err == nil {
						cache.Set(cacheKey, lineTexture)
					}
				}
			}

			if lineTexture != nil {
				_, _, lineW, lineH, _ := lineTexture.Query()

				var lineX int32
				switch align {
				case TextAlignCenter:
					lineX = x + (maxWidth-lineW)/2
				case TextAlignRight:
					lineX = x + maxWidth - lineW
				default:
					lineX = x
				}

				lineRect := &sdl.Rect{
					X: lineX,
					Y: lineY,
					W: lineW,
					H: lineH,
				}

				renderer.Copy(lineTexture, nil, lineRect)
			}

			lineY += int32(fontHeight) + lineSpacing

			if charsPerLine >= len(remainingText) {
				break
			}

			remainingText = remainingText[charsPerLine:]
			remainingText = strings.TrimLeft(remainingText, " ")
		}
	}
}

func DrawRoundedRect(renderer *sdl.Renderer, rect *sdl.Rect, radius int32, color sdl.Color) {
	if radius <= 0 {
		renderer.FillRect(rect)
		return
	}

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

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func Abs32(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}

func Min32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
func Max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
