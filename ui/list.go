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
	DefaultTitleXMargin int32 = 10
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
	SelectedItems map[int]bool // NEW: Track multiple selected items
	MultiSelect   bool         // NEW: Flag to enable multi-selection mode
	Settings      MenuSettings
	StartY        int32
	lastInputTime time.Time
	OnSelect      func(index int, item *models.MenuItem)

	VisibleStartIndex int
	MaxVisibleItems   int
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
		TitleXMargin: DefaultTitleXMargin,
	}
}

func NewListController(items []models.MenuItem, startY int32) *ListController {
	selectedItems := make(map[int]bool)
	selectedIndex := 0

	// Just track pre-selected items and the selected index
	// But don't force selection of any items
	for i, item := range items {
		if item.Selected {
			selectedIndex = i
			selectedItems[i] = true
		}
	}

	// Note: we're removing the automatic selection of the first item
	// This allows having no items selected as a valid state

	return &ListController{
		Items:         items,
		SelectedIndex: selectedIndex, // Still keep track of which item has focus
		SelectedItems: selectedItems, // This may be empty now
		MultiSelect:   false,
		Settings:      DefaultMenuSettings(),
		StartY:        startY,
		lastInputTime: time.Now(),
	}
}

func (lc *ListController) EnableMultiSelect(enable bool) {
	lc.MultiSelect = enable

	// When switching from multi to single and multiple items are selected
	if !enable && len(lc.SelectedItems) > 1 {
		// Clear all selections
		for i := range lc.Items {
			lc.Items[i].Selected = false
		}
		// Create a new map to clear all selections
		lc.SelectedItems = make(map[int]bool)

		// In single-select mode, the focused item gets selected
		lc.Items[lc.SelectedIndex].Selected = true
		lc.SelectedItems[lc.SelectedIndex] = true
	}
}

func (lc *ListController) toggleSelection(index int) {
	if index < 0 || index >= len(lc.Items) {
		return
	}

	// Simply toggle the selection status
	if lc.Items[index].Selected {
		lc.Items[index].Selected = false
		delete(lc.SelectedItems, index)
	} else {
		lc.Items[index].Selected = true
		lc.SelectedItems[index] = true
	}
}

func (lc *ListController) GetSelectedItems() []int {
	selectedIndices := make([]int, 0, len(lc.SelectedItems))
	for idx := range lc.SelectedItems {
		selectedIndices = append(selectedIndices, idx)
	}
	return selectedIndices
}

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
	case sdl.K_LEFT:
		lc.moveSelection(4)
		return true
	case sdl.K_RIGHT:
		lc.moveSelection(4)
		return true
	case sdl.K_1:
		if lc.MultiSelect {
			lc.toggleSelection(lc.SelectedIndex)
		}
		if lc.OnSelect != nil {
			lc.OnSelect(lc.SelectedIndex, &lc.Items[lc.SelectedIndex])
		}
		return true
	case sdl.K_2:
		if lc.MultiSelect {
			lc.toggleSelection(lc.SelectedIndex)
			return true
		}
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
	case sdl.CONTROLLER_BUTTON_DPAD_LEFT:
		lc.moveSelection(4)
		return true
	case sdl.CONTROLLER_BUTTON_DPAD_RIGHT:
		lc.moveSelection(4)
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

	// In multi-select mode, we don't automatically change selections when moving focus
	if !lc.MultiSelect {
		// In single-select mode, we deselect the previously focused item
		lc.Items[lc.SelectedIndex].Selected = false
		delete(lc.SelectedItems, lc.SelectedIndex)
	}

	lc.SelectedIndex = (lc.SelectedIndex + direction + len(lc.Items)) % len(lc.Items)

	// In single-select mode, we always mark the focused item as selected
	if !lc.MultiSelect {
		lc.Items[lc.SelectedIndex].Selected = true
		lc.SelectedItems[lc.SelectedIndex] = true
	}

	// Scrolling logic
	if lc.SelectedIndex < lc.VisibleStartIndex {
		lc.VisibleStartIndex = lc.SelectedIndex
	} else if lc.SelectedIndex >= lc.VisibleStartIndex+lc.MaxVisibleItems {
		lc.VisibleStartIndex = lc.SelectedIndex - lc.MaxVisibleItems + 1
	}
}

func (lc *ListController) Draw(renderer *sdl.Renderer) {
	for i := range lc.Items {
		// In multi-select mode, only the currently navigated item is focused
		lc.Items[i].Focused = (i == lc.SelectedIndex)
	}

	// Get visible items
	endIndex := lc.VisibleStartIndex + lc.MaxVisibleItems
	if endIndex > len(lc.Items) {
		endIndex = len(lc.Items)
	}
	visibleItems := lc.Items[lc.VisibleStartIndex:endIndex]

	drawScrollIndicators := len(lc.Items) > lc.MaxVisibleItems

	if lc.MultiSelect {
		// Reset focused state on all items
		for i := range lc.Items {
			lc.Items[i].Focused = false
		}

		// Set focused state on the current item if it's in view
		if lc.SelectedIndex >= lc.VisibleStartIndex &&
			lc.SelectedIndex < lc.VisibleStartIndex+lc.MaxVisibleItems {
			focusedItemIndex := lc.SelectedIndex - lc.VisibleStartIndex
			lc.Items[lc.VisibleStartIndex+focusedItemIndex].Focused = true
		}
	}

	DrawScrollableMenu(renderer, GetFont(), visibleItems, lc.StartY, lc.Settings,
		lc.VisibleStartIndex, drawScrollIndicators, true)
}

func DrawScrollableMenu(renderer *sdl.Renderer, font *ttf.Font, visibleItems []models.MenuItem,
	startY int32, settings MenuSettings, visibleStartIndex int,
	showScrollIndicators bool, multiSelect bool) {

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
	if settings.TitleXMargin < 0 {
		settings.TitleXMargin = DefaultTitleXMargin
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
		var bgColor sdl.Color

		itemText := item.Text

		// Different appearance for focus vs selection in multi-select mode
		if multiSelect {
			// Add selection indicator (checkbox)
			if item.Selected {
				itemText = "✓ " + itemText // Selected - show checkmark
			} else {
				itemText = "□ " + itemText // Not selected - show empty box
			}

			if item.Focused && item.Selected {
				// Focused and selected
				textColor = sdl.Color{R: 0, G: 0, B: 0, A: 255}     // Black text
				bgColor = sdl.Color{R: 220, G: 220, B: 255, A: 255} // Light blue bg
			} else if item.Focused {
				// Focused but not selected
				textColor = sdl.Color{R: 255, G: 255, B: 255, A: 255} // White text
				bgColor = sdl.Color{R: 100, G: 100, B: 180, A: 255}   // Darker blue bg
			} else if item.Selected {
				// Selected but not focused
				textColor = sdl.Color{R: 0, G: 0, B: 0, A: 255}     // Black text
				bgColor = sdl.Color{R: 180, G: 180, B: 180, A: 255} // Gray bg
			} else {
				// Neither focused nor selected
				textColor = sdl.Color{R: 255, G: 255, B: 255, A: 255} // White text
				// No background
			}
		} else {
			// Original single-select behavior
			if item.Selected {
				textColor = sdl.Color{R: 0, G: 0, B: 0, A: 255}
				bgColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
			} else {
				textColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
				// No background
			}
		}

		// Render text
		textSurface, err := font.RenderUTF8Blended(itemText, textColor)
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

		// Draw background for selected/focused item with rounded corners
		if item.Selected || (multiSelect && item.Focused) {
			pillRect := sdl.Rect{
				X: settings.XMargin,
				Y: itemY,
				W: textWidth + (settings.TextXPad * 2),
				H: textHeight + (settings.TextYPad * 2),
			}
			drawRoundedRect(renderer, &pillRect, 12, bgColor)
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
