package internal

import (
	"strings"
	"time"

	"github.com/UncleJunVIP/gabagool/v2/pkg/gabagool/constants"
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

func RenderMultilineText(renderer *sdl.Renderer, text string, font *ttf.Font, maxWidth int32, x, startY int32, color sdl.Color, alignment ...constants.TextAlign) {

	textAlign := constants.TextAlignCenter
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
	if textAlign == constants.TextAlignCenter {

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

			if textAlign == constants.TextAlignCenter {
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
	align constants.TextAlign,
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
					case constants.TextAlignCenter:
						lineX = x + (maxWidth-lineW)/2
					case constants.TextAlignRight:
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
				case constants.TextAlignCenter:
					lineX = x + (maxWidth-lineW)/2
				case constants.TextAlignRight:
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

	// Top-left corner
	drawRoundedCorner(renderer, rect.X+radius, rect.Y+radius, radius, color)
	// Top-right corner
	drawRoundedCorner(renderer, rect.X+rect.W-radius, rect.Y+radius, radius, color)
	// Bottom-left corner
	drawRoundedCorner(renderer, rect.X+radius, rect.Y+rect.H-radius, radius, color)
	// Bottom-right corner
	drawRoundedCorner(renderer, rect.X+rect.W-radius, rect.Y+rect.H-radius, radius, color)
}

func drawRoundedCorner(renderer *sdl.Renderer, centerX, centerY, radius int32, color sdl.Color) {
	// Fill the corner
	gfx.FilledCircleColor(renderer, centerX, centerY, radius, color)

	// Add anti-aliased edge for smooth appearance
	gfx.AACircleColor(renderer, centerX, centerY, radius, color)

	// Add additional anti-aliased circles based on radius size for extra smoothness
	// Larger radii benefit from multiple AA layers to eliminate jaggedness
	if radius > 15 {
		// Large pills (like list items) - add 3 layers of AA
		gfx.AACircleColor(renderer, centerX, centerY, radius-1, color)
		gfx.AACircleColor(renderer, centerX, centerY, radius-2, color)
	} else if radius > 5 {
		// Medium pills (like footer buttons) - add 2 layers of AA
		gfx.AACircleColor(renderer, centerX, centerY, radius-1, color)
	} else if radius > 2 {
		// Small pills - add 1 layer of AA
		gfx.AACircleColor(renderer, centerX, centerY, radius-1, color)
	}
}

// DrawSmoothScrollbar renders a scrollbar with anti-aliased rounded ends
func DrawSmoothScrollbar(renderer *sdl.Renderer, x, y, width, height int32, color sdl.Color) {
	if width <= 0 || height <= 0 {
		return
	}

	// For narrow scrollbars, use fully rounded ends
	radius := width / 2
	if height < width {
		radius = height / 2
	}

	DrawRoundedRect(renderer, &sdl.Rect{X: x, Y: y, W: width, H: height}, radius, color)
}

// DrawSmoothProgressBar renders a progress bar with smooth rounded edges
func DrawSmoothProgressBar(renderer *sdl.Renderer, bgRect *sdl.Rect, fillWidth int32, bgColor, fillColor sdl.Color) {
	if bgRect == nil {
		return
	}

	// Draw background with rounded corners
	radius := bgRect.H / 2
	DrawRoundedRect(renderer, bgRect, radius, bgColor)

	// Draw fill with rounded corners if there's progress
	if fillWidth > 0 && fillWidth <= bgRect.W {
		fillRect := &sdl.Rect{
			X: bgRect.X,
			Y: bgRect.Y,
			W: Min32(fillWidth, bgRect.W),
			H: bgRect.H,
		}
		DrawRoundedRect(renderer, fillRect, radius, fillColor)
	}
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

func HexToColor(hex uint32) sdl.Color {
	r := uint8((hex >> 16) & 0xFF)
	g := uint8((hex >> 8) & 0xFF)
	b := uint8(hex & 0xFF)

	return sdl.Color{R: r, G: g, B: b, A: 255}
}
