package ui

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"nextui-sdl2/models"
	"os"
	"time"
)

const (
	// Default values for menu layout
	DefaultMenuSpacing int32 = 60
	DefaultMenuXMargin int32 = 40 // Left margin for the menu items
	DefaultMenuYMargin int32 = 5  // Top/bottom margin within each menu item
	DefaultTextPadding int32 = 20 // Padding around text in the pill
	DefaultInputDelay        = 200 * time.Millisecond
)

// MenuSettings holds configuration for menu rendering
type MenuSettings struct {
	Spacing    int32         // Vertical spacing between menu items
	XMargin    int32         // Left margin for items and background
	YMargin    int32         // Top/bottom margin within each menu item
	TextXPad   int32         // Horizontal padding around text in the pill
	TextYPad   int32         // Vertical padding around text in the pill
	InputDelay time.Duration // Delay between input processing
}

// ListController handles menu navigation and selection
type ListController struct {
	Items         []models.MenuItem
	SelectedIndex int
	Settings      MenuSettings
	StartY        int32
	lastInputTime time.Time
	OnSelect      func(index int, item *models.MenuItem) // Callback for item selection
}

// DefaultMenuSettings returns a MenuSettings struct with default values
func DefaultMenuSettings() MenuSettings {
	return MenuSettings{
		Spacing:    DefaultMenuSpacing,
		XMargin:    DefaultMenuXMargin,
		YMargin:    DefaultMenuYMargin,
		TextXPad:   DefaultTextPadding,
		TextYPad:   5,
		InputDelay: DefaultInputDelay,
	}
}

// NewListController creates a new ListController with default settings
func NewListController(items []models.MenuItem, startY int32) *ListController {
	// Find the initially selected item
	selectedIndex := 0
	for i, item := range items {
		if item.Selected {
			selectedIndex = i
			break
		}
	}

	// Ensure at least one item is selected
	if len(items) > 0 {
		items[selectedIndex].Selected = true
	}

	return &ListController{
		Items:         items,
		SelectedIndex: selectedIndex,
		Settings:      DefaultMenuSettings(),
		StartY:        startY,
		lastInputTime: time.Now(),
	}
}

// HandleEvent processes SDL events for the list
func (lc *ListController) HandleEvent(event sdl.Event) bool {
	currentTime := time.Now()
	if currentTime.Sub(lc.lastInputTime) < lc.Settings.InputDelay {
		return false
	}

	switch t := event.(type) {
	case *sdl.KeyboardEvent:
		if t.Type == sdl.KEYDOWN {
			return lc.handleKeyDown(t.Keysym.Sym)
		}
	case *sdl.ControllerButtonEvent:
		if t.Type == sdl.CONTROLLERBUTTONDOWN {
			return lc.handleControllerButton(t.Button)
		}
	}
	return false
}

// handleKeyDown processes keyboard input
func (lc *ListController) handleKeyDown(key sdl.Keycode) bool {
	lc.lastInputTime = time.Now()

	switch key {
	case sdl.K_UP:
		lc.moveSelection(-1)
		return true
	case sdl.K_DOWN:
		lc.moveSelection(1)
		return true
	case sdl.K_RETURN:
		if lc.OnSelect != nil {
			lc.OnSelect(lc.SelectedIndex, &lc.Items[lc.SelectedIndex])
		}
		return true
	}
	return false
}

// handleControllerButton processes controller button input
func (lc *ListController) handleControllerButton(button uint8) bool {
	lc.lastInputTime = time.Now()

	switch button {
	case sdl.CONTROLLER_BUTTON_DPAD_UP:
		lc.moveSelection(-1)
		return true
	case sdl.CONTROLLER_BUTTON_DPAD_DOWN:
		lc.moveSelection(1)
		return true
	case sdl.CONTROLLER_BUTTON_A:
		if lc.OnSelect != nil {
			lc.OnSelect(lc.SelectedIndex, &lc.Items[lc.SelectedIndex])
		}
		return true
	}
	return false
}

// moveSelection changes the selected item
func (lc *ListController) moveSelection(direction int) {
	if len(lc.Items) == 0 {
		return
	}

	lc.Items[lc.SelectedIndex].Selected = false
	lc.SelectedIndex = (lc.SelectedIndex + direction + len(lc.Items)) % len(lc.Items)
	lc.Items[lc.SelectedIndex].Selected = true
}

// Draw renders the menu to the screen
func (lc *ListController) Draw(renderer *sdl.Renderer, font *ttf.Font) {
	DrawMenu(renderer, font, lc.Items, lc.StartY, lc.Settings)
}

// DrawMenu renders a menu without requiring a controller
func DrawMenu(renderer *sdl.Renderer, font *ttf.Font, menuItems []models.MenuItem, startY int32, settings MenuSettings) {
	// Use default settings for any invalid values
	if settings.Spacing <= 0 {
		settings.Spacing = DefaultMenuSpacing
	}
	if settings.XMargin < 0 {
		settings.XMargin = DefaultMenuXMargin
	}
	if settings.YMargin < 0 {
		settings.YMargin = DefaultMenuYMargin
	}
	if settings.TextXPad < 0 {
		settings.TextXPad = DefaultTextPadding
	}
	if settings.TextYPad < 0 {
		settings.TextYPad = 5
	}

	for i, item := range menuItems {
		var textSurface *sdl.Surface
		var textColor sdl.Color

		if item.Selected {
			textColor = sdl.Color{R: 0, G: 0, B: 0, A: 255}
		} else {
			textColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
		}

		textSurface, err := font.RenderUTF8Blended(item.Text, textColor)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to render text: %s\n", err)
			continue
		}

		textTexture, err := renderer.CreateTextureFromSurface(textSurface)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create texture: %s\n", err)
			textSurface.Free()
			continue
		}

		textWidth := textSurface.W
		textHeight := textSurface.H
		textSurface.Free()

		// Calculate positions based on the variable spacing and margins
		itemY := startY + int32(i)*settings.Spacing

		// Draw background for selected item with rounded corners
		if item.Selected {
			pillRect := sdl.Rect{
				X: settings.XMargin,
				Y: itemY,
				W: textWidth + (settings.TextXPad * 2),
				H: textHeight + (settings.TextYPad * 2),
			}
			drawRoundedRect(renderer, &pillRect, 12, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		}

		// Draw text
		textRect := sdl.Rect{
			X: settings.XMargin + settings.TextXPad,
			Y: itemY + settings.YMargin,
			W: textWidth,
			H: textHeight,
		}
		renderer.Copy(textTexture, nil, &textRect)
		textTexture.Destroy()
	}
}

// For backward compatibility with existing code
func DrawMenuSimple(renderer *sdl.Renderer, font *ttf.Font, menuItems []models.MenuItem, startY int32, spacing int32) {
	settings := DefaultMenuSettings()
	settings.Spacing = spacing
	DrawMenu(renderer, font, menuItems, startY, settings)
}

func drawRoundedRect(renderer *sdl.Renderer, rect *sdl.Rect, radius int32, color sdl.Color) {
	// Set the color
	renderer.SetDrawColor(color.R, color.G, color.B, color.A)

	// Draw the middle rectangle
	middleRect := sdl.Rect{
		X: rect.X + radius,
		Y: rect.Y,
		W: rect.W - 2*radius,
		H: rect.H,
	}
	renderer.FillRect(&middleRect)

	// Draw the left and right rectangles
	leftRect := sdl.Rect{
		X: rect.X,
		Y: rect.Y + radius,
		W: radius,
		H: rect.H - 2*radius,
	}
	rightRect := sdl.Rect{
		X: rect.X + rect.W - radius,
		Y: rect.Y + radius,
		W: radius,
		H: rect.H - 2*radius,
	}
	renderer.FillRect(&leftRect)
	renderer.FillRect(&rightRect)

	// Draw the four corner circles
	drawFilledCircle(renderer, rect.X+radius, rect.Y+radius, radius, color)
	drawFilledCircle(renderer, rect.X+rect.W-radius, rect.Y+radius, radius, color)
	drawFilledCircle(renderer, rect.X+radius, rect.Y+rect.H-radius, radius, color)
	drawFilledCircle(renderer, rect.X+rect.W-radius, rect.Y+rect.H-radius, radius, color)
}

func drawFilledCircle(renderer *sdl.Renderer, centerX, centerY, radius int32, color sdl.Color) {
	renderer.SetDrawColor(color.R, color.G, color.B, color.A)

	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			// If point is within the circle
			if x*x+y*y <= radius*radius {
				renderer.DrawPoint(centerX+x, centerY+y)
			}
		}
	}
}

// Update the drawButton function to use rounded rectangles for the POWER highlight
func drawButton(renderer *sdl.Renderer, font *ttf.Font, button Button) {
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
