package ui

import (
	"github.com/UncleJunVIP/gabagool/internal"
	"github.com/UncleJunVIP/gabagool/models"
	"github.com/patrickhuber/go-types"
	"github.com/patrickhuber/go-types/option"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"math"
	"time"
)

type textScrollData struct {
	needsScrolling bool
	scrollOffset   int32
	textWidth      int32
	containerWidth int32
	direction      int // 1 for right to left, -1 for left to right
	lastUpdateTime time.Time
	pauseCounter   int // To create a pause at the beginning and end of scrolling
}

type ListSettings struct {
	Margins           models.Padding
	ItemSpacing       int32
	InputDelay        time.Duration
	Title             string
	TitleAlign        internal.TextAlignment
	TitleSpacing      int32
	MultiSelectKey    sdl.Keycode
	MultiSelectButton uint8
	ReorderKey        sdl.Keycode
	ReorderButton     uint8
	ScrollSpeed       float32
	ScrollPauseTime   int
	FooterText        string
	FooterTextColor   sdl.Color
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

	EnableAction bool

	HelpEnabled bool
	helpOverlay *HelpOverlay
	ShowingHelp bool

	itemScrollData map[int]*textScrollData
}

var defaultListHelpLines = []string{
	"Navigation Controls:",
	"• Up / Down: Navigate through items",
	"• A: Select current item",
	"• B: Cancel and exit",
}

func DefaultListSettings(title string) ListSettings {
	return ListSettings{
		Margins:         models.UniformPadding(20),
		ItemSpacing:     internal.DefaultMenuSpacing,
		InputDelay:      internal.DefaultInputDelay,
		Title:           title,
		TitleAlign:      internal.AlignLeft,
		TitleSpacing:    internal.DefaultTitleSpacing,
		ScrollSpeed:     150.0, // pixels per second
		ScrollPauseTime: 25,    // frames to pause at edges
		FooterText:      "",    // Empty by default
		FooterTextColor: sdl.Color{R: 180, G: 180, B: 180, A: 255},
	}
}

func NewListController(title string, items []models.MenuItem) *ListController {
	selectedItems := make(map[int]bool)
	selectedIndex := 0

	for i, item := range items {
		if item.Selected {
			selectedIndex = i
			break
		}
	}

	for i := range items {
		items[i].Selected = (i == selectedIndex)
		if items[i].Selected {
			selectedItems[i] = true
		}
	}

	return &ListController{
		Items:          items,
		SelectedIndex:  selectedIndex,
		SelectedItems:  selectedItems,
		MultiSelect:    false,
		Settings:       DefaultListSettings(title),
		StartY:         20,
		lastInputTime:  time.Now(),
		itemScrollData: make(map[int]*textScrollData),
	}
}

func NewBlockingList(title string, items []models.MenuItem, footerText string, enableAction bool, enableMultiSelect bool, enableReordering bool) (types.Option[models.ListReturn], error) {
	window := internal.GetWindow()
	renderer := window.Renderer

	listController := NewListController(title, items)

	listController.MaxVisibleItems = 8
	listController.EnableAction = enableAction
	listController.Settings.FooterText = footerText

	if enableMultiSelect {
		listController.Settings.MultiSelectKey = sdl.K_SPACE
		listController.Settings.MultiSelectButton = BrickButton_SELECT
	}

	if enableReordering {
		listController.Settings.ReorderKey = sdl.K_SPACE
		listController.Settings.ReorderButton = BrickButton_SELECT
	}

	running := true
	result := models.ListReturn{
		SelectedIndex:  -1,
		SelectedItem:   nil,
		LastPressedBtn: 0,
		Cancelled:      true,
	}
	var err error

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
				err = sdl.GetError()

			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {

					if e.Keysym.Sym == sdl.K_RETURN && listController.MultiSelect {
						running = false
						selectedIndices := listController.GetSelectedItems()

						// Populate the result with multiple selections
						if len(selectedIndices) > 0 {
							// Set the primary selection for backward compatibility
							result.SelectedIndex = selectedIndices[0]
							result.SelectedItem = &items[selectedIndices[0]]

							// Set all selections
							result.SelectedIndices = selectedIndices
							result.SelectedItems = make([]*models.MenuItem, len(selectedIndices))
							for i, idx := range selectedIndices {
								result.SelectedItems[i] = &items[idx]
							}

							result.Cancelled = false
						}
						break
					} else if e.Keysym.Sym == sdl.K_a && listController.MultiSelect {
						listController.HandleEvent(event)
					} else if e.Keysym.Sym == sdl.K_a && !listController.MultiSelect {
						running = false
						result.SelectedIndex = listController.SelectedIndex
						result.SelectedItem = &items[listController.SelectedIndex]
						result.SelectedIndices = []int{listController.SelectedIndex}
						result.SelectedItems = []*models.MenuItem{&items[listController.SelectedIndex]}
						result.Cancelled = false
						break
					} else if e.Keysym.Sym == sdl.K_b {
						running = false
						result.SelectedIndex = -1
						result.Cancelled = true
						break
					}

					listController.HandleEvent(event)
				}

			case *sdl.ControllerButtonEvent:
				if e.Type == sdl.CONTROLLERBUTTONDOWN {
					result.LastPressedBtn = e.Button

					if e.Button == BrickButton_START && listController.MultiSelect {
						running = false
						selectedIndices := listController.GetSelectedItems()

						// Populate the result with multiple selections
						if len(selectedIndices) > 0 {
							// Set the primary selection for backward compatibility
							result.SelectedIndex = selectedIndices[0]
							result.SelectedItem = &items[selectedIndices[0]]

							// Set all selections
							result.SelectedIndices = selectedIndices
							result.SelectedItems = make([]*models.MenuItem, len(selectedIndices))
							for i, idx := range selectedIndices {
								result.SelectedItems[i] = &items[idx]
							}

							result.Cancelled = false

						}
						break
					} else if e.Button == BrickButton_A && listController.MultiSelect {
						listController.HandleEvent(event)
					} else if e.Button == BrickButton_A && !listController.MultiSelect {
						running = false
						result.SelectedIndex = listController.SelectedIndex
						result.SelectedItem = &items[listController.SelectedIndex]
						result.SelectedIndices = []int{listController.SelectedIndex}
						result.SelectedItems = []*models.MenuItem{&items[listController.SelectedIndex]}
						result.Cancelled = false
						break
					} else if e.Button == BrickButton_B {
						result.SelectedIndex = -1
						running = false
						break
					}

					listController.HandleEvent(event)
				}
			}
		}

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		listController.Render(renderer)

		renderer.Present()

		sdl.Delay(16)
	}

	if err != nil || result.Cancelled {
		return option.None[models.ListReturn](), err
	}

	return option.Some(result), nil
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

	if index >= lc.VisibleStartIndex && index < lc.VisibleStartIndex+lc.MaxVisibleItems {
		return
	}

	if index < lc.VisibleStartIndex {
		lc.VisibleStartIndex = index
	} else {
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

	if key == sdl.K_h {
		lc.ToggleHelp()
		return true
	}

	if lc.ShowingHelp {
		if key == sdl.K_UP {
			lc.ScrollHelpOverlay(-1)
			return true
		}
		if key == sdl.K_DOWN {
			lc.ScrollHelpOverlay(1)
			return true
		}

		if key != sdl.K_UP && key != sdl.K_DOWN {
			lc.ShowingHelp = false
		}
		return true
	}

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
		lc.moveSelection(-4)
		return true
	case sdl.K_RIGHT:
		lc.moveSelection(4)
		return true
	case lc.Settings.MultiSelectKey:
		lc.ToggleMultiSelect()
		return true
	case sdl.K_a:
		if lc.MultiSelect {
			lc.ToggleSelection(lc.SelectedIndex)
		}
		if lc.OnSelect != nil {
			lc.OnSelect(lc.SelectedIndex, &lc.Items[lc.SelectedIndex])
		}
		return true
	case sdl.K_z:
		if lc.MultiSelect {
			lc.ToggleSelection(lc.SelectedIndex)
			if lc.OnSelect != nil {
				lc.OnSelect(lc.SelectedIndex, &lc.Items[lc.SelectedIndex])
			}
			return true
		}
	case lc.Settings.ReorderKey:
		lc.ToggleReorderMode()
		return true
	}
	return false
}

func (lc *ListController) handleButtonPress(button uint8) bool {
	lc.lastInputTime = time.Now()

	if button == BrickButton_MENU {
		lc.ToggleHelp()
		return true
	}

	if lc.ShowingHelp {
		if button == BrickButton_UP {
			lc.ScrollHelpOverlay(-1)
			return true
		}
		if button == BrickButton_DOWN {
			lc.ScrollHelpOverlay(1)
			return true
		}

		return true
	}

	if lc.ReorderMode {
		switch button {
		case BrickButton_UP:
			return lc.MoveItemUp()
		case BrickButton_DOWN:
			return lc.MoveItemDown()
		case BrickButton_B:
			lc.ReorderMode = false
			return true
		case BrickButton_A:
			lc.ReorderMode = false
			return true
		}
	}

	switch button {
	case BrickButton_UP:
		lc.moveSelection(-1)
		return true
	case BrickButton_DOWN:
		lc.moveSelection(1)
		return true
	case BrickButton_LEFT:
		lc.moveSelection(-4)
		return true
	case BrickButton_RIGHT:
		lc.moveSelection(4)
		return true
	case BrickButton_A:
		if lc.MultiSelect {
			lc.ToggleSelection(lc.SelectedIndex)
		}
		if lc.OnSelect != nil {
			lc.OnSelect(lc.SelectedIndex, &lc.Items[lc.SelectedIndex])
		}
		return true
	case lc.Settings.MultiSelectButton:
		lc.ToggleMultiSelect()
		return true
	case lc.Settings.ReorderButton:
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
	// First update any animated items
	lc.updateScrollingAnimations()

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
		lc.Settings.TitleAlign = internal.AlignCenter

		if lc.SelectedIndex >= lc.VisibleStartIndex &&
			lc.SelectedIndex < lc.VisibleStartIndex+lc.MaxVisibleItems {
			selectedDisplayIndex := lc.SelectedIndex - lc.VisibleStartIndex
			visibleItems[selectedDisplayIndex].Text = "↕ " + visibleItems[selectedDisplayIndex].Text
		}
	}

	drawScrollableMenu(renderer, internal.GetFont(), visibleItems, lc.StartY, lc.Settings, lc.MultiSelect, lc)

	lc.Settings.Title = originalTitle
	lc.Settings.TitleAlign = originalAlign

	if lc.ShowingHelp && lc.helpOverlay != nil {
		lc.helpOverlay.Render(renderer, internal.GetSmallFont())
	}
}

func drawScrollableMenu(renderer *sdl.Renderer, font *ttf.Font, visibleItems []models.MenuItem,
	startY int32, settings ListSettings, multiSelect bool, controller *ListController) {

	if settings.ItemSpacing <= 0 {
		settings.ItemSpacing = internal.DefaultMenuSpacing
	}

	if settings.Margins.Left <= 0 && settings.Margins.Right <= 0 &&
		settings.Margins.Top <= 0 && settings.Margins.Bottom <= 0 {
		settings.Margins = models.UniformPadding(10)
	}

	if settings.TitleSpacing <= 0 {
		settings.TitleSpacing = internal.DefaultTitleSpacing
	}

	itemStartY := startY

	if settings.Title != "" {
		itemStartY = drawTitle(renderer, internal.GetTitleFont(), settings.Title,
			settings.TitleAlign, startY, settings.Margins.Left) + settings.TitleSpacing
	}

	constantPillHeight := int32(60)
	screenWidth, _, err := renderer.GetOutputSize()
	if err != nil {
		screenWidth = 768 // fallback width
	}

	// Maximum width available for text (considering margins)
	maxTextWidth := screenWidth - settings.Margins.Left - settings.Margins.Right - 15 // Extra padding

	for i, item := range visibleItems {
		var textSurface *sdl.Surface
		var textColor sdl.Color
		var bgColor sdl.Color

		itemText := item.Text

		if multiSelect {
			if item.Selected {
				itemText = "☑ " + itemText
			} else {
				itemText = "☐ " + itemText
			}

			if item.Focused && item.Selected {
				textColor = sdl.Color{R: 0, G: 0, B: 0, A: 255}
				bgColor = sdl.Color{R: 220, G: 220, B: 255, A: 255}
			} else if item.Focused {
				textColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
				bgColor = sdl.Color{R: 100, G: 100, B: 180, A: 255}
			} else if item.Selected {
				textColor = sdl.Color{R: 0, G: 0, B: 0, A: 255}
				bgColor = sdl.Color{R: 180, G: 180, B: 180, A: 255}
			} else {
				textColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
			}
		} else {
			if item.Selected {
				textColor = sdl.Color{R: 0, G: 0, B: 0, A: 255}
				bgColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
			} else if item.Focused {
				textColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
				bgColor = sdl.Color{R: 100, G: 100, B: 180, A: 255}
			} else {
				textColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
			}
		}

		textSurface, err := font.RenderUTF8Blended(itemText, textColor)
		if err != nil {
			continue
		}

		textTexture, err := renderer.CreateTextureFromSurface(textSurface)
		if err != nil {
			textSurface.Free()
			continue
		}

		textWidth := textSurface.W
		textHeight := textSurface.H
		textSurface.Free()

		itemY := itemStartY + int32(i)*(constantPillHeight+settings.ItemSpacing)

		// Find the global index of this visible item
		globalIndex := controller.VisibleStartIndex + i

		// Check if we have scroll data for this item
		scrollData, hasScrollData := controller.itemScrollData[globalIndex]

		// Determine if the text needs scrolling
		needsScrolling := hasScrollData && scrollData.needsScrolling && item.Focused

		// Calculate pill width
		var pillWidth int32
		if needsScrolling {
			pillWidth = maxTextWidth + 10
		} else {
			pillWidth = textWidth + 10
		}

		if item.Selected || item.Focused {
			pillRect := sdl.Rect{
				X: settings.Margins.Left,
				Y: itemY,
				W: pillWidth,
				H: constantPillHeight,
			}
			drawRoundedRect(renderer, &pillRect, 12, bgColor)
		}

		// Center text vertically within the fixed-height pill
		textVerticalOffset := (constantPillHeight - textHeight) / 2

		// Small visual adjustment if needed
		textVerticalOffset += 1

		// For scrolling text, we need to use the clip rectangle
		if needsScrolling {
			// Create a clip rectangle for the text
			clipRect := &sdl.Rect{
				X: scrollData.scrollOffset,
				Y: 0,
				W: maxTextWidth,
				H: textHeight,
			}

			textRect := sdl.Rect{
				X: settings.Margins.Left + 5,
				Y: itemY + textVerticalOffset,
				W: maxTextWidth,
				H: textHeight,
			}

			renderer.Copy(textTexture, clipRect, &textRect)
		} else {
			// Normal text rendering for text that fits
			textRect := sdl.Rect{
				X: settings.Margins.Left + 5,
				Y: itemY + textVerticalOffset,
				W: textWidth,
				H: textHeight,
			}
			renderer.Copy(textTexture, nil, &textRect)
		}

		textTexture.Destroy()
	}

	if settings.FooterText != "" || controller.MultiSelect {
		_, screenHeight, err := renderer.GetOutputSize()

		footerText := settings.FooterText

		if controller.MultiSelect {
			footerText = "A Add / Remove | Select Cancel | Start Confirm"
		}

		// Render the footer text
		footerSurface, err := internal.GetSmallFont().RenderUTF8Blended(
			footerText,
			settings.FooterTextColor,
		)
		if err != nil {
			return
		}

		footerTexture, err := renderer.CreateTextureFromSurface(footerSurface)
		if err != nil {
			footerSurface.Free()
			return
		}

		footerWidth := footerSurface.W
		footerHeight := footerSurface.H
		footerSurface.Free()

		// Position in the bottom left with some padding
		footerRect := sdl.Rect{
			X: settings.Margins.Left,
			Y: screenHeight - footerHeight - settings.Margins.Bottom,
			W: footerWidth,
			H: footerHeight,
		}

		renderer.Copy(footerTexture, nil, &footerRect)
		footerTexture.Destroy()
	}
}

func (lc *ListController) updateScrollingAnimations() {
	// Get screen dimensions
	screenWidth, _, err := internal.GetWindow().Renderer.GetOutputSize()
	if err != nil {
		screenWidth = 768 // fallback width
	}

	maxTextWidth := screenWidth - lc.Settings.Margins.Left - lc.Settings.Margins.Right - 15

	// Update scrolling for visible items
	for idx := lc.VisibleStartIndex; idx < lc.VisibleStartIndex+lc.MaxVisibleItems && idx < len(lc.Items); idx++ {
		item := lc.Items[idx]

		// We only animate scrolling for the focused item
		if !item.Focused {
			// Clear scrolling data for non-focused items
			delete(lc.itemScrollData, idx)
			continue
		}

		// Get or create scrolling data for this item
		scrollData, exists := lc.itemScrollData[idx]
		if !exists {
			// Measure text width to determine if scrolling is needed
			prefix := ""
			if lc.MultiSelect {
				if item.Selected {
					prefix = "☑ "
				} else {
					prefix = "☐ "
				}
			}
			if lc.ReorderMode && idx == lc.SelectedIndex {
				prefix = "↕ " + prefix
			}

			textSurface, err := internal.GetFont().RenderUTF8Blended(prefix+item.Text, sdl.Color{R: 255, G: 255, B: 255, A: 255})
			if err != nil {
				continue
			}

			textWidth := textSurface.W
			textSurface.Free()

			scrollData = &textScrollData{
				needsScrolling: textWidth > maxTextWidth,
				textWidth:      textWidth,
				containerWidth: maxTextWidth,
				direction:      1, // Start scrolling left to right
				scrollOffset:   0,
				lastUpdateTime: time.Now(),
				pauseCounter:   lc.Settings.ScrollPauseTime, // Start with a pause
			}

			lc.itemScrollData[idx] = scrollData
		}

		// Skip updating if text doesn't need scrolling
		if !scrollData.needsScrolling {
			continue
		}

		// Only update animation at appropriate intervals
		currentTime := time.Now()
		elapsed := currentTime.Sub(scrollData.lastUpdateTime).Seconds()
		scrollData.lastUpdateTime = currentTime

		// If we're pausing at an edge, count down the pause frames
		if scrollData.pauseCounter > 0 {
			scrollData.pauseCounter--
			continue
		}

		// Calculate new scroll offset
		pixelsToScroll := int32(float32(elapsed) * lc.Settings.ScrollSpeed)
		if pixelsToScroll < 1 {
			pixelsToScroll = 1 // Always scroll at least 1 pixel
		}

		// Update scroll position based on direction
		if scrollData.direction > 0 {
			// Scrolling right to left
			scrollData.scrollOffset += pixelsToScroll

			// Check if we've reached the end
			if scrollData.scrollOffset >= scrollData.textWidth-scrollData.containerWidth {
				scrollData.scrollOffset = scrollData.textWidth - scrollData.containerWidth
				scrollData.direction = -1                             // Reverse direction
				scrollData.pauseCounter = lc.Settings.ScrollPauseTime // Pause before scrolling back
			}
		} else {
			// Scrolling left to right
			scrollData.scrollOffset -= pixelsToScroll

			// Check if we've reached the beginning
			if scrollData.scrollOffset <= 0 {
				scrollData.scrollOffset = 0
				scrollData.direction = 1                              // Reverse direction
				scrollData.pauseCounter = lc.Settings.ScrollPauseTime // Pause before scrolling again
			}
		}
	}
}

func drawTitle(renderer *sdl.Renderer, font *ttf.Font, title string, titleAlign internal.TextAlignment, startY int32, titleXMargin int32) int32 {
	titleColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
	titleSurface, err := font.RenderUTF8Blended(title, titleColor)
	if err != nil {
		return startY
	}

	titleTexture, err := renderer.CreateTextureFromSurface(titleSurface)
	if err != nil {
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
	case internal.AlignLeft:
		titleX = titleXMargin
	case internal.AlignCenter:
		titleX = (screenWidth - titleWidth) / 2
	case internal.AlignRight:
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
	if radius <= 0 {
		renderer.SetDrawColor(color.R, color.G, color.B, color.A)
		renderer.FillRect(rect)
		return
	}

	if radius*2 > rect.W {
		radius = rect.W / 2
	}
	if radius*2 > rect.H {
		radius = rect.H / 2
	}

	renderer.SetDrawColor(color.R, color.G, color.B, color.A)

	centerRect := &sdl.Rect{
		X: rect.X + radius,
		Y: rect.Y,
		W: rect.W - 2*radius,
		H: rect.H,
	}
	renderer.FillRect(centerRect)

	sideRectLeft := &sdl.Rect{
		X: rect.X,
		Y: rect.Y + radius,
		W: radius,
		H: rect.H - 2*radius,
	}
	renderer.FillRect(sideRectLeft)

	sideRectRight := &sdl.Rect{
		X: rect.X + rect.W - radius,
		Y: rect.Y + radius,
		W: radius,
		H: rect.H - 2*radius,
	}
	renderer.FillRect(sideRectRight)

	drawFilledCorner(renderer, rect.X+radius, rect.Y+radius, radius, 1, color)
	drawFilledCorner(renderer, rect.X+rect.W-radius, rect.Y+radius, radius, 2, color)
	drawFilledCorner(renderer, rect.X+radius, rect.Y+rect.H-radius, radius, 3, color)
	drawFilledCorner(renderer, rect.X+rect.W-radius, rect.Y+rect.H-radius, radius, 4, color)
}

func drawFilledCorner(renderer *sdl.Renderer, centerX, centerY, radius int32, corner int, color sdl.Color) {
	renderer.SetDrawColor(color.R, color.G, color.B, color.A)

	radiusSquared := radius * radius

	// Determine corner offset based on which corner we're drawing
	var xOffset, yOffset int32
	switch corner {
	case 1: // Top-left
		xOffset = -1
		yOffset = -1
	case 2: // Top-right
		xOffset = 1
		yOffset = -1
	case 3: // Bottom-left
		xOffset = -1
		yOffset = 1
	case 4: // Bottom-right
		xOffset = 1
		yOffset = 1
	}

	// For each y-value, draw a horizontal line segment
	for dy := int32(0); dy <= radius; dy++ {
		// Calculate width at this height using the circle equation
		width := int32(math.Sqrt(float64(radiusSquared - dy*dy)))

		// Skip empty lines
		if width <= 0 {
			continue
		}

		// Calculate the starting y position for this line
		y := centerY + yOffset*dy

		// Calculate starting and ending x positions
		var startX, endX int32

		if xOffset < 0 {
			// Left side corners
			startX = centerX - width
			endX = centerX
		} else {
			// Right side corners
			startX = centerX
			endX = centerX + width
		}

		// Draw the horizontal line
		renderer.DrawLine(startX, y, endX, y)
	}
}

func (lc *ListController) ToggleHelp() {
	if !lc.HelpEnabled {
		return
	}

	if lc.helpOverlay == nil {
		helpLines := defaultListHelpLines

		lc.helpOverlay = NewHelpOverlay(helpLines)
	}

	lc.helpOverlay.Toggle()
	lc.ShowingHelp = lc.helpOverlay.ShowingHelp
}

func (lc *ListController) ScrollHelpOverlay(direction int) {
	if lc.helpOverlay != nil {
		lc.helpOverlay.Scroll(direction)
	}
}
