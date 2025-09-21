package gabagool

import (
	"strings"
	"time"

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
	MultiSelectKey    sdl.Keycode
	MultiSelectButton Button
	ReorderKey        sdl.Keycode
	ReorderButton     Button

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
		MultiSelectKey:    sdl.K_SPACE,
		MultiSelectButton: ButtonSelect,
		ReorderKey:        sdl.K_SPACE,
		ReorderButton:     ButtonSelect,
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
		options.MaxVisibleItems = 9
	}

	lc := newListController(options)
	if options.SelectedIndex > 0 {
		lc.scrollTo(options.SelectedIndex)
	}

	running := true
	result := ListReturn{SelectedIndex: -1}

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				lc.handleInput(e, &running, &result)
			case *sdl.ControllerButtonEvent:
				lc.handleInput(e, &running, &result)
			}
		}

		lc.handleDirectionalRepeats()

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()
		renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

		window.RenderBackground()
		lc.render(renderer)
		renderer.Present()
		sdl.Delay(8)
	}

	result.Items = lc.Options.Items
	return option.Some(result), nil
}

func (lc *listController) handleInput(event interface{}, running *bool, result *ListReturn) {
	var keyPressed sdl.Keycode
	var buttonPressed Button
	var isKeyboard bool
	var isPressed bool

	switch e := event.(type) {
	case *sdl.KeyboardEvent:
		if e.Type != sdl.KEYDOWN {
			return
		}
		keyPressed = e.Keysym.Sym
		isKeyboard = true
		isPressed = true
	case *sdl.ControllerButtonEvent:
		isPressed = e.Type == sdl.CONTROLLERBUTTONDOWN
		if !isPressed && e.Type != sdl.CONTROLLERBUTTONUP {
			return
		}
		buttonPressed = Button(e.Button)
		result.LastPressedBtn = Button(e.Button)

		// Handle directional button holds
		lc.updateHeldDirections(buttonPressed, isPressed)
		if !isPressed {
			return
		}
	}

	// Handle help screen input separately
	if lc.ShowingHelp {
		lc.handleHelpInput(keyPressed, buttonPressed, isKeyboard)
		return
	}

	// Exit reorder mode on non-directional input
	if lc.ReorderMode && !lc.isDirectionalInput(keyPressed, buttonPressed, isKeyboard) {
		lc.ReorderMode = false
		return
	}

	// Handle navigation
	if lc.handleNavigation(keyPressed, buttonPressed, isKeyboard) {
		return
	}

	// Handle action buttons
	lc.handleActionButtons(keyPressed, buttonPressed, isKeyboard, running, result)
}

func (lc *listController) handleHelpInput(key sdl.Keycode, button Button, isKeyboard bool) {
	if isKeyboard {
		switch key {
		case sdl.K_UP:
			if lc.helpOverlay != nil {
				lc.helpOverlay.scroll(-1)
			}
		case sdl.K_DOWN:
			if lc.helpOverlay != nil {
				lc.helpOverlay.scroll(1)
			}
		case sdl.K_h:
			lc.ShowingHelp = false
		default:
			lc.ShowingHelp = false
		}
	} else {
		switch button {
		case ButtonUp:
			if lc.helpOverlay != nil {
				lc.helpOverlay.scroll(-1)
			}
		case ButtonDown:
			if lc.helpOverlay != nil {
				lc.helpOverlay.scroll(1)
			}
		case ButtonMenu:
			lc.ShowingHelp = false
		default:
			lc.ShowingHelp = false
		}
	}
}

func (lc *listController) isDirectionalInput(key sdl.Keycode, button Button, isKeyboard bool) bool {
	if isKeyboard {
		return key == sdl.K_UP || key == sdl.K_DOWN || key == sdl.K_LEFT || key == sdl.K_RIGHT
	}
	return button == ButtonUp || button == ButtonDown || button == ButtonLeft || button == ButtonRight
}

func (lc *listController) updateHeldDirections(button Button, pressed bool) {
	switch button {
	case ButtonUp:
		lc.heldDirections.up = pressed
	case ButtonDown:
		lc.heldDirections.down = pressed
	case ButtonLeft:
		lc.heldDirections.left = pressed
	case ButtonRight:
		lc.heldDirections.right = pressed
	}

	if pressed && len(lc.Options.Items) > 0 {
		lc.lastRepeatTime = time.Now()
	}
}

func (lc *listController) handleNavigation(key sdl.Keycode, button Button, isKeyboard bool) bool {
	if len(lc.Options.Items) == 0 {
		return false
	}

	direction := ""
	if isKeyboard {
		switch key {
		case sdl.K_UP:
			direction = "up"
		case sdl.K_DOWN:
			direction = "down"
		case sdl.K_LEFT:
			direction = "left"
		case sdl.K_RIGHT:
			direction = "right"
		}
	} else {
		switch button {
		case ButtonUp:
			direction = "up"
		case ButtonDown:
			direction = "down"
		case ButtonLeft:
			direction = "left"
		case ButtonRight:
			direction = "right"
		}
	}

	if direction != "" {
		lc.navigate(direction)
		return true
	}
	return false
}

func (lc *listController) handleActionButtons(key sdl.Keycode, button Button, isKeyboard bool, running *bool, result *ListReturn) {
	if len(lc.Options.Items) == 0 && key != sdl.K_b && button != ButtonB && key != sdl.K_h && button != ButtonMenu {
		return
	}

	// Primary action (A button / Enter key)
	if (isKeyboard && key == sdl.K_a) || (!isKeyboard && button == ButtonA) {
		if lc.MultiSelect && len(lc.Options.Items) > 0 {
			lc.toggleSelection(lc.Options.SelectedIndex)
		} else if len(lc.Options.Items) > 0 {
			*running = false
			result.populateSingleSelection(lc.Options.SelectedIndex, lc.Options.Items, lc.Options.VisibleStartIndex)
		}
	}

	// Back button
	if (isKeyboard && key == sdl.K_b) || (!isKeyboard && button == ButtonB) {
		if !lc.Options.DisableBackButton {
			*running = false
			result.SelectedIndex = -1
		}
	}

	// Action button (X)
	if (isKeyboard && key == sdl.K_x) || (!isKeyboard && button == ButtonX) {
		if lc.Options.EnableAction {
			*running = false
			result.ActionTriggered = true
		}
	}

	// Help
	if (isKeyboard && key == sdl.K_h) || (!isKeyboard && button == ButtonMenu) {
		if lc.Options.EnableHelp {
			lc.ShowingHelp = !lc.ShowingHelp
		}
	}

	// Multi-select confirmation (Enter/Start)
	if (isKeyboard && key == sdl.K_RETURN) || (!isKeyboard && button == ButtonStart) {
		if lc.MultiSelect && len(lc.Options.Items) > 0 {
			*running = false
			if indices := lc.getSelectedItems(); len(indices) > 0 {
				result.populateMultiSelection(indices, lc.Options.Items, lc.Options.VisibleStartIndex)
			}
		}
	}

	// Toggle multi-select mode
	if (isKeyboard && key == lc.Options.MultiSelectKey) || (!isKeyboard && button == lc.Options.MultiSelectButton) {
		if lc.Options.EnableMultiSelect && len(lc.Options.Items) > 0 {
			lc.toggleMultiSelect()
		}
	}

	// Toggle reorder mode
	if (isKeyboard && key == lc.Options.ReorderKey) || (!isKeyboard && button == lc.Options.ReorderButton) {
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

func (lc *listController) render(renderer *sdl.Renderer) {
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
	lc.renderContent(renderer, visibleItems)

	if lc.ShowingHelp && lc.helpOverlay != nil {
		lc.helpOverlay.ShowingHelp = true
		lc.helpOverlay.render(renderer, fonts.smallFont)
	}
}

func (lc *listController) renderContent(renderer *sdl.Renderer, visibleItems []MenuItem) {
	itemStartY := lc.StartY

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
	if lc.Options.EnableImages && lc.Options.SelectedIndex < len(lc.Options.Items) {
		selectedItem := lc.Options.Items[lc.Options.SelectedIndex]
		if selectedItem.ImageFilename != "" {
			lc.renderSelectedItemImage(renderer, selectedItem.ImageFilename)
		}
	}

	// Render footer
	renderFooter(renderer, fonts.smallFont, lc.Options.FooterHelpItems, lc.Options.Margins.Bottom, true)
}

func (lc *listController) renderItems(renderer *sdl.Renderer, font *ttf.Font, visibleItems []MenuItem, startY int32) {
	const pillHeight = int32(60)
	const pillPadding = int32(40)

	screenWidth, _, _ := renderer.GetOutputSize()
	availableWidth := screenWidth - lc.Options.Margins.Left - lc.Options.Margins.Right
	if lc.Options.EnableImages {
		availableWidth -= screenWidth / 7
	}

	maxPillWidth := availableWidth
	if lc.Options.EnableImages {
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
			drawRoundedRect(renderer, &pillRect, 30, bgColor)
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

func (lc *listController) renderScrollingText(renderer *sdl.Renderer, font *ttf.Font, text string, color sdl.Color, globalIndex int, itemY, pillHeight, maxWidth int32) {
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

	destRect := sdl.Rect{
		X: lc.Options.Margins.Left + 20,
		Y: itemY + (pillHeight-surface.H)/2,
		W: clipRect.W,
		H: surface.H,
	}

	renderer.Copy(texture, clipRect, &destRect)
}

func (lc *listController) renderStaticText(renderer *sdl.Renderer, font *ttf.Font, text string, color sdl.Color, itemY, pillHeight int32) {
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

	destRect := sdl.Rect{
		X: lc.Options.Margins.Left + 20,
		Y: itemY + (pillHeight-surface.H)/2,
		W: surface.W,
		H: surface.H,
	}

	renderer.Copy(texture, nil, &destRect)
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

func (lc *listController) renderSelectedItemImage(renderer *sdl.Renderer, imageFilename string) {
	texture, err := img.LoadTexture(renderer, imageFilename)
	if err != nil {
		return
	}
	defer texture.Destroy()

	_, _, textureWidth, textureHeight, _ := texture.Query()
	screenWidth, screenHeight, _ := renderer.GetOutputSize()

	maxImageWidth := screenWidth / 3
	maxImageHeight := screenHeight / 2

	scale := min32(maxImageWidth/textureWidth, maxImageHeight/textureHeight)
	imageWidth := textureWidth * scale
	imageHeight := textureHeight * scale

	destRect := sdl.Rect{
		X: screenWidth - imageWidth - 20,
		Y: (screenHeight - imageHeight) / 2,
		W: imageWidth,
		H: imageHeight,
	}

	renderer.Copy(texture, nil, &destRect)
}

func (lc *listController) renderScrollableTitle(renderer *sdl.Renderer, font *ttf.Font, title string, align TextAlign, startY, marginLeft int32) int32 {
	surface, _ := font.RenderUTF8Blended(title, GetTheme().ListTextColor)
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
		return GetTheme().ListTextSelectedColor, GetTheme().MainColor
	} else if item.Focused {
		return GetTheme().ListTextSelectedColor, GetTheme().MainColor
	} else if item.Selected {
		return GetTheme().ListTextColor, sdl.Color{R: 255, G: 0, B: 0, A: 0}
	}
	return GetTheme().ListTextColor, sdl.Color{}
}

func (lc *listController) getTextColor(focused bool) sdl.Color {
	if focused {
		return GetTheme().ListTextSelectedColor
	}
	return GetTheme().ListTextColor
}
