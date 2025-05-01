package ui

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"nextui-sdl2/models"
	"os"
	"time"
)

type TextAlignment int

const (
	AlignLeft TextAlignment = iota
	AlignCenter
	AlignRight
)

const (
	DefaultMenuSpacing  int32 = 60
	DefaultMenuXMargin  int32 = 40 // Left margin for the menu items
	DefaultMenuYMargin  int32 = 5  // Top/bottom margin within each menu item
	DefaultTextPadding  int32 = 20 // Padding around text in the pill
	DefaultInputDelay         = 200 * time.Millisecond
	DefaultTitleSpacing int32 = 30 // Space between title and first menu item
)

type MenuSettings struct {
	Spacing      int32         // Vertical spacing between menu items
	XMargin      int32         // Left margin for items and background
	YMargin      int32         // Top/bottom margin within each menu item
	TextXPad     int32         // Horizontal padding around text in the pill
	TextYPad     int32         // Vertical padding around text in the pill
	InputDelay   time.Duration // Delay between input processing
	Title        string        // Optional title text
	TitleAlign   TextAlignment // Title alignment (left, center, right)
	TitleSpacing int32         // Space between title and first item
	TitleXMargin int32
}

type ListController struct {
	Items         []models.MenuItem
	SelectedIndex int
	Settings      MenuSettings
	StartY        int32
	lastInputTime time.Time
	OnSelect      func(index int, item *models.MenuItem) // Callback for item selection

	// New fields for scrolling
	VisibleStartIndex int // First visible item index
	MaxVisibleItems   int // Maximum number of items that can be displayed at once
}

func DefaultMenuSettings() MenuSettings {
	return MenuSettings{
		Spacing:      DefaultMenuSpacing,
		XMargin:      DefaultMenuXMargin,
		YMargin:      DefaultMenuYMargin,
		TextXPad:     DefaultTextPadding,
		TextYPad:     5,
		InputDelay:   DefaultInputDelay,
		Title:        "",
		TitleAlign:   AlignLeft,
		TitleSpacing: DefaultTitleSpacing,
	}
}

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

// SetTitle sets the title text and alignment for the list
func (lc *ListController) SetTitle(title string, alignment TextAlignment) {
	lc.Settings.Title = title
	lc.Settings.TitleAlign = alignment
}

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

func (lc *ListController) moveSelection(direction int) {
	if len(lc.Items) == 0 {
		return
	}

	lc.Items[lc.SelectedIndex].Selected = false
	lc.SelectedIndex = (lc.SelectedIndex + direction + len(lc.Items)) % len(lc.Items)
	lc.Items[lc.SelectedIndex].Selected = true

	// Scrolling logic - adjust visible area when selection moves outside of it
	if lc.SelectedIndex < lc.VisibleStartIndex {
		// Selection moved above visible area - scroll up
		lc.VisibleStartIndex = lc.SelectedIndex
	} else if lc.SelectedIndex >= lc.VisibleStartIndex+lc.MaxVisibleItems {
		// Selection moved below visible area - scroll down
		lc.VisibleStartIndex = lc.SelectedIndex - lc.MaxVisibleItems + 1
	}
}

func (lc *ListController) Draw(renderer *sdl.Renderer, font *ttf.Font) {
	// Calculate how many items we can display based on screen height
	_, screenHeight, _ := renderer.GetOutputSize()
	availableHeight := screenHeight - lc.StartY

	// Calculate max visible items if not already set
	if lc.MaxVisibleItems <= 0 {
		lc.MaxVisibleItems = int(availableHeight / lc.Settings.Spacing)
	}

	// Limit MaxVisibleItems to avoid going beyond available space
	if lc.MaxVisibleItems > len(lc.Items) {
		lc.MaxVisibleItems = len(lc.Items)
	}

	// Ensure VisibleStartIndex is within bounds
	endIndex := lc.VisibleStartIndex + lc.MaxVisibleItems
	if endIndex > len(lc.Items) {
		endIndex = len(lc.Items)
	}

	// Draw only the visible portion of the list
	visibleItems := lc.Items[lc.VisibleStartIndex:endIndex]

	// Add scroll indicators if needed
	drawScrollIndicators := len(lc.Items) > lc.MaxVisibleItems

	DrawScrollableMenu(renderer, font, visibleItems, lc.StartY, lc.Settings,
		lc.VisibleStartIndex, drawScrollIndicators)
}

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
	if settings.TitleSpacing <= 0 {
		settings.TitleSpacing = DefaultTitleSpacing
	}

	// Adjusted startY to account for title if present
	itemStartY := startY

	// Draw title if one is set
	if settings.Title != "" {
		// Draw the title with underline and get the new starting Y position
		itemStartY = drawUnderlinedTitle(renderer, font, settings.Title,
			settings.TitleAlign, startY, settings.TitleXMargin) + settings.TitleSpacing
	}

	// Draw menu items
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
		itemY := itemStartY + int32(i)*settings.Spacing

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

func DrawScrollableMenu(renderer *sdl.Renderer, font *ttf.Font, visibleItems []models.MenuItem,
	startY int32, settings MenuSettings, visibleStartIndex int, showScrollIndicators bool) {

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
	if settings.TitleSpacing <= 0 {
		settings.TitleSpacing = DefaultTitleSpacing
	}

	// Adjusted startY to account for title if present
	itemStartY := startY

	// Draw title if one is set
	if settings.Title != "" {
		// Draw the title with underline and get the new starting Y position
		itemStartY = drawUnderlinedTitle(renderer, font, settings.Title,
			settings.TitleAlign, startY, settings.TitleXMargin) + settings.TitleSpacing
	}

	// Draw scroll up indicator if necessary
	if showScrollIndicators && visibleStartIndex > 0 {
		arrowUp := "▲" // Unicode up arrow
		arrowColor := sdl.Color{R: 180, G: 180, B: 180, A: 255}

		arrowSurface, err := font.RenderUTF8Blended(arrowUp, arrowColor)
		if err == nil {
			arrowTexture, err := renderer.CreateTextureFromSurface(arrowSurface)
			if err == nil {
				screenWidth, _, _ := renderer.GetOutputSize()
				arrowRect := sdl.Rect{
					X: screenWidth - 30,
					Y: itemStartY - 25,
					W: arrowSurface.W,
					H: arrowSurface.H,
				}
				renderer.Copy(arrowTexture, nil, &arrowRect)
				arrowTexture.Destroy()
			}
			arrowSurface.Free()
		}
	}

	// Draw menu items
	for i, item := range visibleItems {
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
		// Each item's offset is based on its relative position in the visible portion
		itemY := itemStartY + int32(i)*settings.Spacing

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

	// Draw scroll down indicator if necessary
	if showScrollIndicators && visibleStartIndex+len(visibleItems) < len(visibleItems) {
		arrowDown := "▼" // Unicode down arrow
		arrowColor := sdl.Color{R: 180, G: 180, B: 180, A: 255}

		arrowSurface, err := font.RenderUTF8Blended(arrowDown, arrowColor)
		if err == nil {
			arrowTexture, err := renderer.CreateTextureFromSurface(arrowSurface)
			if err == nil {
				screenWidth, _, _ := renderer.GetOutputSize()
				lastItemY := itemStartY + int32(len(visibleItems)-1)*settings.Spacing
				arrowRect := sdl.Rect{
					X: screenWidth - 30,
					Y: lastItemY + 40,
					W: arrowSurface.W,
					H: arrowSurface.H,
				}
				renderer.Copy(arrowTexture, nil, &arrowRect)
				arrowTexture.Destroy()
			}
			arrowSurface.Free()
		}
	}
}

// ScrollTo scrolls the list to make a specific index visible
func (lc *ListController) ScrollTo(index int) {
	if index < 0 || index >= len(lc.Items) {
		return // Invalid index
	}

	// If the item is already visible, don't change anything
	if index >= lc.VisibleStartIndex && index < lc.VisibleStartIndex+lc.MaxVisibleItems {
		return
	}

	// Scroll to make the item visible
	if index < lc.VisibleStartIndex {
		// Item is above visible area, show it at the top
		lc.VisibleStartIndex = index
	} else {
		// Item is below visible area, show it at the bottom
		lc.VisibleStartIndex = index - lc.MaxVisibleItems + 1
		if lc.VisibleStartIndex < 0 {
			lc.VisibleStartIndex = 0
		}
	}
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

// Add a function to draw an underlined title
func drawUnderlinedTitle(renderer *sdl.Renderer, font *ttf.Font, title string,
	titleAlign TextAlignment, startY int32, titleXMargin int32) int32 {

	// Render title with white color
	titleColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
	titleSurface, err := font.RenderUTF8Blended(title, titleColor)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to render title text: %s\n", err)
		return startY // Return original startY if we can't render
	}

	titleTexture, err := renderer.CreateTextureFromSurface(titleSurface)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create title texture: %s\n", err)
		titleSurface.Free()
		return startY
	}

	// Get the width of the title and the renderer
	titleWidth := titleSurface.W
	titleHeight := titleSurface.H
	screenWidth, _, err := renderer.GetOutputSize()
	if err != nil {
		screenWidth = 800 // Fallback width if can't get screen width
	}

	// Calculate x position based on alignment
	var titleX int32
	switch titleAlign {
	case AlignLeft:
		titleX = titleXMargin
	case AlignCenter:
		titleX = (screenWidth - titleWidth) / 2
	case AlignRight:
		titleX = screenWidth - titleWidth - titleXMargin
	default:
		titleX = titleXMargin // Default to left alignment
	}

	// Draw title
	titleRect := sdl.Rect{
		X: titleX,
		Y: startY,
		W: titleWidth,
		H: titleHeight,
	}
	renderer.Copy(titleTexture, nil, &titleRect)

	// Draw underline
	underlineY := startY + titleHeight + 2    // Small gap between text and underline
	renderer.SetDrawColor(255, 255, 255, 255) // White color for underline

	// Determine underline width based on alignment
	var underlineX, underlineWidth int32

	switch titleAlign {
	case AlignLeft:
		underlineX = titleX
		underlineWidth = titleWidth
	case AlignCenter:
		underlineX = titleX
		underlineWidth = titleWidth
	case AlignRight:
		underlineX = titleX
		underlineWidth = titleWidth
	}

	// Draw the line
	renderer.DrawLine(underlineX, underlineY, underlineX+underlineWidth, underlineY)

	// Clean up
	titleTexture.Destroy()
	titleSurface.Free()

	// Return the new Y position after the title and underline
	return underlineY + 5 // Add a small gap after the underline
}
