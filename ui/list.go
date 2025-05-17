package ui

import (
	"time"

	"github.com/UncleJunVIP/gabagool/internal"
	"github.com/UncleJunVIP/gabagool/models"
	"github.com/patrickhuber/go-types"
	"github.com/patrickhuber/go-types/option"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type ListOptions struct {
	Title             string
	Items             []models.MenuItem
	SelectedIndex     int
	VisibleStartIndex int
	MaxVisibleItems   int

	EnableAction      bool
	EnableMultiSelect bool
	EnableReordering  bool
	HelpEnabled       bool

	Margins         models.Padding
	ItemSpacing     int32
	TitleAlign      internal.TextAlignment
	TitleSpacing    int32
	FooterText      string
	FooterTextColor sdl.Color
	FooterHelpItems []FooterHelpItem

	ScrollSpeed     float32
	ScrollPauseTime int

	InputDelay        time.Duration
	MultiSelectKey    sdl.Keycode
	MultiSelectButton uint8
	ReorderKey        sdl.Keycode
	ReorderButton     uint8

	OnSelect  func(index int, item *models.MenuItem)
	OnReorder func(from, to int)
}

func DefaultListOptions(title string, items []models.MenuItem) ListOptions {
	return ListOptions{
		Title:             title,
		Items:             items,
		SelectedIndex:     0,
		MaxVisibleItems:   9,
		EnableAction:      false,
		EnableMultiSelect: false,
		EnableReordering:  false,
		HelpEnabled:       false,
		Margins:           models.UniformPadding(20),
		TitleAlign:        internal.AlignLeft,
		TitleSpacing:      internal.DefaultTitleSpacing,
		FooterText:        "",
		FooterTextColor:   sdl.Color{R: 180, G: 180, B: 180, A: 255},
		FooterHelpItems:   []FooterHelpItem{},
		ScrollSpeed:       1.0,
		ScrollPauseTime:   1000,
		InputDelay:        internal.DefaultInputDelay,
		MultiSelectKey:    sdl.K_SPACE,
		MultiSelectButton: BrickButton_SELECT,
		ReorderKey:        sdl.K_SPACE,
		ReorderButton:     BrickButton_SELECT,
		OnSelect:          nil,
		OnReorder:         nil,
	}
}

type textScrollData struct {
	needsScrolling      bool
	scrollOffset        int32
	textWidth           int32
	containerWidth      int32
	direction           int
	lastDirectionChange *time.Time
}

type listSettings struct {
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
	FooterHelpItems   []FooterHelpItem
	FooterTextColor   sdl.Color
}

type listController struct {
	Items             []models.MenuItem
	SelectedIndex     int
	SelectedItems     map[int]bool
	MultiSelect       bool
	EnableMultiSelect bool
	ReorderMode       bool
	EnableReorderMode bool
	Settings          listSettings
	StartY            int32
	lastInputTime     time.Time
	OnSelect          func(index int, item *models.MenuItem)

	VisibleStartIndex int
	MaxVisibleItems   int
	OnReorder         func(from, to int)

	EnableAction bool

	HelpEnabled bool
	helpOverlay *helpOverlay
	ShowingHelp bool

	itemScrollData map[int]*textScrollData

	heldDirections struct {
		up    bool
		down  bool
		left  bool
		right bool
	}
	lastRepeatTime time.Time
	repeatDelay    time.Duration
	repeatInterval time.Duration
}

func newListController(options ListOptions) *listController {
	selectedItems := make(map[int]bool)
	selectedIndex := options.SelectedIndex

	if selectedIndex < 0 || selectedIndex >= len(options.Items) {
		selectedIndex = 0
	}

	for i := range options.Items {
		options.Items[i].Selected = i == selectedIndex
		if options.Items[i].Selected {
			selectedItems[i] = true
		}
	}

	settings := listSettings{
		Margins:         options.Margins,
		ItemSpacing:     options.ItemSpacing,
		InputDelay:      options.InputDelay,
		Title:           options.Title,
		TitleAlign:      options.TitleAlign,
		TitleSpacing:    options.TitleSpacing,
		ScrollSpeed:     options.ScrollSpeed,
		ScrollPauseTime: options.ScrollPauseTime,
		FooterText:      options.FooterText,
		FooterTextColor: options.FooterTextColor,
		FooterHelpItems: options.FooterHelpItems,
	}

	if options.EnableMultiSelect {
		settings.MultiSelectKey = options.MultiSelectKey
		settings.MultiSelectButton = options.MultiSelectButton
	}

	if options.EnableReordering {
		settings.ReorderKey = options.ReorderKey
		settings.ReorderButton = options.ReorderButton
	}

	return &listController{
		Items:             options.Items,
		SelectedIndex:     selectedIndex,
		SelectedItems:     selectedItems,
		MultiSelect:       false,
		EnableMultiSelect: options.EnableMultiSelect,
		ReorderMode:       false,
		EnableReorderMode: options.EnableReordering,
		Settings:          settings,
		StartY:            20,
		lastInputTime:     time.Now(),
		VisibleStartIndex: options.VisibleStartIndex,
		MaxVisibleItems:   options.MaxVisibleItems,
		EnableAction:      options.EnableAction,
		HelpEnabled:       options.HelpEnabled,
		OnSelect:          options.OnSelect,
		OnReorder:         options.OnReorder,
		itemScrollData:    make(map[int]*textScrollData),
		lastRepeatTime:    time.Now(),
		repeatDelay:       200 * time.Millisecond,
		repeatInterval:    0 * time.Millisecond,
	}
}

// List presents a basic list of items to the user.
// Two specialty modes are provided: multi-select and reorder.
//   - Multi-select allows the user to select multiple items.
//   - Reorder allows the user to reorder the items in the list.
func List(options ListOptions) (types.Option[models.ListReturn], error) {
	window := internal.GetWindow()
	renderer := window.Renderer

	if options.MaxVisibleItems <= 0 {
		options.MaxVisibleItems = 9
	}

	listController := newListController(options)

	if options.SelectedIndex > 0 {
		listController.scrollTo(options.SelectedIndex)
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
				listController.handleKeyboardInput(e, &running, &result)
			case *sdl.ControllerButtonEvent:
				listController.handleControllerInput(e, &running, &result)
			}
		}

		listController.handleDirectionalRepeats()

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		window.RenderBackground()
		listController.render(renderer)

		renderer.Present()

		sdl.Delay(8)
	}

	if err != nil || result.Cancelled {
		return option.None[models.ListReturn](), err
	}

	return option.Some(result), nil
}

func (lc *listController) toggleMultiSelect() {
	lc.MultiSelect = !lc.MultiSelect

	if !lc.MultiSelect {
		for i := range lc.Items {
			lc.Items[i].Selected = false
		}

		lc.SelectedItems = make(map[int]bool)

		lc.Items[lc.SelectedIndex].Selected = true
		lc.SelectedItems[lc.SelectedIndex] = true
	}
}

func (lc *listController) toggleReorderMode() {
	lc.ReorderMode = !lc.ReorderMode
}

func (lc *listController) moveItemUp() bool {
	if !lc.ReorderMode || lc.SelectedIndex <= 0 {
		return false
	}

	currentIndex := lc.SelectedIndex
	prevIndex := currentIndex - 1

	lc.Items[currentIndex], lc.Items[prevIndex] = lc.Items[prevIndex], lc.Items[currentIndex]

	if lc.MultiSelect {
		lc.updateSelectionAfterMove(currentIndex, prevIndex)
	}

	lc.SelectedIndex = prevIndex
	lc.scrollTo(lc.SelectedIndex)

	if lc.OnReorder != nil {
		lc.OnReorder(currentIndex, prevIndex)
	}

	return true
}

func (lc *listController) moveItemDown() bool {
	if !lc.ReorderMode || lc.SelectedIndex >= len(lc.Items)-1 {
		return false
	}

	currentIndex := lc.SelectedIndex
	nextIndex := currentIndex + 1

	lc.Items[currentIndex], lc.Items[nextIndex] = lc.Items[nextIndex], lc.Items[currentIndex]

	if lc.MultiSelect {
		lc.updateSelectionAfterMove(currentIndex, nextIndex)
	}

	lc.SelectedIndex = nextIndex
	lc.scrollTo(lc.SelectedIndex)

	if lc.OnReorder != nil {
		lc.OnReorder(currentIndex, nextIndex)
	}

	return true
}

func (lc *listController) updateSelectionAfterMove(fromIdx, toIdx int) {
	switch {
	case lc.SelectedItems[fromIdx]:
		delete(lc.SelectedItems, fromIdx)
		lc.SelectedItems[toIdx] = true
	case lc.SelectedItems[toIdx]:
		delete(lc.SelectedItems, toIdx)
		lc.SelectedItems[fromIdx] = true
	}
}

func (lc *listController) toggleSelection(index int) {
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

func (lc *listController) getSelectedItems() []int {
	selectedIndices := make([]int, 0, len(lc.SelectedItems))
	for idx := range lc.SelectedItems {
		selectedIndices = append(selectedIndices, idx)
	}
	return selectedIndices
}

func (lc *listController) scrollTo(index int) {
	if index < 0 || index >= len(lc.Items) {
		return
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

func (lc *listController) handleKeyboardInput(e *sdl.KeyboardEvent, running *bool, result *models.ListReturn) {
	if e.Type != sdl.KEYDOWN {
		return
	}

	switch e.Keysym.Sym {
	case sdl.K_UP:
		lc.navigateUp()
	case sdl.K_DOWN:
		lc.navigateDown()
	case sdl.K_LEFT:
		lc.navigateLeft()
	case sdl.K_RIGHT:
		lc.navigateRight()

	case sdl.K_a:
		if lc.MultiSelect {
			lc.toggleSelection(lc.SelectedIndex)
		} else {
			*running = false
			result.PopulateSingleSelection(lc.SelectedIndex, lc.Items, lc.VisibleStartIndex)
			result.Cancelled = false
		}
	case sdl.K_b:
		*running = false
		result.SelectedIndex = -1
		result.Cancelled = true
	case sdl.K_x:
		if lc.EnableAction {
			*running = false
			result.ActionTriggered = true
			result.Cancelled = false
		}

	case sdl.K_h:
		if lc.HelpEnabled {
			lc.ShowingHelp = !lc.ShowingHelp
		}

	case sdl.K_RETURN:
		if lc.MultiSelect {
			*running = false
			if indices := lc.getSelectedItems(); len(indices) > 0 {
				result.PopulateMultiSelection(indices, lc.Items)
				result.Cancelled = false
			}
		}

	case lc.Settings.MultiSelectKey:
		if lc.EnableMultiSelect {
			lc.toggleMultiSelect()
		}
	case lc.Settings.ReorderKey:
		if lc.EnableReorderMode {
			lc.toggleReorderMode()
		}
	}
}

func (lc *listController) handleControllerInput(e *sdl.ControllerButtonEvent, running *bool, result *models.ListReturn) {
	if e.Type != sdl.CONTROLLERBUTTONDOWN && e.Type != sdl.CONTROLLERBUTTONUP {
		return
	}

	result.LastPressedBtn = e.Button

	switch e.Button {
	case BrickButton_UP:
		lc.heldDirections.up = e.Type == sdl.CONTROLLERBUTTONDOWN
		if e.Type == sdl.CONTROLLERBUTTONDOWN {

			lc.navigateUp()

			lc.lastRepeatTime = time.Now()
		}
	case BrickButton_DOWN:
		lc.heldDirections.down = e.Type == sdl.CONTROLLERBUTTONDOWN
		if e.Type == sdl.CONTROLLERBUTTONDOWN {

			lc.navigateDown()

			lc.lastRepeatTime = time.Now()
		}
	case BrickButton_LEFT:
		lc.heldDirections.left = e.Type == sdl.CONTROLLERBUTTONDOWN
		if e.Type == sdl.CONTROLLERBUTTONDOWN {

			lc.navigateLeft()

			lc.lastRepeatTime = time.Now()
		}
	case BrickButton_RIGHT:
		lc.heldDirections.right = e.Type == sdl.CONTROLLERBUTTONDOWN
		if e.Type == sdl.CONTROLLERBUTTONDOWN {

			lc.navigateRight()

			lc.lastRepeatTime = time.Now()
		}

	case BrickButton_A:
		if e.Type == sdl.CONTROLLERBUTTONDOWN {
			if lc.MultiSelect {
				lc.toggleSelection(lc.SelectedIndex)
			} else {
				*running = false
				result.PopulateSingleSelection(lc.SelectedIndex, lc.Items, lc.VisibleStartIndex)
				result.Cancelled = false
			}
		}
	case BrickButton_B:
		if e.Type == sdl.CONTROLLERBUTTONDOWN {
			*running = false
			result.SelectedIndex = -1
			result.Cancelled = true
		}
	case BrickButton_X:
		if lc.EnableAction && e.Type == sdl.CONTROLLERBUTTONDOWN {
			*running = false
			result.ActionTriggered = true
			result.Cancelled = false
		}

	case BrickButton_MENU:
		if lc.HelpEnabled && e.Type == sdl.CONTROLLERBUTTONDOWN {
			lc.ShowingHelp = !lc.ShowingHelp
		}

	case BrickButton_START:
		if lc.MultiSelect && e.Type == sdl.CONTROLLERBUTTONDOWN {
			*running = false
			if indices := lc.getSelectedItems(); len(indices) > 0 {
				result.PopulateMultiSelection(indices, lc.Items)
				result.Cancelled = false
			}
		}

	case lc.Settings.MultiSelectButton:
		if lc.EnableMultiSelect && e.Type == sdl.CONTROLLERBUTTONDOWN {
			lc.toggleMultiSelect()
		}
	case lc.Settings.ReorderButton:
		if lc.EnableReorderMode && e.Type == sdl.CONTROLLERBUTTONDOWN {
			lc.toggleReorderMode()
		}
	}
}

func (lc *listController) navigateUp() {
	if time.Since(lc.lastInputTime) < lc.Settings.InputDelay {
		return
	}

	lc.lastInputTime = time.Now()

	if lc.ReorderMode {
		lc.moveItemUp()
		return
	}

	if lc.SelectedIndex > 0 {
		lc.SelectedIndex--
	} else {

		lc.SelectedIndex = len(lc.Items) - 1

		if len(lc.Items) > lc.MaxVisibleItems {
			lc.VisibleStartIndex = len(lc.Items) - lc.MaxVisibleItems
		} else {
			lc.VisibleStartIndex = 0
		}
	}

	if !lc.MultiSelect {
		for i := range lc.Items {
			lc.Items[i].Selected = i == lc.SelectedIndex
		}
		lc.SelectedItems = map[int]bool{lc.SelectedIndex: true}
	}

	if lc.SelectedIndex < lc.VisibleStartIndex {

		lc.VisibleStartIndex = lc.SelectedIndex
	}
}

func (lc *listController) navigateDown() {
	if time.Since(lc.lastInputTime) < lc.Settings.InputDelay {
		return
	}

	lc.lastInputTime = time.Now()

	if lc.ReorderMode {
		lc.moveItemDown()
		return
	}

	if lc.SelectedIndex < len(lc.Items)-1 {
		lc.SelectedIndex++
	} else {

		lc.SelectedIndex = 0

		lc.VisibleStartIndex = 0
	}

	if !lc.MultiSelect {
		for i := range lc.Items {
			lc.Items[i].Selected = i == lc.SelectedIndex
		}
		lc.SelectedItems = map[int]bool{lc.SelectedIndex: true}
	}

	if lc.SelectedIndex >= lc.VisibleStartIndex+lc.MaxVisibleItems {

		lc.VisibleStartIndex = lc.SelectedIndex - lc.MaxVisibleItems + 1

		maxStartIndex := len(lc.Items) - lc.MaxVisibleItems
		if maxStartIndex < 0 {
			maxStartIndex = 0
		}
		if lc.VisibleStartIndex > maxStartIndex {
			lc.VisibleStartIndex = maxStartIndex
		}
	}
}

func (lc *listController) navigateLeft() {
	if time.Since(lc.lastInputTime) < lc.Settings.InputDelay {
		return
	}

	lc.lastInputTime = time.Now()

	visibleItems := lc.VisibleStartIndex + lc.MaxVisibleItems
	if visibleItems > len(lc.Items) {
		visibleItems = len(lc.Items)
	}

	if len(lc.Items) <= lc.MaxVisibleItems || lc.VisibleStartIndex == 0 {
		lc.SelectedIndex = 0
		lc.VisibleStartIndex = 0
	} else {

		skipAmount := lc.MaxVisibleItems

		newIndex := lc.VisibleStartIndex - skipAmount

		if newIndex < 0 {
			newIndex = 0
		}

		lc.SelectedIndex = newIndex

		lc.VisibleStartIndex = newIndex
	}

	if !lc.MultiSelect {
		for i := range lc.Items {
			lc.Items[i].Selected = i == lc.SelectedIndex
		}
		lc.SelectedItems = map[int]bool{lc.SelectedIndex: true}
	}
}

func (lc *listController) navigateRight() {
	if time.Since(lc.lastInputTime) < lc.Settings.InputDelay {
		return
	}

	lc.lastInputTime = time.Now()

	if len(lc.Items) <= lc.MaxVisibleItems {
		lc.SelectedIndex = len(lc.Items) - 1
	} else {

		skipAmount := lc.MaxVisibleItems

		newIndex := lc.VisibleStartIndex + skipAmount

		maxStartIndex := len(lc.Items) - lc.MaxVisibleItems
		if maxStartIndex < 0 {
			maxStartIndex = 0
		}

		if newIndex >= len(lc.Items) || newIndex > maxStartIndex {

			lc.SelectedIndex = len(lc.Items) - 1
			lc.VisibleStartIndex = maxStartIndex
		} else {

			lc.SelectedIndex = newIndex

			lc.VisibleStartIndex = newIndex
		}
	}

	if !lc.MultiSelect {
		for i := range lc.Items {
			lc.Items[i].Selected = i == lc.SelectedIndex
		}
		lc.SelectedItems = map[int]bool{lc.SelectedIndex: true}
	}
}

func (lc *listController) handleDirectionalRepeats() {

	if !lc.heldDirections.up && !lc.heldDirections.down &&
		!lc.heldDirections.left && !lc.heldDirections.right {

		lc.lastRepeatTime = time.Now()
		return
	}

	currentTime := time.Now()

	if time.Since(lc.lastRepeatTime) < lc.repeatDelay {

		return
	}

	if time.Since(lc.lastRepeatTime) >= lc.repeatInterval {
		lc.lastRepeatTime = currentTime

		if lc.heldDirections.up {
			lc.navigateUp()
		}
		if lc.heldDirections.down {
			lc.navigateDown()
		}
		if lc.heldDirections.left {
			lc.navigateLeft()
		}
		if lc.heldDirections.right {
			lc.navigateRight()
		}
	}
}

func (lc *listController) render(renderer *sdl.Renderer) {
	lc.updateScrollingAnimations()

	for i := range lc.Items {
		lc.Items[i].Focused = i == lc.SelectedIndex
	}

	endIndex := min(lc.VisibleStartIndex+lc.MaxVisibleItems, len(lc.Items))
	visibleItems := make([]models.MenuItem, endIndex-lc.VisibleStartIndex)
	copy(visibleItems, lc.Items[lc.VisibleStartIndex:endIndex])

	if lc.MultiSelect {
		for i := range visibleItems {
			visibleItems[i].Focused = false
		}

		focusedIdx := lc.SelectedIndex - lc.VisibleStartIndex
		if focusedIdx >= 0 && focusedIdx < len(visibleItems) {
			visibleItems[focusedIdx].Focused = true
		}
	}

	originalTitle := lc.Settings.Title
	originalAlign := lc.Settings.TitleAlign

	if lc.ReorderMode {
		lc.Settings.Title = "Reordering Mode"
		lc.Settings.TitleAlign = internal.AlignCenter

		selectedIdx := lc.SelectedIndex - lc.VisibleStartIndex
		if selectedIdx >= 0 && selectedIdx < len(visibleItems) {
			visibleItems[selectedIdx].Text = "↕ " + visibleItems[selectedIdx].Text
		}
	}

	drawScrollableMenu(renderer, internal.GetSmallFont(), visibleItems, lc.StartY, lc.Settings, lc.MultiSelect, lc)

	renderFooter(
		renderer,
		internal.GetSmallFont(),
		lc.Settings.FooterHelpItems,
		lc.Settings.Margins.Bottom,
	)

	lc.Settings.Title = originalTitle
	lc.Settings.TitleAlign = originalAlign

	if lc.ShowingHelp && lc.helpOverlay != nil {
		lc.helpOverlay.render(renderer, internal.GetSmallFont())
	}
}

func drawScrollableMenu(renderer *sdl.Renderer, font *ttf.Font, visibleItems []models.MenuItem,
	startY int32, settings listSettings, multiSelect bool, controller *listController) {

	if settings.Margins.Left <= 0 && settings.Margins.Right <= 0 &&
		settings.Margins.Top <= 0 && settings.Margins.Bottom <= 0 {
		settings.Margins = models.UniformPadding(10)
	}

	if settings.TitleSpacing <= 0 {
		settings.TitleSpacing = internal.DefaultTitleSpacing
	}

	itemStartY := startY

	if settings.Title != "" {
		itemStartY = drawTitle(renderer, internal.GetXLargeFont(), settings.Title,
			settings.TitleAlign, startY, settings.Margins.Left+10) + settings.TitleSpacing
	}

	const pillHeight = int32(60)
	screenWidth, _, err := renderer.GetOutputSize()
	if err != nil {
		screenWidth = 768
	}

	// Add more horizontal padding for text within pills (increased from 15 to 30)
	maxTextWidth := screenWidth - settings.Margins.Left - settings.Margins.Right - 30

	for i, item := range visibleItems {

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

		scrollData, hasScrollData := controller.itemScrollData[globalIndex]
		needsScrolling := hasScrollData && scrollData.needsScrolling && item.Focused

		// Increase horizontal padding within the pill (from 30 to 40)
		pillWidth := textWidth + 40
		if needsScrolling {
			pillWidth = maxTextWidth + 40
		}

		if item.Selected || item.Focused {
			pillRect := sdl.Rect{
				X: settings.Margins.Left,
				Y: itemY,
				W: pillWidth,
				H: pillHeight,
			}
			drawListRoundedRect(renderer, &pillRect, 30, bgColor)
		}

		textVerticalOffset := (pillHeight-textHeight)/2 + 1

		if needsScrolling {
			renderScrollingText(renderer, textTexture, textHeight, maxTextWidth, settings.Margins.Left,
				itemY, textVerticalOffset, scrollData.scrollOffset)
		} else {
			renderStaticText(renderer, textTexture, nil, textWidth, textHeight,
				settings.Margins.Left, itemY, textVerticalOffset)
		}
	}
}

func getItemColors(item models.MenuItem, multiSelect bool) (textColor, bgColor sdl.Color) {
	if multiSelect {
		if item.Focused && item.Selected {
			return internal.GetTheme().ListTextSelectedColor, internal.GetTheme().MainColor
		} else if item.Focused {
			return internal.GetTheme().ListTextSelectedColor, internal.GetTheme().MainColor
		} else if item.Selected {
			return internal.GetTheme().HintInfoColor, internal.GetTheme().PrimaryAccentColor
		}
		return internal.GetTheme().ListTextColor, sdl.Color{}
	}

	if item.Focused && item.Selected {
		return internal.GetTheme().ListTextSelectedColor, internal.GetTheme().MainColor
	} else if item.Focused {
		return internal.GetTheme().ListTextSelectedColor, internal.GetTheme().MainColor
	} else if item.Selected {
		return internal.GetTheme().ListTextSelectedColor, internal.GetTheme().PrimaryAccentColor
	}

	return internal.GetTheme().ListTextColor, sdl.Color{}
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

	// Get the full texture size
	_, _, fullWidth, _, err := texture.Query()
	if err != nil {
		return
	}

	// The display width will always be maxWidth (the container/pill width)
	displayWidth := maxWidth

	// Ensure the scrollOffset is never negative and never beyond the texture width minus display width
	if scrollOffset < 0 {
		scrollOffset = 0
	}

	maxOffset := fullWidth - displayWidth
	if maxOffset < 0 {
		maxOffset = 0
	}

	if scrollOffset > maxOffset {
		scrollOffset = maxOffset
	}

	// Create clip and destination rectangles
	clipRect := &sdl.Rect{
		X: scrollOffset,
		Y: 0,
		W: min(displayWidth, fullWidth-scrollOffset), // Ensure we don't try to display beyond texture bounds
		H: textHeight,
	}

	// Add horizontal padding
	textRect := sdl.Rect{
		X: marginLeft + 20, // Fixed horizontal padding
		Y: itemY + vertOffset,
		W: clipRect.W, // Match the width of what we're clipping
		H: textHeight,
	}

	renderer.Copy(texture, clipRect, &textRect)
}

func renderStaticText(renderer *sdl.Renderer, texture *sdl.Texture, src *sdl.Rect,
	width, height, marginLeft, itemY, vertOffset int32) {

	// For centering text within the pill, we need to know the pill width
	pillWidth := width + 30 // The pillWidth from the drawScrollableMenu function

	// Calculate position to center the text within the pill
	textX := marginLeft + (pillWidth-width)/2

	textRect := sdl.Rect{
		X: textX,
		Y: itemY + vertOffset,
		W: width,
		H: height,
	}
	renderer.Copy(texture, src, &textRect)
}

func drawListRoundedRect(renderer *sdl.Renderer, rect *sdl.Rect, radius int32, color sdl.Color) {
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

	mainRect := &sdl.Rect{
		X: rect.X + radius,
		Y: rect.Y,
		W: rect.W - 2*radius,
		H: rect.H,
	}
	renderer.FillRect(mainRect)

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

	for y := int32(0); y <= radius; y++ {
		for x := int32(0); x <= radius; x++ {

			if x*x+y*y <= radius*radius {

				renderer.DrawPoint(rect.X+radius-x, rect.Y+radius-y)

				renderer.DrawPoint(rect.X+rect.W-radius+x-1, rect.Y+radius-y)

				renderer.DrawPoint(rect.X+radius-x, rect.Y+rect.H-radius+y-1)

				renderer.DrawPoint(rect.X+rect.W-radius+x-1, rect.Y+rect.H-radius+y-1)
			}
		}
	}
}

func (lc *listController) updateScrollingAnimations() {
	screenWidth, _, err := internal.GetWindow().Renderer.GetOutputSize()
	if err != nil {
		screenWidth = 768
	}

	maxTextWidth := screenWidth - lc.Settings.Margins.Left - lc.Settings.Margins.Right - 30

	endIdx := min(lc.VisibleStartIndex+lc.MaxVisibleItems, len(lc.Items))

	for idx := lc.VisibleStartIndex; idx < endIdx; idx++ {
		item := lc.Items[idx]

		if !item.Focused {
			delete(lc.itemScrollData, idx)
			continue
		}

		scrollData, exists := lc.itemScrollData[idx]
		if !exists {
			scrollData = lc.createScrollDataForItem(idx, item, maxTextWidth)
			lc.itemScrollData[idx] = scrollData
		}

		if !scrollData.needsScrolling {
			continue
		}

		lc.updateScrollAnimation(scrollData)
	}
}

func (lc *listController) createScrollDataForItem(idx int, item models.MenuItem, maxWidth int32) *textScrollData {
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

	// Use a consistent font to ensure proper rendering
	textSurface, err := internal.GetSmallFont().RenderUTF8Blended(
		prefix+item.Text,
		sdl.Color{R: 255, G: 255, B: 255, A: 255},
	)
	if err != nil {
		return &textScrollData{}
	}
	defer textSurface.Free()

	textWidth := textSurface.W

	// Only enable scrolling if the text is actually wider than the available space
	needsScrolling := textWidth > maxWidth

	return &textScrollData{
		needsScrolling: needsScrolling,
		textWidth:      textWidth,
		containerWidth: maxWidth,
		direction:      1,
		scrollOffset:   0,
	}
}

func (lc *listController) updateScrollAnimation(data *textScrollData) {
	pixelsToScroll := lc.Settings.ScrollSpeed

	// Calculate the maximum valid scroll offset more accurately
	// Make sure we're using the actual rendered text width
	maxScrollOffset := (data.textWidth - data.containerWidth)

	if maxScrollOffset < 0 {
		maxScrollOffset = 0
	}

	// Add time tracking for direction changes
	currentTime := time.Now()
	if data.lastDirectionChange != nil &&
		currentTime.Sub(*data.lastDirectionChange).Milliseconds() < int64(lc.Settings.ScrollPauseTime) {
		return
	}

	if data.direction > 0 {
		// Scrolling right (forward)
		data.scrollOffset += int32(pixelsToScroll)

		// Make sure we don't scroll beyond the end of the text
		if data.scrollOffset >= maxScrollOffset {
			data.scrollOffset = maxScrollOffset
			data.direction = -1
			// Record time of direction change
			now := time.Now()
			data.lastDirectionChange = &now
		}
	} else {
		// Scrolling left (backward)
		data.scrollOffset -= int32(pixelsToScroll)

		// Make sure we don't scroll beyond the beginning of the text
		if data.scrollOffset <= 0 {
			data.scrollOffset = 0
			data.direction = 1
			// Record time of direction change
			now := time.Now()
			data.lastDirectionChange = &now
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
		screenWidth = 768
	}

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
	default:
		return margin
	}
}
