package gabagool

import (
	"strings"
	"time"

	"github.com/UncleJunVIP/gabagool/pkg/gabagool/core"
	"github.com/patrickhuber/go-types"
	"github.com/patrickhuber/go-types/option"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type ListOptions struct {
	Title             string
	Items             []MenuItem
	SelectedIndex     int
	VisibleStartIndex int
	MaxVisibleItems   int

	EnableAction      bool
	EnableMultiSelect bool
	EnableReordering  bool
	EnableHelp        bool
	EnableImages      bool

	StartInMultiSelectMode bool
	DisableBackButton      bool

	HelpTitle string
	HelpText  []string

	Margins         padding
	ItemSpacing     int32
	SmallTitle      bool
	TitleAlign      TextAlign
	TitleSpacing    int32
	FooterText      string
	FooterTextColor sdl.Color
	FooterHelpItems []FooterHelpItem

	ScrollSpeed     float32
	ScrollPauseTime int

	InputDelay        time.Duration
	MultiSelectButton InternalButton
	ReorderButton     InternalButton

	EmptyMessage      string
	EmptyMessageColor sdl.Color

	OnSelect  func(index int, item *MenuItem)
	OnReorder func(from, to int)
}

func DefaultListOptions(title string, items []MenuItem) ListOptions {
	return ListOptions{
		Title:             title,
		Items:             items,
		SelectedIndex:     0,
		MaxVisibleItems:   9,
		Margins:           uniformPadding(20),
		TitleAlign:        TextAlignLeft,
		TitleSpacing:      DefaultTitleSpacing,
		FooterTextColor:   sdl.Color{R: 180, G: 180, B: 180, A: 255},
		ScrollSpeed:       4.0,
		ScrollPauseTime:   1250,
		InputDelay:        DefaultInputDelay,
		MultiSelectButton: InternalButtonSelect, // New default
		ReorderButton:     InternalButtonSelect, // New default
		EmptyMessage:      "No items available",
		EmptyMessageColor: sdl.Color{R: 255, G: 255, B: 255, A: 255},
	}
}

type listController struct {
	Options       ListOptions
	SelectedItems map[int]bool
	MultiSelect   bool
	ReorderMode   bool
	ShowingHelp   bool
	StartY        int32
	lastInputTime time.Time

	helpOverlay     *helpOverlay
	itemScrollData  map[int]*textScrollData
	titleScrollData *textScrollData

	heldDirections struct {
		up, down, left, right bool
	}
	lastRepeatTime time.Time
	repeatDelay    time.Duration
	repeatInterval time.Duration
}

func newListController(options ListOptions) *listController {
	selectedItems := make(map[int]bool)
	if options.SelectedIndex < 0 || options.SelectedIndex >= len(options.Items) {
		options.SelectedIndex = 0
	}

	for i := range options.Items {
		if options.Items[i].Selected {
			selectedItems[i] = true
		}
	}

	var helpOverlay *helpOverlay
	if options.EnableHelp {
		helpOverlay = newHelpOverlay(options.HelpTitle, options.HelpText)
	}

	return &listController{
		Options:         options,
		SelectedItems:   selectedItems,
		MultiSelect:     options.StartInMultiSelectMode,
		StartY:          20,
		lastInputTime:   time.Now(),
		helpOverlay:     helpOverlay,
		itemScrollData:  make(map[int]*textScrollData),
		titleScrollData: &textScrollData{},
		lastRepeatTime:  time.Now(),
		repeatDelay:     150 * time.Millisecond,
		repeatInterval:  20 * time.Millisecond,
	}
}

func List(options ListOptions) (types.Option[ListReturn], error) {
	window := GetWindow()
	renderer := window.Renderer

	if options.MaxVisibleItems <= 0 {
		// This will be set dynamically based on window size
		options.MaxVisibleItems = 9 // fallback default
	}

	lc := newListController(options)

	// Calculate and set MaxVisibleItems based on window size
	lc.Options.MaxVisibleItems = int(lc.calculateMaxVisibleItems(window))

	if options.SelectedIndex > 0 {
		lc.scrollTo(options.SelectedIndex)
	}

	running := true
	result := ListReturn{SelectedIndex: -1}

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent, *sdl.ControllerButtonEvent, *sdl.ControllerAxisEvent, *sdl.JoyButtonEvent, *sdl.JoyAxisEvent, *sdl.JoyHatEvent:
				lc.handleInput(event, &running, &result)
			case *sdl.WindowEvent:
				// Handle window resize events
				we := event.(*sdl.WindowEvent)
				if we.Event == sdl.WINDOWEVENT_RESIZED {
					newMaxItems := lc.calculateMaxVisibleItems(window)
					lc.Options.MaxVisibleItems = int(newMaxItems)
					// Recalculate visibility if selection is out of bounds
					if lc.Options.SelectedIndex >= lc.Options.VisibleStartIndex+lc.Options.MaxVisibleItems {
						lc.scrollTo(lc.Options.SelectedIndex)
					}
				}
			}
		}

		lc.handleDirectionalRepeats()

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()
		renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

		lc.render(window)
		renderer.Present()
		sdl.Delay(8)
	}

	result.Items = lc.Options.Items
	return option.Some(result), nil
}

func (lc *listController) handleInput(event interface{}, running *bool, result *ListReturn) {
	processor := GetInputProcessor()

	// Process raw SDL event through the input processor
	inputEvent := processor.ProcessSDLEvent(event.(sdl.Event))
	if inputEvent == nil {
		return
	}

	if !inputEvent.Pressed {
		return
	}

	if lc.ShowingHelp {
		lc.handleHelpInput(inputEvent.Button)
		return
	}

	// Exit reorder mode on non-directional input
	if lc.ReorderMode && !lc.isDirectionalInput(inputEvent.Button) {
		lc.ReorderMode = false
		return
	}

	// Handle navigation
	if lc.handleNavigation(inputEvent.Button) {
		return
	}

	// Handle action buttons
	lc.handleActionButtons(inputEvent.Button, running, result)
}

func (lc *listController) handleHelpInput(button InternalButton) {
	switch button {
	case InternalButtonUp:
		if lc.helpOverlay != nil {
			lc.helpOverlay.scroll(-1)
		}
	case InternalButtonDown:
		if lc.helpOverlay != nil {
			lc.helpOverlay.scroll(1)
		}
	case InternalButtonMenu:
		lc.ShowingHelp = false
	default:
		lc.ShowingHelp = false
	}
}

func (lc *listController) isDirectionalInput(button InternalButton) bool {
	return button == InternalButtonUp || button == InternalButtonDown ||
		button == InternalButtonLeft || button == InternalButtonRight
}

func (lc *listController) updateHeldDirections(button InternalButton, pressed bool) {
	switch button {
	case InternalButtonUp:
		lc.heldDirections.up = pressed
	case InternalButtonDown:
		lc.heldDirections.down = pressed
	case InternalButtonLeft:
		lc.heldDirections.left = pressed
	case InternalButtonRight:
		lc.heldDirections.right = pressed
	}

	if pressed && len(lc.Options.Items) > 0 {
		lc.lastRepeatTime = time.Now()
	}
}

func (lc *listController) handleNavigation(button InternalButton) bool {
	if len(lc.Options.Items) == 0 {
		return false
	}

	direction := ""
	switch button {
	case InternalButtonUp:
		direction = "up"
	case InternalButtonDown:
		direction = "down"
	case InternalButtonLeft:
		direction = "left"
	case InternalButtonRight:
		direction = "right"
	}

	if direction != "" {
		lc.navigate(direction)
		return true
	}
	return false
}

func (lc *listController) handleActionButtons(button InternalButton, running *bool, result *ListReturn) {
	if len(lc.Options.Items) == 0 && button != InternalButtonB && button != InternalButtonMenu {
		return
	}

	// Primary action (A button)
	if button == InternalButtonA {
		if lc.MultiSelect && len(lc.Options.Items) > 0 {
			lc.toggleSelection(lc.Options.SelectedIndex)
		} else if len(lc.Options.Items) > 0 {
			*running = false
			result.populateSingleSelection(lc.Options.SelectedIndex, lc.Options.Items, lc.Options.VisibleStartIndex)
		}
	}

	// Back button
	if button == InternalButtonB {
		if !lc.Options.DisableBackButton {
			*running = false
			result.SelectedIndex = -1
		}
	}

	// Action button (X)
	if button == InternalButtonX {
		if lc.Options.EnableAction {
			*running = false
			result.ActionTriggered = true
			// New logic to handle returning the value(s) selected when X is pressed, if the app wants to use them
			// If not multi select, returns a single value like A button
			// if multi select, returns selected values like start button
			if len(lc.Options.Items) > 0 {
				if lc.MultiSelect {
					if indices := lc.getSelectedItems(); len(indices) > 0 {
						result.populateMultiSelection(indices, lc.Options.Items, lc.Options.VisibleStartIndex)
					}
				} else {
					result.populateSingleSelection(lc.Options.SelectedIndex, lc.Options.Items, lc.Options.VisibleStartIndex)
				}
			}
		}
	}

	// Help
	if button == InternalButtonMenu {
		if lc.Options.EnableHelp {
			lc.ShowingHelp = !lc.ShowingHelp
		}
	}

	// Multi-select confirmation (Start)
	if button == InternalButtonStart {
		if lc.MultiSelect && len(lc.Options.Items) > 0 {
			*running = false
			if indices := lc.getSelectedItems(); len(indices) > 0 {
				result.populateMultiSelection(indices, lc.Options.Items, lc.Options.VisibleStartIndex)
			}
		}
	}

	// Toggle multi-select mode
	if button == lc.Options.MultiSelectButton {
		if lc.Options.EnableMultiSelect && len(lc.Options.Items) > 0 {
			lc.toggleMultiSelect()
		}
	}

	// Toggle reorder mode
	if button == lc.Options.ReorderButton {
		if lc.Options.EnableReordering && len(lc.Options.Items) > 0 {
			lc.ReorderMode = !lc.ReorderMode
		}
	}
}

func (lc *listController) navigate(direction string) {
	if time.Since(lc.lastInputTime) < lc.Options.InputDelay {
		return
	}
	lc.lastInputTime = time.Now()

	switch direction {
	case "up":
		if lc.ReorderMode {
			lc.moveItem(-1)
		} else {
			lc.moveSelection(-1)
		}
	case "down":
		if lc.ReorderMode {
			lc.moveItem(1)
		} else {
			lc.moveSelection(1)
		}
	case "left":
		if lc.ReorderMode {
			lc.moveItem(-lc.Options.MaxVisibleItems)
		} else {
			lc.moveSelection(-lc.Options.MaxVisibleItems)
		}
	case "right":
		if lc.ReorderMode {
			lc.moveItem(lc.Options.MaxVisibleItems)
		} else {
			lc.moveSelection(lc.Options.MaxVisibleItems)
		}
	}
}

func (lc *listController) moveSelection(delta int) {
	newIndex := lc.Options.SelectedIndex + delta

	// Handle wrapping and page jumps
	if delta == 1 { // Down
		if newIndex >= len(lc.Options.Items) {
			newIndex = 0
			lc.Options.VisibleStartIndex = 0
		}
	} else if delta == -1 { // Up
		if newIndex < 0 {
			newIndex = len(lc.Options.Items) - 1
			if len(lc.Options.Items) > lc.Options.MaxVisibleItems {
				lc.Options.VisibleStartIndex = len(lc.Options.Items) - lc.Options.MaxVisibleItems
			}
		}
	} else { // Page jumps
		if delta > 0 { // Page right
			if len(lc.Options.Items) <= lc.Options.MaxVisibleItems {
				newIndex = len(lc.Options.Items) - 1
			} else {
				maxStart := len(lc.Options.Items) - lc.Options.MaxVisibleItems
				if lc.Options.VisibleStartIndex+lc.Options.MaxVisibleItems >= len(lc.Options.Items) {
					newIndex = len(lc.Options.Items) - 1
					lc.Options.VisibleStartIndex = maxStart
				} else {
					newIndex = min(lc.Options.VisibleStartIndex+lc.Options.MaxVisibleItems, len(lc.Options.Items)-1)
					lc.Options.VisibleStartIndex = newIndex
				}
			}
		} else { // Page left
			if lc.Options.VisibleStartIndex == 0 {
				newIndex = 0
			} else {
				newIndex = max(lc.Options.VisibleStartIndex-lc.Options.MaxVisibleItems, 0)
				lc.Options.VisibleStartIndex = newIndex
			}
		}
	}

	lc.Options.SelectedIndex = newIndex
	lc.scrollTo(newIndex)
	lc.updateSelectionState()
}

func (lc *listController) moveItem(delta int) {
	if delta == 1 && lc.Options.SelectedIndex >= len(lc.Options.Items)-1 {
		return
	}
	if delta == -1 && lc.Options.SelectedIndex <= 0 {
		return
	}

	// Handle multi-step moves for page jumps
	steps := delta
	if delta > 1 || delta < -1 {
		steps = delta / abs(delta) // Get direction
		targetIndex := lc.Options.SelectedIndex + delta
		targetIndex = max(0, min(targetIndex, len(lc.Options.Items)-1))

		// Move item step by step to target
		for lc.Options.SelectedIndex != targetIndex {
			if !lc.moveItemOneStep(steps) {
				break
			}
		}
		return
	}

	lc.moveItemOneStep(steps)
}

func (lc *listController) moveItemOneStep(direction int) bool {
	currentIndex := lc.Options.SelectedIndex
	var targetIndex int

	if direction > 0 {
		if currentIndex >= len(lc.Options.Items)-1 {
			return false
		}
		targetIndex = currentIndex + 1
	} else {
		if currentIndex <= 0 {
			return false
		}
		targetIndex = currentIndex - 1
	}

	// Swap items
	lc.Options.Items[currentIndex], lc.Options.Items[targetIndex] = lc.Options.Items[targetIndex], lc.Options.Items[currentIndex]

	// Update selection states
	if lc.MultiSelect {
		currentSelected := lc.SelectedItems[currentIndex]
		targetSelected := lc.SelectedItems[targetIndex]

		delete(lc.SelectedItems, currentIndex)
		delete(lc.SelectedItems, targetIndex)

		if currentSelected {
			lc.SelectedItems[targetIndex] = true
		}
		if targetSelected {
			lc.SelectedItems[currentIndex] = true
		}
	}

	lc.Options.SelectedIndex = targetIndex
	lc.scrollTo(targetIndex)

	if lc.Options.OnReorder != nil {
		lc.Options.OnReorder(currentIndex, targetIndex)
	}

	return true
}

func (lc *listController) toggleMultiSelect() {
	lc.MultiSelect = !lc.MultiSelect

	if !lc.MultiSelect {
		// Clear all selections
		for i := range lc.Options.Items {
			lc.Options.Items[i].Selected = false
		}
		lc.SelectedItems = make(map[int]bool)
	}

	lc.updateSelectionState()
}

func (lc *listController) toggleSelection(index int) {
	if index < 0 || index >= len(lc.Options.Items) || lc.Options.Items[index].NotMultiSelectable {
		return
	}

	lc.Options.Items[index].Selected = !lc.Options.Items[index].Selected
	if lc.Options.Items[index].Selected {
		lc.SelectedItems[index] = true
	} else {
		delete(lc.SelectedItems, index)
	}
}

func (lc *listController) updateSelectionState() {
	if !lc.MultiSelect {
		for i := range lc.Options.Items {
			lc.Options.Items[i].Selected = i == lc.Options.SelectedIndex
		}
		lc.SelectedItems = map[int]bool{lc.Options.SelectedIndex: true}
	}
	// In multi-select mode, don't automatically select items during navigation
	// Selection should only happen when 'A' button is explicitly pressed via toggleSelection()
}

func (lc *listController) getSelectedItems() []int {
	var indices []int
	for idx := range lc.SelectedItems {
		indices = append(indices, idx)
	}
	return indices
}

func (lc *listController) scrollTo(index int) {
	if index < lc.Options.VisibleStartIndex {
		lc.Options.VisibleStartIndex = index
	} else if index >= lc.Options.VisibleStartIndex+lc.Options.MaxVisibleItems {
		lc.Options.VisibleStartIndex = index - lc.Options.MaxVisibleItems + 1
		if lc.Options.VisibleStartIndex < 0 {
			lc.Options.VisibleStartIndex = 0
		}
	}
}

func (lc *listController) handleDirectionalRepeats() {
	if len(lc.Options.Items) == 0 || (!lc.heldDirections.up && !lc.heldDirections.down && !lc.heldDirections.left && !lc.heldDirections.right) {
		lc.lastRepeatTime = time.Now()
		return
	}

	if time.Since(lc.lastRepeatTime) < lc.repeatDelay {
		return
	}

	if time.Since(lc.lastRepeatTime) >= lc.repeatInterval {
		lc.lastRepeatTime = time.Now()

		if lc.heldDirections.up {
			lc.navigate("up")
		} else if lc.heldDirections.down {
			lc.navigate("down")
		} else if lc.heldDirections.left {
			lc.navigate("left")
		} else if lc.heldDirections.right {
			lc.navigate("right")
		}
	}
}

func (lc *listController) render(window *Window) {
	lc.updateScrolling()

	// Update focus states
	for i := range lc.Options.Items {
		lc.Options.Items[i].Focused = i == lc.Options.SelectedIndex
	}

	// Prepare visible items
	endIndex := min(lc.Options.VisibleStartIndex+lc.Options.MaxVisibleItems, len(lc.Options.Items))
	visibleItems := make([]MenuItem, endIndex-lc.Options.VisibleStartIndex)
	copy(visibleItems, lc.Options.Items[lc.Options.VisibleStartIndex:endIndex])

	// Add reorder indicator
	if lc.ReorderMode {
		selectedIdx := lc.Options.SelectedIndex - lc.Options.VisibleStartIndex
		if selectedIdx >= 0 && selectedIdx < len(visibleItems) {
			visibleItems[selectedIdx].Text = "↕ " + visibleItems[selectedIdx].Text
		}
	}

	// Render everything
	lc.renderContent(window, visibleItems)

	if lc.ShowingHelp && lc.helpOverlay != nil {
		lc.helpOverlay.ShowingHelp = true
		lc.helpOverlay.render(window.Renderer, fonts.smallFont)
	}
}

func (lc *listController) renderContent(window *Window, visibleItems []MenuItem) {
	renderer := window.Renderer

	itemStartY := lc.StartY

	// Render background
	if lc.Options.EnableImages && lc.Options.SelectedIndex < len(lc.Options.Items) {
		selectedItem := lc.Options.Items[lc.Options.SelectedIndex]
		if selectedItem.BackgroundFilename != "" {
			lc.renderSelectedItemBackground(window, selectedItem.BackgroundFilename)
		} else {
			window.RenderBackground()
		}
	} else {
		window.RenderBackground()
	}

	// Render title
	if lc.Options.Title != "" {
		titleFont := fonts.extraLargeFont
		if lc.Options.SmallTitle {
			titleFont = fonts.largeFont
		}
		itemStartY = lc.renderScrollableTitle(renderer, titleFont, lc.Options.Title, lc.Options.TitleAlign, lc.StartY, lc.Options.Margins.Left+10) + lc.Options.TitleSpacing
	}

	// Render items or empty message
	if len(lc.Options.Items) == 0 {
		lc.renderEmptyMessage(renderer, fonts.mediumFont, itemStartY)
	} else {
		lc.renderItems(renderer, fonts.smallFont, visibleItems, itemStartY)
	}

	// Render selected item image
	if lc.imageIsDisplayed() {
		lc.renderSelectedItemImage(renderer, lc.Options.Items[lc.Options.SelectedIndex].ImageFilename)
	}

	// Render footer
	renderFooter(renderer, fonts.smallFont, lc.Options.FooterHelpItems, lc.Options.Margins.Bottom, true)
}

func (lc *listController) imageIsDisplayed() bool {
	if lc.Options.EnableImages && lc.Options.SelectedIndex < len(lc.Options.Items) {
		selectedItem := lc.Options.Items[lc.Options.SelectedIndex]
		if selectedItem.ImageFilename != "" {
			return true
		}
	}
	return false
}

func (lc *listController) renderItems(renderer *sdl.Renderer, font *ttf.Font, visibleItems []MenuItem, startY int32) {
	scaleFactor := GetScaleFactor()

	pillHeight := int32(float32(60) * scaleFactor)
	pillPadding := int32(float32(40) * scaleFactor)

	screenWidth, _, _ := renderer.GetOutputSize()
	availableWidth := screenWidth - lc.Options.Margins.Left - lc.Options.Margins.Right
	if lc.imageIsDisplayed() {
		availableWidth -= screenWidth / 7
	}

	maxPillWidth := availableWidth
	if lc.imageIsDisplayed() {
		maxPillWidth = availableWidth * 3 / 4
	}
	maxTextWidth := maxPillWidth - pillPadding

	for i, item := range visibleItems {
		itemText := lc.formatItemText(item, lc.MultiSelect)
		itemY := startY + int32(i)*(pillHeight+lc.Options.ItemSpacing)
		globalIndex := lc.Options.VisibleStartIndex + i

		// Render background pill
		if item.Selected || item.Focused {
			_, bgColor := lc.getItemColors(item)
			pillWidth := min32(maxPillWidth, lc.measureText(font, itemText)+pillPadding)

			pillRect := sdl.Rect{
				X: lc.Options.Margins.Left,
				Y: itemY,
				W: pillWidth,
				H: pillHeight,
			}
			drawRoundedRect(renderer, &pillRect, int32(float32(30)*scaleFactor), bgColor)
		}

		// Render text (scrolling or static)
		lc.renderItemText(renderer, font, itemText, item.Focused, globalIndex, itemY, pillHeight, maxTextWidth)
	}
}

func (lc *listController) renderItemText(renderer *sdl.Renderer, font *ttf.Font, text string, focused bool, globalIndex int, itemY, pillHeight, maxWidth int32) {
	textColor := lc.getTextColor(focused)

	if focused && lc.shouldScroll(font, text, maxWidth) {
		lc.renderScrollingText(renderer, font, text, textColor, globalIndex, itemY, pillHeight, maxWidth)
	} else {
		truncatedText := lc.truncateText(font, text, maxWidth)
		lc.renderStaticText(renderer, font, truncatedText, textColor, itemY, pillHeight)
	}
}

func (lc *listController) renderStaticText(renderer *sdl.Renderer, font *ttf.Font, text string, color sdl.Color, itemY, pillHeight int32) {
	scaleFactor := GetScaleFactor()

	surface, _ := font.RenderUTF8Blended(text, color)
	if surface == nil {
		return
	}
	defer surface.Free()

	texture, _ := renderer.CreateTextureFromSurface(surface)
	if texture == nil {
		return
	}
	defer texture.Destroy()

	textPadding := int32(float32(20) * scaleFactor)
	destRect := sdl.Rect{
		X: lc.Options.Margins.Left + textPadding,
		Y: itemY + (pillHeight-surface.H)/2,
		W: surface.W,
		H: surface.H,
	}

	renderer.Copy(texture, nil, &destRect)
}

func (lc *listController) renderScrollingText(renderer *sdl.Renderer, font *ttf.Font, text string, color sdl.Color, globalIndex int, itemY, pillHeight, maxWidth int32) {
	scaleFactor := GetScaleFactor()
	scrollData := lc.getOrCreateScrollData(globalIndex, text, font, maxWidth)

	surface, _ := font.RenderUTF8Blended(text, color)
	if surface == nil {
		return
	}
	defer surface.Free()

	texture, _ := renderer.CreateTextureFromSurface(surface)
	if texture == nil {
		return
	}
	defer texture.Destroy()

	clipRect := &sdl.Rect{
		X: scrollData.scrollOffset,
		Y: 0,
		W: min32(maxWidth, surface.W-scrollData.scrollOffset),
		H: surface.H,
	}

	textPadding := int32(float32(20) * scaleFactor)
	destRect := sdl.Rect{
		X: lc.Options.Margins.Left + textPadding,
		Y: itemY + (pillHeight-surface.H)/2,
		W: clipRect.W,
		H: surface.H,
	}

	renderer.Copy(texture, clipRect, &destRect)
}

func (lc *listController) renderEmptyMessage(renderer *sdl.Renderer, font *ttf.Font, startY int32) {
	lines := strings.Split(lc.Options.EmptyMessage, "\n")
	screenWidth, screenHeight, _ := renderer.GetOutputSize()

	lineHeight := int32(25)
	totalHeight := int32(len(lines)) * lineHeight
	centerY := startY + (screenHeight-startY-lc.Options.Margins.Bottom-totalHeight)/2

	for i, line := range lines {
		if line == "" {
			continue
		}

		surface, _ := font.RenderUTF8Blended(line, lc.Options.EmptyMessageColor)
		if surface == nil {
			continue
		}

		texture, _ := renderer.CreateTextureFromSurface(surface)
		if texture == nil {
			surface.Free()
			continue
		}

		rect := sdl.Rect{
			X: (screenWidth - surface.W) / 2,
			Y: centerY + int32(i)*lineHeight,
			W: surface.W,
			H: surface.H,
		}

		renderer.Copy(texture, nil, &rect)
		texture.Destroy()
		surface.Free()
	}
}

func (lc *listController) renderSelectedItemBackground(window *Window, imageFilename string) {
	bgTexture, err := img.LoadTexture(window.Renderer, imageFilename)
	if err != nil {
		return
	}
	defer bgTexture.Destroy()
	window.Renderer.Copy(bgTexture, nil, &sdl.Rect{X: 0, Y: 0, W: window.GetWidth(), H: window.GetHeight()})
}

func (lc *listController) renderSelectedItemImage(renderer *sdl.Renderer, imageFilename string) {
	texture, err := img.LoadTexture(renderer, imageFilename)
	if err != nil {
		return
	}
	defer texture.Destroy()

	_, _, textureWidth, textureHeight, _ := texture.Query()
	screenWidth, screenHeight, _ := renderer.GetOutputSize()

	// Ensure texture has valid dimensions
	if textureWidth == 0 || textureHeight == 0 {
		return
	}

	maxImageWidth := screenWidth / 3
	maxImageHeight := screenHeight / 2

	// Calculate scale using float arithmetic to avoid integer division issues
	scaleX := float32(maxImageWidth) / float32(textureWidth)
	scaleY := float32(maxImageHeight) / float32(textureHeight)

	// Use the smaller scale to maintain aspect ratio
	scale := scaleX
	if scaleY < scaleX {
		scale = scaleY
	}

	imageWidth := int32(float32(textureWidth) * scale)
	imageHeight := int32(float32(textureHeight) * scale)

	// Ensure we have valid dimensions after scaling
	if imageWidth <= 0 || imageHeight <= 0 {
		return
	}

	destRect := sdl.Rect{
		X: screenWidth - imageWidth - 20,
		Y: (screenHeight - imageHeight) / 2,
		W: imageWidth,
		H: imageHeight,
	}

	renderer.Copy(texture, nil, &destRect)
}

func (lc *listController) renderScrollableTitle(renderer *sdl.Renderer, font *ttf.Font, title string, align TextAlign, startY, marginLeft int32) int32 {
	surface, _ := font.RenderUTF8Blended(title, core.GetTheme().ListTextColor)
	if surface == nil {
		return startY + 40
	}
	defer surface.Free()

	texture, _ := renderer.CreateTextureFromSurface(surface)
	if texture == nil {
		return startY + 40
	}
	defer texture.Destroy()

	screenWidth, _, _ := renderer.GetOutputSize()
	availableWidth := screenWidth - (marginLeft * 2)

	if surface.W > availableWidth {
		lc.renderScrollingTitle(renderer, texture, surface.H, availableWidth, marginLeft, startY)
	} else {
		var titleX int32
		switch align {
		case TextAlignCenter:
			titleX = (screenWidth - surface.W) / 2
		case TextAlignRight:
			titleX = screenWidth - surface.W - marginLeft
		default:
			titleX = marginLeft
		}

		rect := sdl.Rect{X: titleX, Y: startY, W: surface.W, H: surface.H}
		renderer.Copy(texture, nil, &rect)
	}

	return startY + surface.H
}

func (lc *listController) renderScrollingTitle(renderer *sdl.Renderer, texture *sdl.Texture, textHeight, maxWidth, titleX, titleY int32) {
	if !lc.titleScrollData.needsScrolling {
		_, _, fullWidth, _, _ := texture.Query()
		lc.titleScrollData.needsScrolling = true
		lc.titleScrollData.textWidth = fullWidth
		lc.titleScrollData.containerWidth = maxWidth
		lc.titleScrollData.direction = 1
	}

	clipRect := &sdl.Rect{
		X: max32(0, lc.titleScrollData.scrollOffset),
		Y: 0,
		W: min32(maxWidth, lc.titleScrollData.textWidth-lc.titleScrollData.scrollOffset),
		H: textHeight,
	}

	destRect := sdl.Rect{X: titleX, Y: titleY, W: clipRect.W, H: textHeight}
	renderer.Copy(texture, clipRect, &destRect)
}

func (lc *listController) updateScrolling() {
	currentTime := time.Now()

	if lc.titleScrollData.needsScrolling {
		lc.updateScrollData(lc.titleScrollData, currentTime)
	}

	for idx := lc.Options.VisibleStartIndex; idx < min(lc.Options.VisibleStartIndex+lc.Options.MaxVisibleItems, len(lc.Options.Items)); idx++ {
		if scrollData, exists := lc.itemScrollData[idx]; exists && scrollData.needsScrolling {
			lc.updateScrollData(scrollData, currentTime)
		}
	}
}

func (lc *listController) updateScrollData(data *textScrollData, currentTime time.Time) {
	if data.lastDirectionChange != nil && currentTime.Sub(*data.lastDirectionChange) < time.Duration(lc.Options.ScrollPauseTime)*time.Millisecond {
		return
	}

	scrollIncrement := int32(lc.Options.ScrollSpeed)
	data.scrollOffset += int32(data.direction) * scrollIncrement

	maxOffset := data.textWidth - data.containerWidth
	if data.scrollOffset <= 0 {
		data.scrollOffset = 0
		if data.direction < 0 {
			data.direction = 1
			now := currentTime
			data.lastDirectionChange = &now
		}
	} else if data.scrollOffset >= maxOffset {
		data.scrollOffset = maxOffset
		if data.direction > 0 {
			data.direction = -1
			now := currentTime
			data.lastDirectionChange = &now
		}
	}
}

func (lc *listController) getOrCreateScrollData(index int, text string, font *ttf.Font, maxWidth int32) *textScrollData {
	data, exists := lc.itemScrollData[index]
	if !exists {
		surface, _ := font.RenderUTF8Blended(text, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		if surface == nil {
			return &textScrollData{}
		}
		defer surface.Free()

		data = &textScrollData{
			needsScrolling: surface.W > maxWidth,
			textWidth:      surface.W,
			containerWidth: maxWidth,
			direction:      1,
		}
		lc.itemScrollData[index] = data
	}
	return data
}

func (lc *listController) shouldScroll(font *ttf.Font, text string, maxWidth int32) bool {
	surface, _ := font.RenderUTF8Blended(text, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if surface == nil {
		return false
	}
	defer surface.Free()
	return surface.W > maxWidth
}

func (lc *listController) calculateMaxVisibleItems(window *Window) int32 {
	scaleFactor := GetScaleFactor()

	pillHeight := int32(float32(60) * scaleFactor)

	_, screenHeight, _ := window.Renderer.GetOutputSize()

	var titleHeight int32 = 0
	if lc.Options.Title != "" {
		if lc.Options.SmallTitle {
			titleHeight = int32(float32(50) * scaleFactor)
		} else {
			titleHeight = int32(float32(60) * scaleFactor)
		}
		titleHeight += lc.Options.TitleSpacing
	}

	footerHeight := int32(float32(50) * scaleFactor)

	availableHeight := screenHeight - titleHeight - footerHeight - (lc.StartY * 2)

	itemHeightWithSpacing := pillHeight + lc.Options.ItemSpacing
	maxItems := availableHeight/itemHeightWithSpacing - 1

	if maxItems < 1 {
		maxItems = 1
	}

	return maxItems
}

func (lc *listController) getTitleFont() *ttf.Font {
	if lc.Options.SmallTitle {
		return fonts.largeFont
	}
	return fonts.extraLargeFont
}

func (lc *listController) measureText(font *ttf.Font, text string) int32 {
	surface, _ := font.RenderUTF8Blended(text, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if surface == nil {
		return 0
	}
	defer surface.Free()
	return surface.W
}

func (lc *listController) truncateText(font *ttf.Font, text string, maxWidth int32) string {
	if !lc.shouldScroll(font, text, maxWidth) {
		return text
	}

	ellipsis := "..."
	runes := []rune(text)
	for len(runes) > 5 {
		runes = runes[:len(runes)-1]
		testText := string(runes) + ellipsis
		if !lc.shouldScroll(font, testText, maxWidth) {
			return testText
		}
	}
	return ellipsis
}

func (lc *listController) formatItemText(item MenuItem, multiSelect bool) string {
	if !multiSelect || item.NotMultiSelectable {
		return item.Text
	}
	if item.Selected {
		return "☑ " + item.Text
	}
	return "☐ " + item.Text
}

func (lc *listController) getItemColors(item MenuItem) (textColor, bgColor sdl.Color) {
	if item.Focused && item.Selected {
		return core.GetTheme().ListTextSelectedColor, core.GetTheme().MainColor
	} else if item.Focused {
		return core.GetTheme().ListTextSelectedColor, core.GetTheme().MainColor
	} else if item.Selected {
		return core.GetTheme().ListTextColor, sdl.Color{R: 255, G: 0, B: 0, A: 0}
	}
	return core.GetTheme().ListTextColor, sdl.Color{}
}

func (lc *listController) getTextColor(focused bool) sdl.Color {
	if focused {
		return core.GetTheme().ListTextSelectedColor
	}
	return core.GetTheme().ListTextColor
}
