package gabagool

import (
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// FooterHelpItem represents a button and its help text that should be displayed in the footer.
// ButtonName is the text that will be displayed in the inner pill.
// HelpText is the text that will be displayed in the outer pill to the right of the button.
type FooterHelpItem struct {
	HelpText   string
	ButtonName string
}

func renderFooter(
	renderer *sdl.Renderer,
	font *ttf.Font,
	footerHelpItems []FooterHelpItem,
	bottomPadding int32,
	transparentBackground bool,
) {
	if len(footerHelpItems) == 0 {
		return
	}
	window := GetWindow()
	windowWidth, windowHeight := window.Window.GetSize()
	y := windowHeight - bottomPadding - 50
	outerPillHeight := int32(60)

	if !transparentBackground {
		// Add a black background for the entire footer area
		footerBackgroundRect := &sdl.Rect{
			X: 0,                    // Start from left edge
			Y: y - 10,               // Same Y as the pills
			W: windowWidth - 15,     // Full window width
			H: outerPillHeight + 50, // Same height as the pills
		}

		// set color to black and draw the footer background
		renderer.SetDrawColor(0, 0, 0, 255) // Black with full opacity
		renderer.FillRect(footerBackgroundRect)
	}

	// Rest of the function remains the same
	innerPillMargin := int32(6)
	var leftItems []FooterHelpItem
	var rightItems []FooterHelpItem
	switch len(footerHelpItems) {
	case 1:
		leftItems = footerHelpItems[0:1]
	case 2:
		leftItems = footerHelpItems[0:1]
		rightItems = footerHelpItems[1:2]
	case 3:
		leftItems = footerHelpItems[0:2]
		rightItems = footerHelpItems[2:3]
	case 4, 5, 6:
		leftItems = footerHelpItems[0:2]
		rightItems = footerHelpItems[2:min(4, len(footerHelpItems))]
	default:
		leftItems = footerHelpItems[0:2]
		rightItems = footerHelpItems[2:4]
	}

	if len(leftItems) > 0 {
		renderGroupAsContinuousPill(renderer, font, leftItems, bottomPadding, y, outerPillHeight, innerPillMargin)
	}
	if len(rightItems) > 0 {
		rightGroupWidth := calculateContinuousPillWidth(font, rightItems, outerPillHeight, innerPillMargin)
		rightX := windowWidth - bottomPadding - rightGroupWidth
		renderGroupAsContinuousPill(renderer, font, rightItems, rightX, y, outerPillHeight, innerPillMargin)
	}
}

func calculateContinuousPillWidth(font *ttf.Font, items []FooterHelpItem, outerPillHeight, innerPillMargin int32) int32 {
	var totalWidth int32 = 20

	innerPillHeight := outerPillHeight - (innerPillMargin * 2)

	for i, item := range items {
		buttonSurface, err := font.RenderUTF8Blended(item.ButtonName, GetTheme().MainColor)
		if err != nil {
			continue
		}
		helpSurface, err := font.RenderUTF8Blended(item.HelpText, GetTheme().PrimaryAccentColor)
		if err != nil {
			buttonSurface.Free()
			continue
		}

		innerPillWidth := calculateInnerPillWidth(buttonSurface, innerPillHeight)

		itemWidth := innerPillWidth + 15 + helpSurface.W
		totalWidth += itemWidth
		if i < len(items)-1 {
			totalWidth += 20
		}
		buttonSurface.Free()
		helpSurface.Free()
	}
	totalWidth += 15
	return totalWidth
}

func calculateInnerPillWidth(buttonSurface *sdl.Surface, innerPillHeight int32) int32 {
	if buttonSurface.W <= innerPillHeight-20 {
		return innerPillHeight
	} else {
		return buttonSurface.W + 20
	}
}

func renderGroupAsContinuousPill(
	renderer *sdl.Renderer,
	font *ttf.Font,
	items []FooterHelpItem,
	startX, y,
	outerPillHeight,
	innerPillMargin int32,
) {
	if len(items) == 0 {
		return
	}
	pillWidth := calculateContinuousPillWidth(font, items, outerPillHeight, innerPillMargin)
	outerPillRect := &sdl.Rect{
		X: startX,
		Y: y,
		W: pillWidth,
		H: outerPillHeight,
	}

	renderer.SetDrawColor(GetSDLColorValues(GetTheme().PrimaryAccentColor))
	drawRoundedRect(renderer, outerPillRect, outerPillHeight/2)

	currentX := startX + 10
	innerPillHeight := outerPillHeight - (innerPillMargin * 2)

	for _, item := range items {
		buttonSurface, err := font.RenderUTF8Blended(item.ButtonName, GetTheme().SecondaryAccentColor)
		if err != nil {
			continue
		}
		helpSurface, err := font.RenderUTF8Blended(item.HelpText, GetTheme().HintInfoColor)
		if err != nil {
			buttonSurface.Free()
			continue
		}

		innerPillWidth := calculateInnerPillWidth(buttonSurface, innerPillHeight)
		isCircle := (innerPillWidth == innerPillHeight)

		renderer.SetDrawColor(GetSDLColorValues(GetTheme().MainColor))

		if isCircle {
			drawCircleShape(renderer, currentX+innerPillHeight/2, y+innerPillMargin+innerPillHeight/2, innerPillHeight/2)
		} else {
			innerPillRect := &sdl.Rect{
				X: currentX,
				Y: y + innerPillMargin,
				W: innerPillWidth,
				H: innerPillHeight,
			}
			drawRoundedRect(renderer, innerPillRect, innerPillHeight/2)
		}

		buttonTexture, err := renderer.CreateTextureFromSurface(buttonSurface)
		if err == nil {
			buttonTextRect := &sdl.Rect{
				X: currentX + (innerPillWidth-buttonSurface.W)/2,
				Y: y + (outerPillHeight-buttonSurface.H)/2,
				W: buttonSurface.W,
				H: buttonSurface.H,
			}
			renderer.Copy(buttonTexture, nil, buttonTextRect)
			buttonTexture.Destroy()
		}

		currentX += innerPillWidth + 10

		helpTexture, err := renderer.CreateTextureFromSurface(helpSurface)
		if err == nil {
			helpTextRect := &sdl.Rect{
				X: currentX,
				Y: y + (outerPillHeight-helpSurface.H)/2,
				W: helpSurface.W,
				H: helpSurface.H,
			}
			renderer.Copy(helpTexture, nil, helpTextRect)
			helpTexture.Destroy()
		}

		currentX += helpSurface.W + 30
		buttonSurface.Free()
		helpSurface.Free()
	}
}

func drawCircleShape(renderer *sdl.Renderer, centerX, centerY, radius int32) {
	r, g, b, a, _ := renderer.GetDrawColor()
	color := sdl.Color{R: r, G: g, B: b, A: a}

	gfx.FilledCircleColor(
		renderer,
		centerX,
		centerY,
		radius,
		color,
	)

	gfx.AACircleColor(
		renderer,
		centerX,
		centerY,
		radius,
		color,
	)

	if radius > 2 {
		gfx.AACircleColor(
			renderer,
			centerX,
			centerY,
			radius-1,
			color,
		)
	}
}
