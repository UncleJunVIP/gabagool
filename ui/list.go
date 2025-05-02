package ui

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"nextui-sdl2/models"
	"time"
)

type ListSettings struct {
	ContentPadding     Padding       // Padding inside menu items
	Margins            Padding       // Outer margins of the entire menu
	ItemSpacing        int32         // Vertical spacing between menu items
	InputDelay         time.Duration // Delay between input processing
	Title              string        // Optional title text
	TitleAlign         TextAlignment // Title alignment (left, center, right)
	TitleSpacing       int32         // Space between title and first item
	MultiSelectKey     sdl.Keycode   // Key to toggle multi-select mode
	MultiSelectButton  uint8         // Controller button to toggle multi-select mode
	ReorderKey         sdl.Keycode   // Key to toggle reorder mode
	ReorderButton      uint8         // Controller button to toggle reorder mode
	ToggleSelectionKey sdl.Keycode   // Key to toggle selection in multi-select mode
	ToggleSelectionBtn uint8         // Controller button to toggle selection in multi-select mode
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

	HelpLines        []string
	ShowingHelp      bool
	HelpScrollOffset int32
	MaxHelpScroll    int32
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

	controller := &ListController{
		Items:         items,
		SelectedIndex: selectedIndex,
		SelectedItems: selectedItems,
		MultiSelect:   false,
		Settings:      settings,
		StartY:        startY,
		lastInputTime: time.Now(),
	}

	return controller
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

	// Help toggle
	if key == sdl.K_h || key == sdl.K_QUESTION {
		lc.ToggleHelp()
		return true
	}

	// Handle scrolling when help is showing
	if lc.ShowingHelp {
		if key == sdl.K_UP {
			lc.ScrollHelpOverlay(-1) // Scroll up
			return true
		}
		if key == sdl.K_DOWN {
			lc.ScrollHelpOverlay(1) // Scroll down
			return true
		}

		// Any other key dismisses help
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
		lc.moveSelection(4)
		return true
	case sdl.K_RIGHT:
		lc.moveSelection(4)
		return true
	case lc.Settings.MultiSelectKey:
		lc.ToggleMultiSelect()
		return true
	case lc.Settings.ToggleSelectionKey:
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
			lc.ScrollHelpOverlay(-1) // Scroll up
			return true
		}
		if button == BrickButton_DOWN {
			lc.ScrollHelpOverlay(1) // Scroll down
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
		lc.moveSelection(4)
		return true
	case BrickButton_RIGHT:
		lc.moveSelection(4)
		return true
	case lc.Settings.ToggleSelectionBtn:
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

	lc.RenderHelpPrompt(renderer)

	if lc.ShowingHelp {
		lc.RenderHelpOverlay(renderer)
	}
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

func (lc *ListController) ToggleHelp() {
	lc.ShowingHelp = !lc.ShowingHelp
	lc.HelpScrollOffset = 0 // Reset scroll position when toggling

	// Update help text based on current mode
	if lc.ShowingHelp {
		if lc.ReorderMode {
			lc.HelpLines = []string{
				"REORDER MODE",
				"↑/↓: Move item up/down",
				"Esc/B: Cancel reordering",
				"Enter/A: Confirm reordering",
				"H/?: Hide help",
			}
		} else if lc.MultiSelect {
			lc.HelpLines = []string{
				"MULTI-SELECT MODE",
				"↑/↓: Navigate list",
				"1/A: Toggle selection",
				"0/Y: Exit multi-select mode",
				"H/?: Hide help",
			}
		}
	}
}

func (lc *ListController) RenderHelpPrompt(renderer *sdl.Renderer) {
	if len(lc.HelpLines) == 0 {
		return
	}

	screenWidth, screenHeight, err := renderer.GetOutputSize()
	if err != nil {
		Logger.Error("Failed to get output size", "error", err)
		return
	}

	if !lc.ShowingHelp {
		font := GetFont()
		promptText := "Help (Menu)"

		promptSurface, err := font.RenderUTF8Blended(promptText, sdl.Color{R: 180, G: 180, B: 180, A: 200})
		if err != nil {
			return
		}

		promptTexture, err := renderer.CreateTextureFromSurface(promptSurface)
		if err != nil {
			promptSurface.Free()
			return
		}

		padding := int32(20)
		promptRect := sdl.Rect{
			X: screenWidth - promptSurface.W - padding,
			Y: screenHeight - promptSurface.H - padding,
			W: promptSurface.W,
			H: promptSurface.H,
		}

		renderer.Copy(promptTexture, nil, &promptRect)

		promptTexture.Destroy()
		promptSurface.Free()
	}
}

func (lc *ListController) RenderHelpOverlay(renderer *sdl.Renderer) {
	if !lc.ShowingHelp || len(lc.HelpLines) == 0 {
		return
	}

	// Get screen dimensions
	screenWidth, screenHeight, err := renderer.GetOutputSize()
	if err != nil {
		Logger.Error("Failed to get output size", "error", err)
		return
	}

	// Create semi-transparent overlay for the entire screen
	renderer.SetDrawColor(0, 0, 0, 200)
	overlay := sdl.Rect{X: 0, Y: 0, W: screenWidth, H: screenHeight}
	renderer.FillRect(&overlay)

	font := GetFont()

	// Pre-render title to get its dimensions
	titleText := "Help"
	titleSurface, err := font.RenderUTF8Blended(titleText, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err != nil {
		Logger.Error("Failed to render title", "error", err)
		return
	}
	defer titleSurface.Free()

	// Pre-render dismiss text to get its dimensions
	dismissText := "Press any key to dismiss (use ↑/↓ to scroll)"
	if lc.MaxHelpScroll == 0 {
		dismissText = "Press any key to dismiss"
	}

	dismissSurface, err := font.RenderUTF8Blended(dismissText, sdl.Color{R: 180, G: 180, B: 180, A: 255})
	if err != nil {
		Logger.Error("Failed to render dismiss text", "error", err)
		dismissSurface = nil
	}

	// Calculate dimensions for the dark blue content box
	boxWidth := int32(float32(screenWidth) * 0.85) // 85% of screen width

	// Calculate vertical spacing
	titleHeight := titleSurface.H
	titleSpacing := int32(30)      // Space between top of screen and title
	titleToBoxSpacing := int32(20) // Space between title and content box

	var dismissHeight, boxToHelpSpacing int32
	if dismissSurface != nil {
		dismissHeight = dismissSurface.H
		boxToHelpSpacing = int32(30) // INCREASED space between content box and help text
		defer dismissSurface.Free()
	} else {
		dismissHeight = 0
		boxToHelpSpacing = 0
	}

	// Add extra bottom margin
	bottomMargin := int32(40) // INCREASED bottom margin

	// Calculate maximum available height for content box
	maxBoxHeight := screenHeight - titleHeight - titleSpacing - titleToBoxSpacing - dismissHeight - boxToHelpSpacing - bottomMargin

	// Set box height to 90% of available height
	boxHeight := int32(float32(maxBoxHeight) * 0.9)

	// Center the box horizontally
	boxX := (screenWidth - boxWidth) / 2

	// Position the box vertically after the title
	titleY := titleSpacing
	boxY := titleY + titleHeight + titleToBoxSpacing

	// Draw the title
	titleTexture, err := renderer.CreateTextureFromSurface(titleSurface)
	if err == nil {
		titleX := (screenWidth - titleSurface.W) / 2 // Center title
		titleRect := sdl.Rect{X: titleX, Y: titleY, W: titleSurface.W, H: titleSurface.H}
		renderer.Copy(titleTexture, nil, &titleRect)
		titleTexture.Destroy()
	}

	// Draw the dark blue content box
	renderer.SetDrawColor(30, 30, 45, 255)
	contentBox := sdl.Rect{X: boxX, Y: boxY, W: boxWidth, H: boxHeight}
	renderer.FillRect(&contentBox)

	// Draw border for content box
	renderer.SetDrawColor(60, 60, 90, 255)
	renderer.DrawRect(&contentBox)

	// Content padding inside the dark blue box
	textPadding := int32(20)
	textX := boxX + textPadding
	textY := boxY + textPadding
	textWidth := boxWidth - (textPadding * 2) - int32(25) // Leave room for scrollbar
	textHeight := boxHeight - (textPadding * 2)

	// Calculate total content height
	lineHeight := int32(40) // Larger line height
	totalContentHeight := int32(0)

	// Pre-render all lines
	lineSurfaces := make([]*sdl.Surface, len(lc.HelpLines))
	for i, line := range lc.HelpLines {
		surface, err := font.RenderUTF8Blended(line, sdl.Color{R: 220, G: 220, B: 220, A: 255})
		if err != nil {
			continue
		}

		lineSurfaces[i] = surface
		totalContentHeight += lineHeight
	}

	// Calculate max scroll
	if totalContentHeight > textHeight {
		lc.MaxHelpScroll = totalContentHeight - textHeight
	} else {
		lc.MaxHelpScroll = 0
	}

	// Define viewport boundaries for manual clipping
	viewportTop := textY
	viewportBottom := textY + textHeight

	// Draw the scrollable content with manual clipping
	for i, surface := range lineSurfaces {
		if surface == nil {
			continue
		}

		// Calculate Y position with scroll offset
		yPos := textY - lc.HelpScrollOffset + (int32(i) * lineHeight)

		// Skip if completely outside the viewport
		if yPos+surface.H < viewportTop || yPos > viewportBottom {
			surface.Free()
			continue
		}

		texture, err := renderer.CreateTextureFromSurface(surface)
		if err != nil {
			surface.Free()
			continue
		}

		// For partial visibility, create a source rectangle that only shows
		// the visible portion of the text
		srcRect := &sdl.Rect{
			X: 0,
			Y: 0,
			W: surface.W,
			H: surface.H,
		}

		dstRect := &sdl.Rect{
			X: textX,
			Y: yPos,
			W: surface.W,
			H: surface.H,
		}

		// Handle partial visibility at top
		if yPos < viewportTop {
			diff := viewportTop - yPos
			srcRect.Y = diff
			srcRect.H = surface.H - diff
			dstRect.Y = viewportTop
			dstRect.H = srcRect.H
		}

		// Handle partial visibility at bottom
		if yPos+surface.H > viewportBottom {
			overlap := (yPos + surface.H) - viewportBottom
			srcRect.H = surface.H - overlap
			dstRect.H = srcRect.H
		}

		// Only render if there's something visible
		if srcRect.H > 0 {
			renderer.Copy(texture, srcRect, dstRect)
		}

		texture.Destroy()
		surface.Free()
	}

	// Draw scrollbar if needed
	if lc.MaxHelpScroll > 0 {
		scrollbarWidth := int32(15) // Wide scrollbar
		scrollbarHeight := textHeight
		scrollbarX := textX + textWidth + int32(5)
		scrollbarY := textY

		// Draw scrollbar background
		renderer.SetDrawColor(60, 60, 80, 200)
		scrollbarBg := sdl.Rect{
			X: scrollbarX,
			Y: scrollbarY,
			W: scrollbarWidth,
			H: scrollbarHeight,
		}
		renderer.FillRect(&scrollbarBg)

		// Calculate thumb size and position
		thumbRatio := float32(textHeight) / float32(totalContentHeight)
		if thumbRatio > 1.0 {
			thumbRatio = 1.0
		}

		thumbHeight := int32(float32(scrollbarHeight) * thumbRatio)
		if thumbHeight < 50 {
			thumbHeight = 50 // Minimum thumb size
		}

		scrollRatio := float32(lc.HelpScrollOffset) / float32(lc.MaxHelpScroll)
		thumbY := scrollbarY + int32(float32(scrollbarHeight-thumbHeight)*scrollRatio)

		// Draw the scrollbar thumb
		renderer.SetDrawColor(180, 180, 200, 255)
		scrollThumb := sdl.Rect{
			X: scrollbarX,
			Y: thumbY,
			W: scrollbarWidth,
			H: thumbHeight,
		}
		renderer.FillRect(&scrollThumb)
	}

	// Draw dismiss text below the content box
	if dismissSurface != nil {
		dismissTexture, err := renderer.CreateTextureFromSurface(dismissSurface)
		if err == nil {
			dismissX := (screenWidth - dismissSurface.W) / 2 // Center text
			dismissY := boxY + boxHeight + boxToHelpSpacing

			dismissRect := sdl.Rect{
				X: dismissX,
				Y: dismissY,
				W: dismissSurface.W,
				H: dismissSurface.H,
			}

			renderer.Copy(dismissTexture, nil, &dismissRect)
			dismissTexture.Destroy()
		}
	}
}

func (lc *ListController) ScrollHelpOverlay(direction int32) {
	scrollAmount := int32(30) // Scroll by this many pixels at once

	if direction < 0 { // Scroll up
		lc.HelpScrollOffset -= scrollAmount
		if lc.HelpScrollOffset < 0 {
			lc.HelpScrollOffset = 0
		}
	} else { // Scroll down
		lc.HelpScrollOffset += scrollAmount
		if lc.HelpScrollOffset > lc.MaxHelpScroll {
			lc.HelpScrollOffset = lc.MaxHelpScroll
		}
	}
}
