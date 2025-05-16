package ui

import (
	"github.com/UncleJunVIP/gabagool/internal"
	"github.com/veandco/go-sdl2/gfx"
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
	footerHelpItems []FooterHelpItem,
	bottomPadding int32,
) {
	if len(footerHelpItems) == 0 {
		return
	}
	window := internal.GetWindow()
	windowWidth, windowHeight := window.Window.GetSize()
	y := windowHeight - bottomPadding - 50
	outerPillHeight := int32(60)
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
	// Render left group if it exists
	if len(leftItems) > 0 {
		renderGroupAsContinuousPill(renderer, font, leftItems, bottomPadding, y, outerPillHeight, innerPillMargin)
	}
	// Render right group if it exists
	if len(rightItems) > 0 {
		// Calculate total width of right group
		rightGroupWidth := calculateContinuousPillWidth(font, rightItems, outerPillHeight, innerPillMargin)
		rightX := windowWidth - bottomPadding - rightGroupWidth
		renderGroupAsContinuousPill(renderer, font, rightItems, rightX, y, outerPillHeight, innerPillMargin)
	}
}

// Helper function to calculate the width of a continuous pill containing multiple items
func calculateContinuousPillWidth(font *ttf.Font, items []FooterHelpItem, outerPillHeight, innerPillMargin int32) int32 {
	var totalWidth int32 = 0
	// Add left padding for the outer pill
	totalWidth += 20

	innerPillHeight := outerPillHeight - (innerPillMargin * 2)

	for i, item := range items {
		buttonSurface, err := font.RenderUTF8Blended(item.ButtonName, internal.GetTheme().MainColor)
		if err != nil {
			continue
		}
		helpSurface, err := font.RenderUTF8Blended(item.HelpText, internal.GetTheme().PrimaryAccentColor)
		if err != nil {
			buttonSurface.Free()
			continue
		}

		// Determine if inner pill should be circle or pill
		innerPillWidth := calculateInnerPillWidth(buttonSurface, innerPillHeight)

		// Calculate item width (button + help text + spacing)
		itemWidth := innerPillWidth + 15 + helpSurface.W
		totalWidth += itemWidth
		// Add spacing between items
		if i < len(items)-1 {
			totalWidth += 20
		}
		buttonSurface.Free()
		helpSurface.Free()
	}
	// Add right padding for the outer pill
	totalWidth += 15
	return totalWidth
}

// Helper function to calculate inner pill width based on content
func calculateInnerPillWidth(buttonSurface *sdl.Surface, innerPillHeight int32) int32 {
	// If text width is small enough, make a circle
	if buttonSurface.W <= innerPillHeight-20 { // Allow some padding
		return innerPillHeight // Circle has width = height
	} else {
		// Otherwise, make a pill with padding
		return buttonSurface.W + 20
	}
}

// Helper function to render a group of items as one continuous pill
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
	// Calculate the total width of the continuous pill
	pillWidth := calculateContinuousPillWidth(font, items, outerPillHeight, innerPillMargin)
	// Draw the outer pill (purple background)
	outerPillRect := &sdl.Rect{
		X: startX,
		Y: y,
		W: pillWidth,
		H: outerPillHeight,
	}

	renderer.SetDrawColor(internal.GetSDLColorValues(internal.GetTheme().PrimaryAccentColor))
	DrawRoundedRect(renderer, outerPillRect, outerPillHeight/2)
	// Start position for rendering items
	currentX := startX + 15 // Left padding
	innerPillHeight := outerPillHeight - (innerPillMargin * 2)
	// Render each button-text pair in sequence
	for _, item := range items {
		buttonSurface, err := font.RenderUTF8Blended(item.ButtonName, internal.GetTheme().HintInfoColor)
		if err != nil {
			continue
		}
		helpSurface, err := font.RenderUTF8Blended(item.HelpText, internal.GetTheme().HintInfoColor)
		if err != nil {
			buttonSurface.Free()
			continue
		}

		// Determine if inner pill should be circle or pill
		innerPillWidth := calculateInnerPillWidth(buttonSurface, innerPillHeight)
		isCircle := (innerPillWidth == innerPillHeight)

		// Set white color for inner pill
		renderer.SetDrawColor(internal.GetSDLColorValues(internal.GetTheme().MainColor))

		if isCircle {
			// Draw as circle
			DrawCircleShape(renderer, currentX+innerPillHeight/2, y+innerPillMargin+innerPillHeight/2, innerPillHeight/2)
		} else {
			// Draw as pill
			innerPillRect := &sdl.Rect{
				X: currentX,
				Y: y + innerPillMargin,
				W: innerPillWidth,
				H: innerPillHeight,
			}
			DrawRoundedRect(renderer, innerPillRect, innerPillHeight/2)
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
		// Move to position for help text
		currentX += innerPillWidth + 15
		// Render help text (white text)
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
		// Move to the next item position
		currentX += helpSurface.W + 30
		buttonSurface.Free()
		helpSurface.Free()
	}
}

func DrawCircleShape(renderer *sdl.Renderer, centerX, centerY, radius int32) {
	// Get current draw color
	r, g, b, a, _ := renderer.GetDrawColor()
	color := sdl.Color{R: r, G: g, B: b, A: a}

	// Draw the main filled circle
	gfx.FilledCircleColor(
		renderer,
		centerX,
		centerY,
		radius,
		color,
	)

	// Draw anti-aliased outline to smooth the edges
	gfx.AACircleColor(
		renderer,
		centerX,
		centerY,
		radius,
		color,
	)

	// Optional: Draw a slightly smaller anti-aliased circle to further smooth inner edges
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
