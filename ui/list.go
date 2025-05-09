package ui

import (
	"math"
	"time"

	"github.com/UncleJunVIP/gabagool/internal"
	"github.com/UncleJunVIP/gabagool/models"
	"github.com/patrickhuber/go-types"
	"github.com/patrickhuber/go-types/option"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	scrollDirectionRight = 1
	scrollDirectionLeft  = -1
)

type textScrollData struct {
	needsScrolling bool
	scrollOffset   int32
	textWidth      int32
	containerWidth int32
	direction      int
	lastUpdateTime time.Time
	pauseCounter   int
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

	// Find first pre-selected item if any
	for i, item := range items {
		if item.Selected {
			selectedIndex = i
			break
		}
	}

	// Update selection state
	for i := range items {
		items[i].Selected = i == selectedIndex
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
				if e.Type != sdl.KEYDOWN {
					continue
				}

				switch {
				case e.Keysym.Sym == sdl.K_x:
					running = false
					result.ActionTriggered = true
					result.Cancelled = false

				case e.Keysym.Sym == sdl.K_RETURN && listController.MultiSelect:
					running = false
					if indices := listController.GetSelectedItems(); len(indices) > 0 {
						result.PopulateMultiSelection(indices, items)
						result.Cancelled = false
					}

				case e.Keysym.Sym == sdl.K_a && listController.MultiSelect:
					listController.HandleEvent(event)

				case e.Keysym.Sym == sdl.K_a && !listController.MultiSelect:
					running = false
					result.PopulateSingleSelection(listController.SelectedIndex, items)
					result.Cancelled = false

				case e.Keysym.Sym == sdl.K_b:
					running = false
					result.SelectedIndex = -1
					result.Cancelled = true

				default:
					listController.HandleEvent(event)
				}

			case *sdl.ControllerButtonEvent:
				if e.Type != sdl.CONTROLLERBUTTONDOWN {
					continue
				}

				result.LastPressedBtn = e.Button

				switch {

				case e.Button == BrickButton_X && listController.EnableAction:
					running = false
					result.ActionTriggered = true
					result.Cancelled = false

				case e.Button == BrickButton_START && listController.MultiSelect:
					running = false
					if indices := listController.GetSelectedItems(); len(indices) > 0 {
						result.PopulateMultiSelection(indices, items)
						result.Cancelled = false
					}

				case e.Button == BrickButton_A && listController.MultiSelect:
					listController.HandleEvent(event)

				case e.Button == BrickButton_A && !listController.MultiSelect:
					running = false
					result.PopulateSingleSelection(listController.SelectedIndex, items)
					result.Cancelled = false

				case e.Button == BrickButton_B:
					result.SelectedIndex = -1
					running = false

				default:
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

// ToggleMultiSelect switches between single and multi-selection modes
func (lc *ListController) ToggleMultiSelect() {
	lc.MultiSelect = !lc.MultiSelect

	if !lc.MultiSelect && len(lc.SelectedItems) > 1 {
		// Reset selections when leaving multi-select mode
		for i := range lc.Items {
			lc.Items[i].Selected = false
		}

		lc.SelectedItems = make(map[int]bool)

		// Keep only the currently focused item selected
		lc.Items[lc.SelectedIndex].Selected = true
		lc.SelectedItems[lc.SelectedIndex] = true
	}
}

// ToggleReorderMode switches item reordering mode on/off
func (lc *ListController) ToggleReorderMode() {
	lc.ReorderMode = !lc.ReorderMode
}

// MoveItemUp moves the selected item up one position in the list
func (lc *ListController) MoveItemUp() bool {
	if !lc.ReorderMode || lc.SelectedIndex <= 0 {
		return false
	}

	currentIndex := lc.SelectedIndex
	prevIndex := currentIndex - 1

	// Swap items
	lc.Items[currentIndex], lc.Items[prevIndex] = lc.Items[prevIndex], lc.Items[currentIndex]

	// Update selection state
	if lc.MultiSelect {
		lc.updateSelectionAfterMove(currentIndex, prevIndex)
	}

	lc.SelectedIndex = prevIndex
	lc.ScrollTo(lc.SelectedIndex)

	if lc.OnReorder != nil {
		lc.OnReorder(currentIndex, prevIndex)
	}

	return true
}

// MoveItemDown moves the selected item down one position in the list
func (lc *ListController) MoveItemDown() bool {
	if !lc.ReorderMode || lc.SelectedIndex >= len(lc.Items)-1 {
		return false
	}

	currentIndex := lc.SelectedIndex
	nextIndex := currentIndex + 1

	// Swap items
	lc.Items[currentIndex], lc.Items[nextIndex] = lc.Items[nextIndex], lc.Items[currentIndex]

	// Update selection state
	if lc.MultiSelect {
		lc.updateSelectionAfterMove(currentIndex, nextIndex)
	}

	lc.SelectedIndex = nextIndex
	lc.ScrollTo(lc.SelectedIndex)

	if lc.OnReorder != nil {
		lc.OnReorder(currentIndex, nextIndex)
	}

	return true
}

// updateSelectionAfterMove adjusts selection state after moving items
func (lc *ListController) updateSelectionAfterMove(fromIdx, toIdx int) {
	switch {
	case lc.SelectedItems[fromIdx]:
		delete(lc.SelectedItems, fromIdx)
		lc.SelectedItems[toIdx] = true
	case lc.SelectedItems[toIdx]:
		delete(lc.SelectedItems, toIdx)
		lc.SelectedItems[fromIdx] = true
	}
}

// ToggleSelection toggles selection state of the item at given index
func (lc *ListController) ToggleSelection(index int) {
	if index < 0 || index >= len(lc.Items) {
		return
	}

	lc.Items[index].Selected = !lc.Items[index].Selected

	if lc.Items[index].Selected {
		lc.SelectedItems[index] = true
	} else {
		delete(lc.SelectedItems, index)
	}
}

// GetSelectedItems returns the indices of all selected items
func (lc *ListController) GetSelectedItems() []int {
	selectedIndices := make([]int, 0, len(lc.SelectedItems))
	for idx := range lc.SelectedItems {
		selectedIndices = append(selectedIndices, idx)
	}
	return selectedIndices
}

// ScrollTo adjusts visible window to ensure the item at index is visible
func (lc *ListController) ScrollTo(index int) {
	if index < 0 || index >= len(lc.Items) {
		return // Invalid index
	}

	// If already visible, do nothing
	if index >= lc.VisibleStartIndex && index < lc.VisibleStartIndex+lc.MaxVisibleItems {
		return
	}

	// Adjust visible window to make index visible
	if index < lc.VisibleStartIndex {
		lc.VisibleStartIndex = index
	} else {
		lc.VisibleStartIndex = index - lc.MaxVisibleItems + 1
		if lc.VisibleStartIndex < 0 {
			lc.VisibleStartIndex = 0
		}
	}
}

// HandleEvent processes input events and updates the list accordingly
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

// handleKeyDown processes keyboard input
func (lc *ListController) handleKeyDown(key sdl.Keycode) bool {
	lc.lastInputTime = time.Now()

	if key == sdl.K_h {
		lc.ToggleHelp()
		return true
	}

	if lc.ShowingHelp {
		return lc.handleHelpScreenInput(key)
	}

	if lc.ReorderMode {
		return lc.handleReorderModeInput(key)
	}

	return lc.handleNormalModeInput(key)
}

// handleHelpScreenInput handles input when help screen is shown
func (lc *ListController) handleHelpScreenInput(key sdl.Keycode) bool {
	switch key {
	case sdl.K_UP:
		lc.ScrollHelpOverlay(-1)
		return true
	case sdl.K_DOWN:
		lc.ScrollHelpOverlay(1)
		return true
	default:
		lc.ShowingHelp = false
		return true
	}
}

// handleReorderModeInput handles input in reorder mode
func (lc *ListController) handleReorderModeInput(key sdl.Keycode) bool {
	switch key {
	case sdl.K_UP:
		return lc.MoveItemUp()
	case sdl.K_DOWN:
		return lc.MoveItemDown()
	case sdl.K_ESCAPE, sdl.K_RETURN:
		lc.ReorderMode = false
		return true
	default:
		return false
	}
}

// handleNormalModeInput handles normal navigation and selection input
func (lc *ListController) handleNormalModeInput(key sdl.Keycode) bool {
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

// handleButtonPress processes controller button input
func (lc *ListController) handleButtonPress(button uint8) bool {
	lc.lastInputTime = time.Now()

	if button == BrickButton_MENU {
		lc.ToggleHelp()
		return true
	}

	if lc.ShowingHelp {
		return lc.handleHelpScreenButtonInput(button)
	}

	if lc.ReorderMode {
		return lc.handleReorderModeButtonInput(button)
	}

	return lc.handleNormalModeButtonInput(button)
}

// handleHelpScreenButtonInput handles button input when help is shown
func (lc *ListController) handleHelpScreenButtonInput(button uint8) bool {
	switch button {
	case BrickButton_UP:
		lc.ScrollHelpOverlay(-1)
		return true
	case BrickButton_DOWN:
		lc.ScrollHelpOverlay(1)
		return true
	default:
		return true
	}
}

// handleReorderModeButtonInput handles button input in reorder mode
func (lc *ListController) handleReorderModeButtonInput(button uint8) bool {
	switch button {
	case BrickButton_UP:
		return lc.MoveItemUp()
	case BrickButton_DOWN:
		return lc.MoveItemDown()
	case BrickButton_B, BrickButton_A:
		lc.ReorderMode = false
		return true
	default:
		return false
	}
}

// handleNormalModeButtonInput handles normal navigation and selection button input
func (lc *ListController) handleNormalModeButtonInput(button uint8) bool {
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
	default:
		return false
	}
}

// moveSelection changes the selected item by the given offset
func (lc *ListController) moveSelection(direction int) {
	if len(lc.Items) == 0 {
		return
	}

	// Clear current selection in single-select mode
	if !lc.MultiSelect {
		lc.Items[lc.SelectedIndex].Selected = false
		delete(lc.SelectedItems, lc.SelectedIndex)
	}

	// Calculate new index with wrap-around
	lc.SelectedIndex = (lc.SelectedIndex + direction + len(lc.Items)) % len(lc.Items)

	// Update selection in single-select mode
	if !lc.MultiSelect {
		lc.Items[lc.SelectedIndex].Selected = true
		lc.SelectedItems[lc.SelectedIndex] = true
	}

	// Ensure selected item is visible
	if lc.SelectedIndex < lc.VisibleStartIndex {
		lc.VisibleStartIndex = lc.SelectedIndex
	} else if lc.SelectedIndex >= lc.VisibleStartIndex+lc.MaxVisibleItems {
		lc.VisibleStartIndex = lc.SelectedIndex - lc.MaxVisibleItems + 1
	}
}

// Render draws the list to the screen
func (lc *ListController) Render(renderer *sdl.Renderer) {
	// Update scrolling animations
	lc.updateScrollingAnimations()

	// Update focus state for all items
	for i := range lc.Items {
		lc.Items[i].Focused = i == lc.SelectedIndex
	}

	// Get visible items
	endIndex := min(lc.VisibleStartIndex+lc.MaxVisibleItems, len(lc.Items))
	visibleItems := make([]models.MenuItem, endIndex-lc.VisibleStartIndex)
	copy(visibleItems, lc.Items[lc.VisibleStartIndex:endIndex])

	// Special handling for multi-select mode
	if lc.MultiSelect {
		for i := range visibleItems {
			visibleItems[i].Focused = false
		}

		// Only focus the currently selected item
		focusedIdx := lc.SelectedIndex - lc.VisibleStartIndex
		if focusedIdx >= 0 && focusedIdx < len(visibleItems) {
			visibleItems[focusedIdx].Focused = true
		}
	}

	// Store original title settings
	originalTitle := lc.Settings.Title
	originalAlign := lc.Settings.TitleAlign

	// Special handling for reorder mode
	if lc.ReorderMode {
		lc.Settings.Title = "REORDER MODE"
		lc.Settings.TitleAlign = internal.AlignCenter

		// Add reorder indicator to selected item
		selectedIdx := lc.SelectedIndex - lc.VisibleStartIndex
		if selectedIdx >= 0 && selectedIdx < len(visibleItems) {
			visibleItems[selectedIdx].Text = "↕ " + visibleItems[selectedIdx].Text
		}
	}

	// Draw the menu
	drawScrollableMenu(renderer, internal.GetFont(), visibleItems, lc.StartY, lc.Settings, lc.MultiSelect, lc)

	// Restore original title settings
	lc.Settings.Title = originalTitle
	lc.Settings.TitleAlign = originalAlign

	// Draw help overlay if active
	if lc.ShowingHelp && lc.helpOverlay != nil {
		lc.helpOverlay.Render(renderer, internal.GetSmallFont())
	}
}

// ToggleHelp shows or hides the help overlay
func (lc *ListController) ToggleHelp() {
	if !lc.HelpEnabled {
		return
	}

	if lc.helpOverlay == nil {
		lc.helpOverlay = NewHelpOverlay(defaultListHelpLines)
	}

	lc.helpOverlay.Toggle()
	lc.ShowingHelp = lc.helpOverlay.ShowingHelp
}

// ScrollHelpOverlay scrolls the help text up or down
func (lc *ListController) ScrollHelpOverlay(direction int) {
	if lc.helpOverlay != nil {
		lc.helpOverlay.Scroll(direction)
	}
}

// measureTextForScrolling measures text width to determine if scrolling is needed
func (lc *ListController) measureTextForScrolling(idx int, item models.MenuItem, maxWidth int32) *textScrollData {
	// Build full text with prefixes
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

	// Measure text width
	textSurface, err := internal.GetFont().RenderUTF8Blended(
		prefix+item.Text,
		sdl.Color{R: 255, G: 255, B: 255, A: 255},
	)
	if err != nil {
		return &textScrollData{}
	}
	defer textSurface.Free()

	textWidth := textSurface.W

	return &textScrollData{
		needsScrolling: textWidth > maxWidth,
		textWidth:      textWidth,
		containerWidth: maxWidth,
		direction:      scrollDirectionRight,
		scrollOffset:   0,
		lastUpdateTime: time.Now(),
		pauseCounter:   lc.Settings.ScrollPauseTime,
	}
}

// updateItemScrollAnimation updates a single item's scroll animation
func (lc *ListController) updateItemScrollAnimation(data *textScrollData) {
	// Calculate elapsed time
	currentTime := time.Now()
	elapsed := currentTime.Sub(data.lastUpdateTime).Seconds()
	data.lastUpdateTime = currentTime

	// Handle pause state
	if data.pauseCounter > 0 {
		data.pauseCounter--
		return
	}

	// Calculate scroll amount (at least 1 pixel)
	pixelsToScroll := max(int32(float32(elapsed)*lc.Settings.ScrollSpeed), 1)

	// Update position based on scroll direction
	if data.direction == scrollDirectionRight {
		// Scrolling right to left
		data.scrollOffset += pixelsToScroll

		// Check if we've reached the end
		if data.scrollOffset >= data.textWidth-data.containerWidth {
			data.scrollOffset = data.textWidth - data.containerWidth
			data.direction = scrollDirectionLeft
			data.pauseCounter = lc.Settings.ScrollPauseTime
		}
	} else {
		// Scrolling left to right
		data.scrollOffset -= pixelsToScroll

		// Check if we've reached the beginning
		if data.scrollOffset <= 0 {
			data.scrollOffset = 0
			data.direction = scrollDirectionRight
			data.pauseCounter = lc.Settings.ScrollPauseTime
		}
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

	const pillHeight = int32(60)
	screenWidth, _, err := renderer.GetOutputSize()
	if err != nil {
		screenWidth = 768 // fallback width
	}

	maxTextWidth := screenWidth - settings.Margins.Left - settings.Margins.Right - 15

	for i, item := range visibleItems {
		// Prepare text and color configuration
		textColor, bgColor := getItemColors(item, multiSelect)
		itemText := formatItemText(item, multiSelect)

		textSurface, err := font.RenderUTF8Blended(itemText, textColor)
		if err != nil {
			continue
		}
		defer textSurface.Free()

		textTexture, err := renderer.CreateTextureFromSurface(textSurface)
		if err != nil {
			continue
		}
		defer textTexture.Destroy()

		textWidth := textSurface.W
		textHeight := textSurface.H

		itemY := itemStartY + int32(i)*(pillHeight+settings.ItemSpacing)
		globalIndex := controller.VisibleStartIndex + i

		// Get scroll data if available
		scrollData, hasScrollData := controller.itemScrollData[globalIndex]
		needsScrolling := hasScrollData && scrollData.needsScrolling && item.Focused

		// Calculate pill width based on scrolling needs
		pillWidth := textWidth + 10
		if needsScrolling {
			pillWidth = maxTextWidth + 10
		}

		// Draw background for selected or focused items
		if item.Selected || item.Focused {
			pillRect := sdl.Rect{
				X: settings.Margins.Left,
				Y: itemY,
				W: pillWidth,
				H: pillHeight,
			}
			drawRoundedRect(renderer, &pillRect, 12, bgColor)
		}

		// Center text vertically within the pill
		textVerticalOffset := (pillHeight-textHeight)/2 + 1 // +1 for visual adjustment

		// Render text differently based on scrolling needs
		if needsScrolling {
			renderScrollingText(renderer, textTexture, textHeight, maxTextWidth, settings.Margins.Left,
				itemY, textVerticalOffset, scrollData.scrollOffset)
		} else {
			renderStaticText(renderer, textTexture, nil, textWidth, textHeight,
				settings.Margins.Left, itemY, textVerticalOffset)
		}
	}

	renderFooter(renderer, settings, controller.MultiSelect)
}

func getItemColors(item models.MenuItem, multiSelect bool) (textColor, bgColor sdl.Color) {
	if multiSelect {
		if item.Focused && item.Selected {
			return sdl.Color{R: 0, G: 0, B: 0, A: 255}, sdl.Color{R: 220, G: 220, B: 255, A: 255}
		} else if item.Focused {
			return sdl.Color{R: 255, G: 255, B: 255, A: 255}, sdl.Color{R: 100, G: 100, B: 180, A: 255}
		} else if item.Selected {
			return sdl.Color{R: 0, G: 0, B: 0, A: 255}, sdl.Color{R: 180, G: 180, B: 180, A: 255}
		}
		return sdl.Color{R: 255, G: 255, B: 255, A: 255}, sdl.Color{}
	}

	if item.Selected {
		return sdl.Color{R: 0, G: 0, B: 0, A: 255}, sdl.Color{R: 255, G: 255, B: 255, A: 255}
	} else if item.Focused {
		return sdl.Color{R: 255, G: 255, B: 255, A: 255}, sdl.Color{R: 100, G: 100, B: 180, A: 255}
	}
	return sdl.Color{R: 255, G: 255, B: 255, A: 255}, sdl.Color{}
}

func formatItemText(item models.MenuItem, multiSelect bool) string {
	if !multiSelect {
		return item.Text
	}

	if item.Selected {
		return "☑ " + item.Text
	}
	return "☐ " + item.Text
}

func renderScrollingText(renderer *sdl.Renderer, texture *sdl.Texture, textHeight, maxWidth, marginLeft,
	itemY, vertOffset, scrollOffset int32) {

	clipRect := &sdl.Rect{
		X: scrollOffset,
		Y: 0,
		W: maxWidth,
		H: textHeight,
	}

	textRect := sdl.Rect{
		X: marginLeft + 5,
		Y: itemY + vertOffset,
		W: maxWidth,
		H: textHeight,
	}

	renderer.Copy(texture, clipRect, &textRect)
}

func renderStaticText(renderer *sdl.Renderer, texture *sdl.Texture, src *sdl.Rect,
	width, height, marginLeft, itemY, vertOffset int32) {

	textRect := sdl.Rect{
		X: marginLeft + 5,
		Y: itemY + vertOffset,
		W: width,
		H: height,
	}
	renderer.Copy(texture, src, &textRect)
}

func renderFooter(renderer *sdl.Renderer, settings ListSettings, isMultiSelect bool) {
	if settings.FooterText == "" && !isMultiSelect {
		return
	}

	_, screenHeight, err := renderer.GetOutputSize()
	if err != nil {
		return
	}

	footerText := settings.FooterText
	if isMultiSelect {
		footerText = "A Add / Remove | Select Cancel | Start Confirm"
	}

	footerSurface, err := internal.GetSmallFont().RenderUTF8Blended(
		footerText,
		settings.FooterTextColor,
	)
	if err != nil {
		return
	}
	defer footerSurface.Free()

	footerTexture, err := renderer.CreateTextureFromSurface(footerSurface)
	if err != nil {
		return
	}
	defer footerTexture.Destroy()

	footerRect := sdl.Rect{
		X: settings.Margins.Left,
		Y: screenHeight - footerSurface.H - settings.Margins.Bottom,
		W: footerSurface.W,
		H: footerSurface.H,
	}

	renderer.Copy(footerTexture, nil, &footerRect)
}

func (lc *ListController) updateScrollingAnimations() {
	screenWidth, _, err := internal.GetWindow().Renderer.GetOutputSize()
	if err != nil {
		screenWidth = 768 // fallback width
	}

	maxTextWidth := screenWidth - lc.Settings.Margins.Left - lc.Settings.Margins.Right - 15

	// Update scrolling only for visible items
	endIdx := min(lc.VisibleStartIndex+lc.MaxVisibleItems, len(lc.Items))

	for idx := lc.VisibleStartIndex; idx < endIdx; idx++ {
		item := lc.Items[idx]

		// Only animate focused items
		if !item.Focused {
			delete(lc.itemScrollData, idx)
			continue
		}

		// Get or create scrolling data
		scrollData, exists := lc.itemScrollData[idx]
		if !exists {
			scrollData = lc.createScrollDataForItem(idx, item, maxTextWidth)
			lc.itemScrollData[idx] = scrollData
		}

		// Skip if no scrolling needed
		if !scrollData.needsScrolling {
			continue
		}

		lc.updateScrollAnimation(scrollData)
	}
}

func (lc *ListController) createScrollDataForItem(idx int, item models.MenuItem, maxWidth int32) *textScrollData {
	// Build the prefix based on item state
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

	// Measure text width
	textSurface, err := internal.GetFont().RenderUTF8Blended(
		prefix+item.Text,
		sdl.Color{R: 255, G: 255, B: 255, A: 255},
	)
	if err != nil {
		return &textScrollData{}
	}
	defer textSurface.Free()

	textWidth := textSurface.W

	return &textScrollData{
		needsScrolling: textWidth > maxWidth,
		textWidth:      textWidth,
		containerWidth: maxWidth,
		direction:      1, // Start scrolling left to right
		scrollOffset:   0,
		lastUpdateTime: time.Now(),
		pauseCounter:   lc.Settings.ScrollPauseTime, // Start with a pause
	}
}

func (lc *ListController) updateScrollAnimation(data *textScrollData) {
	// Only update at appropriate intervals
	currentTime := time.Now()
	elapsed := currentTime.Sub(data.lastUpdateTime).Seconds()
	data.lastUpdateTime = currentTime

	// Handle pause state
	if data.pauseCounter > 0 {
		data.pauseCounter--
		return
	}

	// Calculate scroll amount
	pixelsToScroll := max(int32(float32(elapsed)*lc.Settings.ScrollSpeed), 1)

	// Update scroll position based on direction
	if data.direction > 0 {
		// Scrolling right to left
		data.scrollOffset += pixelsToScroll

		// Check if we've reached the end
		if data.scrollOffset >= data.textWidth-data.containerWidth {
			data.scrollOffset = data.textWidth - data.containerWidth
			data.direction = -1 // Reverse direction
			data.pauseCounter = lc.Settings.ScrollPauseTime
		}
	} else {
		// Scrolling left to right
		data.scrollOffset -= pixelsToScroll

		// Check if we've reached the beginning
		if data.scrollOffset <= 0 {
			data.scrollOffset = 0
			data.direction = 1 // Reverse direction
			data.pauseCounter = lc.Settings.ScrollPauseTime
		}
	}
}

func drawTitle(renderer *sdl.Renderer, font *ttf.Font, title string, titleAlign internal.TextAlignment, startY int32, titleXMargin int32) int32 {
	titleSurface, err := font.RenderUTF8Blended(title, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err != nil {
		return startY
	}
	defer titleSurface.Free()

	titleTexture, err := renderer.CreateTextureFromSurface(titleSurface)
	if err != nil {
		return startY
	}
	defer titleTexture.Destroy()

	screenWidth, _, err := renderer.GetOutputSize()
	if err != nil {
		screenWidth = 768 // fallback width
	}

	// Calculate title position based on alignment
	titleX := getTitleXPosition(titleAlign, screenWidth, titleSurface.W, titleXMargin)

	titleRect := sdl.Rect{
		X: titleX,
		Y: startY,
		W: titleSurface.W,
		H: titleSurface.H,
	}
	renderer.Copy(titleTexture, nil, &titleRect)

	return titleSurface.H + 20
}

func getTitleXPosition(align internal.TextAlignment, screenWidth, titleWidth, margin int32) int32 {
	switch align {
	case internal.AlignCenter:
		return (screenWidth - titleWidth) / 2
	case internal.AlignRight:
		return screenWidth - titleWidth - margin
	default: // AlignLeft and fallback
		return margin
	}
}

func drawRoundedRect(renderer *sdl.Renderer, rect *sdl.Rect, radius int32, color sdl.Color) {
	// If no rounding needed, just draw a regular rectangle
	if radius <= 0 {
		renderer.SetDrawColor(color.R, color.G, color.B, color.A)
		renderer.FillRect(rect)
		return
	}

	// Limit radius to half of width/height
	radius = min(radius, min(rect.W/2, rect.H/2))

	renderer.SetDrawColor(color.R, color.G, color.B, color.A)

	// Draw center rectangle
	centerRect := &sdl.Rect{
		X: rect.X + radius,
		Y: rect.Y,
		W: rect.W - 2*radius,
		H: rect.H,
	}
	renderer.FillRect(centerRect)

	// Draw left side rectangle
	leftRect := &sdl.Rect{
		X: rect.X,
		Y: rect.Y + radius,
		W: radius,
		H: rect.H - 2*radius,
	}
	renderer.FillRect(leftRect)

	// Draw right side rectangle
	rightRect := &sdl.Rect{
		X: rect.X + rect.W - radius,
		Y: rect.Y + radius,
		W: radius,
		H: rect.H - 2*radius,
	}
	renderer.FillRect(rightRect)

	// Draw the four corners
	drawFilledCorner(renderer, rect.X+radius, rect.Y+radius, radius, 1, color)
	drawFilledCorner(renderer, rect.X+rect.W-radius, rect.Y+radius, radius, 2, color)
	drawFilledCorner(renderer, rect.X+radius, rect.Y+rect.H-radius, radius, 3, color)
	drawFilledCorner(renderer, rect.X+rect.W-radius, rect.Y+rect.H-radius, radius, 4, color)
}

func drawFilledCorner(renderer *sdl.Renderer, centerX, centerY, radius int32, corner int, color sdl.Color) {
	renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	radiusSquared := radius * radius

	// Get offset based on corner
	xOffset, yOffset := getCornerOffsets(corner)

	// Draw the corner using horizontal lines
	for dy := int32(0); dy <= radius; dy++ {
		width := int32(math.Sqrt(float64(radiusSquared - dy*dy)))
		if width <= 0 {
			continue
		}

		y := centerY + yOffset*dy
		var startX, endX int32

		if xOffset < 0 {
			startX = centerX - width
			endX = centerX
		} else {
			startX = centerX
			endX = centerX + width
		}

		renderer.DrawLine(startX, y, endX, y)
	}
}

func getCornerOffsets(corner int) (xOffset, yOffset int32) {
	switch corner {
	case 1: // Top-left
		return -1, -1
	case 2: // Top-right
		return 1, -1
	case 3: // Bottom-left
		return -1, 1
	case 4: // Bottom-right
		return 1, 1
	default:
		return 0, 0
	}
}
