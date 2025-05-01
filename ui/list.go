package ui

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"nextui-sdl2/models"
	"time"
)

type ListSettings struct {
	ContentPadding Padding       // Padding inside menu items
	Margins        Padding       // Outer margins of the entire menu
	ItemSpacing    int32         // Vertical spacing between menu items
	InputDelay     time.Duration // Delay between input processing
	Title          string        // Optional title text
	TitleAlign     TextAlignment // Title alignment (left, center, right)
	TitleSpacing   int32         // Space between title and first item
}

type ListController struct {
	Items         []models.MenuItem
	SelectedIndex int
	SelectedItems map[int]bool
	MultiSelect   bool
	ReorderMode   bool
	Settings      ListSettings
	StartY        int32
	lastInputTime time.Time
	OnSelect      func(index int, item *models.MenuItem)

	VisibleStartIndex int
	MaxVisibleItems   int
	OnReorder         func(from, to int)
}

func DefaultListSettings(title string) ListSettings {
	return ListSettings{
		ContentPadding: Padding{
			Top:    5,
			Right:  10,
			Bottom: 5,
			Left:   10,
		},
		Margins: Padding{
			Top:    10,
			Right:  10,
			Bottom: 10,
			Left:   10,
		},
		ItemSpacing:  DefaultMenuSpacing,
		InputDelay:   DefaultInputDelay,
		Title:        title,
		TitleAlign:   AlignLeft,
		TitleSpacing: DefaultTitleSpacing,
	}
}

func NewListController(title string, items []models.MenuItem, startY int32) *ListController {
	selectedItems := make(map[int]bool)
	selectedIndex := 0

	for i, item := range items {
		if item.Selected {
			selectedIndex = i
			selectedItems[i] = true
		}
	}

	for i := range items {
		if i == selectedIndex {
			items[i].Selected = true
			selectedItems[i] = true
		} else {
			items[i].Selected = false
			delete(selectedItems, i)
		}
	}

	settings := DefaultListSettings(title)

	return &ListController{
		Items:         items,
		SelectedIndex: selectedIndex,
		SelectedItems: selectedItems,
		MultiSelect:   false,
		Settings:      settings,
		StartY:        startY,
		lastInputTime: time.Now(),
	}
}

func (lc *ListController) ToggleMultiSelect() {
	lc.MultiSelect = !lc.MultiSelect

	if !lc.MultiSelect && len(lc.SelectedItems) > 1 {
		for i := range lc.Items {
			lc.Items[i].Selected = false
		}

		lc.SelectedItems = make(map[int]bool)

		lc.Items[lc.SelectedIndex].Selected = true
		lc.SelectedItems[lc.SelectedIndex] = true
	}
}

func (lc *ListController) ToggleReorderMode() {
	lc.ReorderMode = !lc.ReorderMode
}

func (lc *ListController) MoveItemUp() bool {
	if !lc.ReorderMode || lc.SelectedIndex <= 0 {
		return false
	}

	currentIndex := lc.SelectedIndex
	prevIndex := currentIndex - 1

	lc.Items[currentIndex], lc.Items[prevIndex] = lc.Items[prevIndex], lc.Items[currentIndex]

	if lc.MultiSelect {
		if lc.SelectedItems[currentIndex] {
			delete(lc.SelectedItems, currentIndex)
			lc.SelectedItems[prevIndex] = true
		} else if lc.SelectedItems[prevIndex] {
			delete(lc.SelectedItems, prevIndex)
			lc.SelectedItems[currentIndex] = true
		}
	}

	lc.SelectedIndex = prevIndex

	lc.ScrollTo(lc.SelectedIndex)

	if lc.OnReorder != nil {
		lc.OnReorder(currentIndex, prevIndex)
	}

	return true
}

func (lc *ListController) MoveItemDown() bool {
	if !lc.ReorderMode || lc.SelectedIndex >= len(lc.Items)-1 {
		return false
	}

	currentIndex := lc.SelectedIndex
	nextIndex := currentIndex + 1

	lc.Items[currentIndex], lc.Items[nextIndex] = lc.Items[nextIndex], lc.Items[currentIndex]

	if lc.MultiSelect {
		if lc.SelectedItems[currentIndex] {
			delete(lc.SelectedItems, currentIndex)
			lc.SelectedItems[nextIndex] = true
		} else if lc.SelectedItems[nextIndex] {
			delete(lc.SelectedItems, nextIndex)
			lc.SelectedItems[currentIndex] = true
		}
	}

	lc.SelectedIndex = nextIndex

	lc.ScrollTo(lc.SelectedIndex)

	if lc.OnReorder != nil {
		lc.OnReorder(currentIndex, nextIndex)
	}

	return true
}

func (lc *ListController) ToggleSelection(index int) {
	if index < 0 || index >= len(lc.Items) {
		return
	}

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
			return lc.handleButtonPress(t.Button)
		}
	}
	return false
}

func (lc *ListController) handleKeyDown(key sdl.Keycode) bool {
	lc.lastInputTime = time.Now()

	if lc.ReorderMode {
		switch key {
		case sdl.K_UP:
			return lc.MoveItemUp()
		case sdl.K_DOWN:
			return lc.MoveItemDown()
		case sdl.K_ESCAPE:
			lc.ReorderMode = false
			return true
		case sdl.K_RETURN:
			lc.ReorderMode = false
			return true
		}
	}

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
	case sdl.K_0:
		lc.ToggleMultiSelect()
		return true
	case sdl.K_1:
		if lc.MultiSelect {
			lc.ToggleSelection(lc.SelectedIndex)
		}
		if lc.OnSelect != nil {
			lc.OnSelect(lc.SelectedIndex, &lc.Items[lc.SelectedIndex])
		}
		return true
	case sdl.K_2:
		if lc.MultiSelect {
			lc.ToggleSelection(lc.SelectedIndex)
			// Add the OnSelect call here
			if lc.OnSelect != nil {
				lc.OnSelect(lc.SelectedIndex, &lc.Items[lc.SelectedIndex])
			}
			return true
		}
	case sdl.K_3:
		lc.ToggleReorderMode()
		return true
	}
	return false
}

func (lc *ListController) handleButtonPress(button uint8) bool {
	lc.lastInputTime = time.Now()

	if lc.ReorderMode {
		switch button {
		case sdl.CONTROLLER_BUTTON_DPAD_UP:
			return lc.MoveItemUp()
		case sdl.CONTROLLER_BUTTON_DPAD_DOWN:
			return lc.MoveItemDown()
		case sdl.CONTROLLER_BUTTON_B:
			lc.ReorderMode = false
			return true
		case sdl.CONTROLLER_BUTTON_A:
			lc.ReorderMode = false
			return true
		}
	}

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
	case sdl.K_0:
		lc.ToggleMultiSelect()
		return true
	case sdl.CONTROLLER_BUTTON_A:
		if lc.MultiSelect {
			lc.ToggleSelection(lc.SelectedIndex)
		}
		if lc.OnSelect != nil {
			lc.OnSelect(lc.SelectedIndex, &lc.Items[lc.SelectedIndex])
		}
		return true
	case sdl.CONTROLLER_BUTTON_START:
		if lc.MultiSelect {
			lc.ToggleSelection(lc.SelectedIndex)
			// Add the OnSelect call here
			if lc.OnSelect != nil {
				lc.OnSelect(lc.SelectedIndex, &lc.Items[lc.SelectedIndex])
			}
			return true
		}
	case sdl.CONTROLLER_BUTTON_GUIDE:
		lc.ToggleReorderMode()
		return true
	}
	return false
}

func (lc *ListController) moveSelection(direction int) {
	if len(lc.Items) == 0 {
		return
	}

	if !lc.MultiSelect {
		lc.Items[lc.SelectedIndex].Selected = false
		delete(lc.SelectedItems, lc.SelectedIndex)
	}

	lc.SelectedIndex = (lc.SelectedIndex + direction + len(lc.Items)) % len(lc.Items)

	if !lc.MultiSelect {
		lc.Items[lc.SelectedIndex].Selected = true
		lc.SelectedItems[lc.SelectedIndex] = true
	}

	if lc.SelectedIndex < lc.VisibleStartIndex {
		lc.VisibleStartIndex = lc.SelectedIndex
	} else if lc.SelectedIndex >= lc.VisibleStartIndex+lc.MaxVisibleItems {
		lc.VisibleStartIndex = lc.SelectedIndex - lc.MaxVisibleItems + 1
	}
}

func (lc *ListController) Render(renderer *sdl.Renderer) {
	for i := range lc.Items {
		lc.Items[i].Focused = i == lc.SelectedIndex
	}

	endIndex := lc.VisibleStartIndex + lc.MaxVisibleItems
	if endIndex > len(lc.Items) {
		endIndex = len(lc.Items)
	}

	visibleItems := make([]models.MenuItem, endIndex-lc.VisibleStartIndex)
	for i, item := range lc.Items[lc.VisibleStartIndex:endIndex] {
		visibleItems[i] = item
	}

	if lc.MultiSelect {
		for i := range visibleItems {
			visibleItems[i].Focused = false
		}

		// Set focused state on the current item if it's in view
		if lc.SelectedIndex >= lc.VisibleStartIndex &&
			lc.SelectedIndex < lc.VisibleStartIndex+lc.MaxVisibleItems {
			focusedItemIndex := lc.SelectedIndex - lc.VisibleStartIndex
			visibleItems[focusedItemIndex].Focused = true
		}
	}

	originalTitle := lc.Settings.Title
	originalAlign := lc.Settings.TitleAlign

	if lc.ReorderMode {
		lc.Settings.Title = "REORDER MODE"
		lc.Settings.TitleAlign = AlignCenter

		if lc.SelectedIndex >= lc.VisibleStartIndex &&
			lc.SelectedIndex < lc.VisibleStartIndex+lc.MaxVisibleItems {
			selectedDisplayIndex := lc.SelectedIndex - lc.VisibleStartIndex
			visibleItems[selectedDisplayIndex].Text = "↕ " + visibleItems[selectedDisplayIndex].Text
		}
	}

	drawScrollableMenu(renderer, GetFont(), visibleItems, lc.StartY, lc.Settings, lc.MultiSelect)

	lc.Settings.Title = originalTitle
	lc.Settings.TitleAlign = originalAlign
}

func drawScrollableMenu(renderer *sdl.Renderer, font *ttf.Font, visibleItems []models.MenuItem,
	startY int32, settings ListSettings, multiSelect bool) {

	if settings.ItemSpacing <= 0 {
		settings.ItemSpacing = DefaultMenuSpacing
	}

	if settings.ContentPadding.Left <= 0 && settings.ContentPadding.Right <= 0 &&
		settings.ContentPadding.Top <= 0 && settings.ContentPadding.Bottom <= 0 {
		settings.ContentPadding = HVPadding(DefaultTextPadding, 5)
	}

	if settings.Margins.Left <= 0 && settings.Margins.Right <= 0 &&
		settings.Margins.Top <= 0 && settings.Margins.Bottom <= 0 {
		settings.Margins = UniformPadding(10)
	}

	if settings.TitleSpacing <= 0 {
		settings.TitleSpacing = DefaultTitleSpacing
	}

	itemStartY := startY

	if settings.Title != "" {
		itemStartY = drawTitle(renderer, GetTitleFont(), settings.Title,
			settings.TitleAlign, startY, settings.Margins.Left) + settings.TitleSpacing
	}

	for i, item := range visibleItems {
		var textSurface *sdl.Surface
		var textColor sdl.Color
		var bgColor sdl.Color

		itemText := item.Text

		if multiSelect {
			if item.Selected {
				itemText = "☑ " + itemText // Selected - show checkmark
			} else {
				itemText = "☐ " + itemText // Not selected - show empty box
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
			// Single-select behavior
			if item.Selected {
				textColor = sdl.Color{R: 0, G: 0, B: 0, A: 255}
				bgColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
			} else if item.Focused {
				textColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
				bgColor = sdl.Color{R: 100, G: 100, B: 180, A: 255} // Highlight focused item
			} else {
				textColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
				// No background
			}
		}

		textSurface, err := font.RenderUTF8Blended(itemText, textColor)
		if err != nil {
			Logger.Error("Failed to render text", "error", err)
			continue
		}

		textTexture, err := renderer.CreateTextureFromSurface(textSurface)
		if err != nil {
			Logger.Error("Failed to create texture", "error", err)
			textSurface.Free()
			continue
		}

		textWidth := textSurface.W
		textHeight := textSurface.H
		textSurface.Free()

		// Update references to padding values
		itemY := itemStartY + int32(i)*(textHeight+settings.ContentPadding.Top+settings.ContentPadding.Bottom+settings.ItemSpacing)

		if item.Selected || item.Focused {
			pillRect := sdl.Rect{
				X: settings.Margins.Left,
				Y: itemY,
				W: textWidth + (settings.ContentPadding.Left + settings.ContentPadding.Right),
				H: textHeight + (settings.ContentPadding.Top + settings.ContentPadding.Bottom),
			}
			drawRoundedRect(renderer, &pillRect, 12, bgColor)
		}

		textRect := sdl.Rect{
			X: settings.Margins.Left + settings.ContentPadding.Left,
			Y: itemY + settings.ContentPadding.Top,
			W: textWidth,
			H: textHeight,
		}
		renderer.Copy(textTexture, nil, &textRect)
		textTexture.Destroy()
	}
}

func drawTitle(renderer *sdl.Renderer, font *ttf.Font, title string, titleAlign TextAlignment, startY int32, titleXMargin int32) int32 {
	titleColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
	titleSurface, err := font.RenderUTF8Blended(title, titleColor)
	if err != nil {
		Logger.Error("Failed to render title text", "error", err)
		return startY
	}

	titleTexture, err := renderer.CreateTextureFromSurface(titleSurface)
	if err != nil {
		Logger.Error("Failed to create title texture", "error", err)
		titleSurface.Free()
		return startY
	}

	titleWidth := titleSurface.W
	titleHeight := titleSurface.H
	screenWidth, _, err := renderer.GetOutputSize()
	if err != nil {
		screenWidth = 768 // fallback width
	}

	var titleX int32
	switch titleAlign {
	case AlignLeft:
		titleX = titleXMargin
	case AlignCenter:
		titleX = (screenWidth - titleWidth) / 2
	case AlignRight:
		titleX = screenWidth - titleWidth - titleXMargin
	default:
		titleX = titleXMargin
	}

	titleRect := sdl.Rect{
		X: titleX,
		Y: startY,
		W: titleWidth,
		H: titleHeight,
	}
	renderer.Copy(titleTexture, nil, &titleRect)

	titleTexture.Destroy()
	titleSurface.Free()

	return titleHeight + 20
}

func drawRoundedRect(renderer *sdl.Renderer, rect *sdl.Rect, radius int32, color sdl.Color) {
	renderer.SetDrawColor(color.R, color.G, color.B, color.A)

	middleRect := sdl.Rect{
		X: rect.X + radius,
		Y: rect.Y,
		W: rect.W - 2*radius,
		H: rect.H,
	}
	renderer.FillRect(&middleRect)

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
