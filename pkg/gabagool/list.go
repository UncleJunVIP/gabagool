package gabagool

import (
	"github.com/veandco/go-sdl2/img"
	"strings"
	"time"

	"github.com/patrickhuber/go-types"
	"github.com/patrickhuber/go-types/option"
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

	// Add empty message customization
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
		EnableAction:      false,
		EnableMultiSelect: false,
		EnableReordering:  false,
		EnableHelp:        false,
		EnableImages:      false,
		Margins:           uniformPadding(20),
		TitleAlign:        AlignLeft,
		TitleSpacing:      DefaultTitleSpacing,
		FooterText:        "",
		FooterTextColor:   sdl.Color{R: 180, G: 180, B: 180, A: 255},
		FooterHelpItems:   []FooterHelpItem{},
		ScrollSpeed:       3.85,
		ScrollPauseTime:   1000,
		InputDelay:        DefaultInputDelay,
		MultiSelectKey:    sdl.K_SPACE,
		MultiSelectButton: ButtonSelect,
		ReorderKey:        sdl.K_SPACE,
		ReorderButton:     ButtonSelect,
		EmptyMessage:      "No items available",
		EmptyMessageColor: sdl.Color{R: 255, G: 255, B: 255, A: 255},
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
	Margins           padding
	ItemSpacing       int32
	InputDelay        time.Duration
	Title             string
	TitleAlign        TextAlign
	TitleSpacing      int32
	SmallTitle        bool
	MultiSelectKey    sdl.Keycode
	MultiSelectButton Button
	ReorderKey        sdl.Keycode
	ReorderButton     Button
	ScrollSpeed       float32
	ScrollPauseTime   int
	FooterText        string
	FooterHelpItems   []FooterHelpItem
	FooterTextColor   sdl.Color
	EmptyMessage      string
	EmptyMessageColor sdl.Color
	EnableImages      bool
}

type listController struct {
	Items             []MenuItem
	SelectedIndex     int
	SelectedItems     map[int]bool
	MultiSelect       bool
	EnableMultiSelect bool
	ReorderMode       bool
	EnableReorderMode bool
	Settings          listSettings
	StartY            int32
	lastInputTime     time.Time
	OnSelect          func(index int, item *MenuItem)

	VisibleStartIndex int
	MaxVisibleItems   int
	OnReorder         func(from, to int)

	EnableAction bool

	EnableHelp  bool
	helpOverlay *helpOverlay
	ShowingHelp bool

	itemScrollData  map[int]*textScrollData
	titleScrollData *textScrollData

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
		Margins:           options.Margins,
		ItemSpacing:       options.ItemSpacing,
		InputDelay:        options.InputDelay,
		Title:             options.Title,
		TitleAlign:        options.TitleAlign,
		TitleSpacing:      options.TitleSpacing,
		SmallTitle:        options.SmallTitle,
		ScrollSpeed:       options.ScrollSpeed,
		ScrollPauseTime:   options.ScrollPauseTime,
		FooterText:        options.FooterText,
		FooterTextColor:   options.FooterTextColor,
		FooterHelpItems:   options.FooterHelpItems,
		EmptyMessage:      options.EmptyMessage,
		EmptyMessageColor: options.EmptyMessageColor,
		EnableImages:      options.EnableImages,
	}

	if options.EnableMultiSelect {
		settings.MultiSelectKey = options.MultiSelectKey
		settings.MultiSelectButton = options.MultiSelectButton
	}

	if options.EnableReordering {
		settings.ReorderKey = options.ReorderKey
		settings.ReorderButton = options.ReorderButton
	}

	var helpOverlay *helpOverlay

	if options.EnableHelp {
		helpOverlay = newHelpOverlay(options.HelpTitle, options.HelpText)
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
		EnableHelp:        options.EnableHelp,
		helpOverlay:       helpOverlay,
		OnSelect:          options.OnSelect,
		OnReorder:         options.OnReorder,
		itemScrollData:    make(map[int]*textScrollData),
		titleScrollData:   &textScrollData{},
		lastRepeatTime:    time.Now(),
		repeatDelay:       0 * time.Millisecond,
		repeatInterval:    0 * time.Millisecond,
	}
}

// List presents a basic list of items to the user.
// Two specialty modes are provided: multi-select and reorder.
//   - Multi-select allows the user to select multiple items.
//   - Reorder allows the user to reorder the items in the list.
func List(options ListOptions) (types.Option[ListReturn], error) {
	window := GetWindow()
	renderer := window.Renderer

	if options.MaxVisibleItems <= 0 {
		options.MaxVisibleItems = 9
	}

	listController := newListController(options)

	if options.SelectedIndex > 0 {
		listController.scrollTo(options.SelectedIndex)
	}

	running := true
	result := ListReturn{
		SelectedIndex:  -1,
		SelectedItem:   nil,
		LastPressedBtn: 0,
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

	if err != nil {
		return option.None[ListReturn](), err
	}

	result.Items = listController.Items
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
	} else {
		// When entering multi-select mode, only select the currently highlighted item
		// if it's not marked as NotMultiSelectable
		if !lc.Items[lc.SelectedIndex].NotMultiSelectable {
			lc.Items[lc.SelectedIndex].Selected = true
			lc.SelectedItems[lc.SelectedIndex] = true
		} else {
			// Make sure item is not selected when it's NotMultiSelectable
			lc.Items[lc.SelectedIndex].Selected = false
			delete(lc.SelectedItems, lc.SelectedIndex)
		}
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

	// Skip toggling selection if the item is marked as not multi-selectable
	if lc.Items[index].NotMultiSelectable {
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

func (lc *listController) handleKeyboardInput(e *sdl.KeyboardEvent, running *bool, result *ListReturn) {
	if e.Type != sdl.KEYDOWN {
		return
	}

	if lc.ShowingHelp {
		lc.ShowingHelp = false
		return
	}

	if e.Keysym.Sym != sdl.K_UP && e.Keysym.Sym != sdl.K_DOWN &&
		e.Keysym.Sym != sdl.K_LEFT && e.Keysym.Sym != sdl.K_RIGHT &&
		lc.ReorderMode {
		lc.toggleReorderMode()
		return
	}

	switch e.Keysym.Sym {
	case sdl.K_UP:
		if len(lc.Items) > 0 {
			lc.navigateUp()
		}
	case sdl.K_DOWN:
		if len(lc.Items) > 0 {
			lc.navigateDown()
		}
	case sdl.K_LEFT:
		if len(lc.Items) > 0 {
			lc.navigateLeft()
		}
	case sdl.K_RIGHT:
		if len(lc.Items) > 0 {
			lc.navigateRight()
		}

	case sdl.K_a:
		if len(lc.Items) > 0 {
			if lc.MultiSelect {
				lc.toggleSelection(lc.SelectedIndex)
			} else {
				*running = false
				result.populateSingleSelection(lc.SelectedIndex, lc.Items, lc.VisibleStartIndex)
			}
		}
	case sdl.K_b:
		*running = false
		result.SelectedIndex = -1
	case sdl.K_x:
		if lc.EnableAction {
			*running = false
			result.ActionTriggered = true
		}

	case sdl.K_h:
		if lc.EnableHelp {
			lc.ShowingHelp = !lc.ShowingHelp
		}

	case sdl.K_RETURN:
		if lc.MultiSelect && len(lc.Items) > 0 {
			*running = false
			if indices := lc.getSelectedItems(); len(indices) > 0 {
				result.populateMultiSelection(indices, lc.Items, lc.VisibleStartIndex)
			}
		}

	case lc.Settings.MultiSelectKey:
		if lc.EnableMultiSelect && len(lc.Items) > 0 {
			lc.toggleMultiSelect()
		}
	case lc.Settings.ReorderKey:
		if lc.EnableReorderMode && len(lc.Items) > 0 {
			lc.toggleReorderMode()
		}
	}
}

func (lc *listController) handleControllerInput(e *sdl.ControllerButtonEvent, running *bool, result *ListReturn) {
	if e.Type != sdl.CONTROLLERBUTTONDOWN && e.Type != sdl.CONTROLLERBUTTONUP {
		return
	}

	result.LastPressedBtn = Button(e.Button)

	if lc.ShowingHelp && e.Type == sdl.CONTROLLERBUTTONDOWN {
		lc.ShowingHelp = false
		return
	}

	if Button(e.Button) != ButtonUp && Button(e.Button) != ButtonDown &&
		Button(e.Button) != ButtonLeft && Button(e.Button) != ButtonRight &&
		lc.ReorderMode && e.Type == sdl.CONTROLLERBUTTONDOWN {
		lc.toggleReorderMode()
		return
	}

	switch Button(e.Button) {
	case ButtonUp:
		lc.heldDirections.up = e.Type == sdl.CONTROLLERBUTTONDOWN
		if e.Type == sdl.CONTROLLERBUTTONDOWN && len(lc.Items) > 0 {
			lc.navigateUp()
			lc.lastRepeatTime = time.Now()
		}
	case ButtonDown:
		lc.heldDirections.down = e.Type == sdl.CONTROLLERBUTTONDOWN
		if e.Type == sdl.CONTROLLERBUTTONDOWN && len(lc.Items) > 0 {
			lc.navigateDown()
			lc.lastRepeatTime = time.Now()
		}
	case ButtonLeft:
		lc.heldDirections.left = e.Type == sdl.CONTROLLERBUTTONDOWN
		if e.Type == sdl.CONTROLLERBUTTONDOWN && len(lc.Items) > 0 {
			lc.navigateLeft()
			lc.lastRepeatTime = time.Now()
		}
	case ButtonRight:
		lc.heldDirections.right = e.Type == sdl.CONTROLLERBUTTONDOWN
		if e.Type == sdl.CONTROLLERBUTTONDOWN && len(lc.Items) > 0 {
			lc.navigateRight()
			lc.lastRepeatTime = time.Now()
		}

	case ButtonA:
		if e.Type == sdl.CONTROLLERBUTTONDOWN && len(lc.Items) > 0 {
			if lc.MultiSelect {
				lc.toggleSelection(lc.SelectedIndex)
			} else {
				*running = false
				result.populateSingleSelection(lc.SelectedIndex, lc.Items, lc.VisibleStartIndex)
			}
		}
	case ButtonB:
		if e.Type == sdl.CONTROLLERBUTTONDOWN {
			*running = false
			result.SelectedIndex = -1
		}
	case ButtonX:
		if lc.EnableAction && e.Type == sdl.CONTROLLERBUTTONDOWN {
			*running = false
			result.ActionTriggered = true
		}

	case ButtonMenu:
		if lc.EnableHelp && e.Type == sdl.CONTROLLERBUTTONDOWN {
			lc.ShowingHelp = !lc.ShowingHelp
		}

	case ButtonStart:
		if lc.MultiSelect && e.Type == sdl.CONTROLLERBUTTONDOWN && len(lc.Items) > 0 {
			*running = false
			if indices := lc.getSelectedItems(); len(indices) > 0 {
				result.populateMultiSelection(indices, lc.Items, lc.VisibleStartIndex)
			}
		}

	case lc.Settings.MultiSelectButton:
		if lc.EnableMultiSelect && e.Type == sdl.CONTROLLERBUTTONDOWN && len(lc.Items) > 0 {
			lc.toggleMultiSelect()
		}
	case lc.Settings.ReorderButton:
		if lc.EnableReorderMode && e.Type == sdl.CONTROLLERBUTTONDOWN && len(lc.Items) > 0 {
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
	// Don't handle repeats if there are no items
	if len(lc.Items) == 0 {
		lc.lastRepeatTime = time.Now()
		return
	}

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
	visibleItems := make([]MenuItem, endIndex-lc.VisibleStartIndex)
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
		selectedIdx := lc.SelectedIndex - lc.VisibleStartIndex
		if selectedIdx >= 0 && selectedIdx < len(visibleItems) {
			visibleItems[selectedIdx].Text = "↕ " + visibleItems[selectedIdx].Text
		}
	}

	drawScrollableMenu(renderer, fonts.smallFont, visibleItems, lc.StartY, lc.Settings, lc.MultiSelect, lc)

	if lc.Settings.EnableImages && lc.SelectedIndex < len(lc.Items) {
		selectedItem := lc.Items[lc.SelectedIndex]
		if selectedItem.ImageFilename != "" {
			lc.renderSelectedItemImage(renderer, selectedItem.ImageFilename)
		}
	}

	renderFooter(
		renderer,
		fonts.smallFont,
		lc.Settings.FooterHelpItems,
		lc.Settings.Margins.Bottom,
		true,
	)

	lc.Settings.Title = originalTitle
	lc.Settings.TitleAlign = originalAlign

	if lc.ShowingHelp && lc.helpOverlay != nil {
		lc.helpOverlay.ShowingHelp = true
		lc.helpOverlay.render(renderer, fonts.smallFont)
	}
}

func (lc *listController) renderSelectedItemImage(renderer *sdl.Renderer, imageFilename string) {
	texture, err := img.LoadTexture(renderer, imageFilename)
	if err != nil {
		// Silently fail if image cannot be loaded
		return
	}
	defer texture.Destroy()

	_, _, textureWidth, textureHeight, err := texture.Query()
	if err != nil {
		return
	}

	screenWidth, screenHeight, err := renderer.GetOutputSize()
	if err != nil {
		return
	}

	imageWidth := textureWidth
	imageHeight := textureHeight

	maxImageWidth := screenWidth / 3
	maxImageHeight := screenHeight / 2

	if imageWidth > maxImageWidth || imageHeight > maxImageHeight {
		widthScale := float64(maxImageWidth) / float64(imageWidth)
		heightScale := float64(maxImageHeight) / float64(imageHeight)

		scale := widthScale
		if heightScale < widthScale {
			scale = heightScale
		}

		imageWidth = int32(float64(imageWidth) * scale)
		imageHeight = int32(float64(imageHeight) * scale)
	}

	imageX := screenWidth - imageWidth - 20
	imageY := (screenHeight - imageHeight) / 2

	destRect := sdl.Rect{
		X: imageX,
		Y: imageY,
		W: imageWidth,
		H: imageHeight,
	}

	renderer.Copy(texture, nil, &destRect)
}

func drawScrollableMenu(renderer *sdl.Renderer, font *ttf.Font, visibleItems []MenuItem,
	startY int32, settings listSettings, multiSelect bool, controller *listController) {

	if settings.Margins.Left <= 0 && settings.Margins.Right <= 0 &&
		settings.Margins.Top <= 0 && settings.Margins.Bottom <= 0 {
		settings.Margins = uniformPadding(20)
	}

	if settings.TitleSpacing <= 0 {
		settings.TitleSpacing = DefaultTitleSpacing
	}

	itemStartY := startY

	if settings.Title != "" {
		titleFont := fonts.extraLargeFont

		if settings.SmallTitle {
			titleFont = fonts.largeFont
		}

		// Use scrollable title function instead of regular drawTitle
		itemStartY = drawScrollableTitle(renderer, titleFont, settings.Title,
			settings.TitleAlign, startY, settings.Margins.Left+10, controller) + settings.TitleSpacing
	}

	if len(controller.Items) == 0 {
		drawEmptyListMessage(renderer, fonts.mediumFont, itemStartY, settings)
		return
	}

	const pillHeight = int32(60)
	const pillPadding = int32(40) // Horizontal padding inside pill

	screenWidth, _, err := renderer.GetOutputSize()
	if err != nil {
		screenWidth = 768
	}

	// Calculate available width for text, always considering image display if enabled
	availableWidth := screenWidth - settings.Margins.Left - settings.Margins.Right
	if settings.EnableImages {
		imageReservedWidth := screenWidth / 7
		availableWidth -= imageReservedWidth
	}

	// Set maximum pill width based on whether image display is enabled
	maxPillWidth := availableWidth
	if settings.EnableImages {
		// When image display is enabled, limit pill width to be shorter
		maxPillWidth = availableWidth * 3 / 4 // Use 3/4 of available width for shorter pills
	}

	// Calculate max text width based on the pill width constraint
	maxTextWidth := maxPillWidth - pillPadding // Account for horizontal padding

	for i, item := range visibleItems {

		textColor, bgColor := getItemColors(item, multiSelect)
		itemText := formatItemText(item, multiSelect)

		// For unfocused items, truncate the text to fit within the pill width
		displayText := itemText
		if !item.Focused {
			displayText = truncateTextToWidth(font, itemText, maxTextWidth)
		}

		textSurface, err := font.RenderUTF8Blended(displayText, textColor)
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

		// Calculate pill width based on text width when images are enabled
		var pillWidth int32
		if settings.EnableImages {
			// For image display mode, use text width + padding, but don't exceed maxPillWidth
			pillWidth = textWidth + pillPadding
			if pillWidth > maxPillWidth {
				pillWidth = maxPillWidth
			}
		} else {
			// For non-image display mode, use full available width
			pillWidth = availableWidth
		}

		// Update scroll data to use the new container width
		if hasScrollData {
			scrollData.containerWidth = pillWidth - pillPadding
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
			// Use original itemText for scrolling, not truncated version
			originalTextSurface, err := font.RenderUTF8Blended(itemText, textColor)
			if err != nil {
				continue
			}
			defer originalTextSurface.Free()

			originalTextTexture, err := renderer.CreateTextureFromSurface(originalTextSurface)
			if err != nil {
				continue
			}
			defer originalTextTexture.Destroy()

			renderScrollingText(renderer, originalTextTexture, originalTextSurface.H, pillWidth-pillPadding, settings.Margins.Left,
				itemY, textVerticalOffset, scrollData.scrollOffset)
		} else {
			renderStaticText(renderer, textTexture, nil, textWidth, textHeight,
				settings.Margins.Left, itemY, textVerticalOffset)
		}
	}
}

func truncateTextToWidth(font *ttf.Font, text string, maxWidth int32) string {
	if text == "" {
		return text
	}

	surface, err := font.RenderUTF8Blended(text, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err != nil {
		return text
	}
	defer surface.Free()

	if surface.W <= maxWidth {
		return text
	}

	ellipsis := "..."
	runes := []rune(text)
	left, right := 0, len(runes)

	for left < right {
		mid := (left + right + 1) / 2
		testText := string(runes[:mid]) + ellipsis

		testSurface, err := font.RenderUTF8Blended(testText, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		if err != nil {
			right = mid - 1
			continue
		}
		defer testSurface.Free()

		if testSurface.W <= maxWidth {
			left = mid
		} else {
			right = mid - 1
		}
	}

	if left > 0 {
		return string(runes[:left]) + ellipsis
	}
	return ellipsis
}

func drawEmptyListMessage(renderer *sdl.Renderer, font *ttf.Font, startY int32, settings listSettings) {
	emptyMessage := settings.EmptyMessage
	if emptyMessage == "" {
		emptyMessage = "No items available"
	}

	lines := strings.Split(emptyMessage, "\n")

	screenWidth, screenHeight, err := renderer.GetOutputSize()
	if err != nil {
		screenWidth = 768
		screenHeight = 1024
	}

	tempSurface, err := font.RenderUTF8Blended("Test", settings.EmptyMessageColor)
	if err != nil {
		return
	}
	lineHeight := tempSurface.H + 5
	tempSurface.Free()

	// Calculate total height of all text lines
	totalTextHeight := int32(len(lines)) * lineHeight

	availableHeight := screenHeight - startY - settings.Margins.Bottom - 30
	centerY := startY + (availableHeight-totalTextHeight)/2

	currentY := centerY

	for _, line := range lines {
		if line == "" {
			currentY += lineHeight
			continue
		}

		textSurface, err := font.RenderUTF8Blended(line, settings.EmptyMessageColor)
		if err != nil {
			currentY += lineHeight
			continue
		}

		textTexture, err := renderer.CreateTextureFromSurface(textSurface)
		if err != nil {
			textSurface.Free()
			currentY += lineHeight
			continue
		}

		messageX := (screenWidth - textSurface.W) / 2

		messageRect := sdl.Rect{
			X: messageX,
			Y: currentY,
			W: textSurface.W,
			H: textSurface.H,
		}
		renderer.Copy(textTexture, nil, &messageRect)

		textTexture.Destroy()
		textSurface.Free()

		currentY += lineHeight
	}
}

func getItemColors(item MenuItem, multiSelect bool) (textColor, bgColor sdl.Color) {
	if multiSelect {
		if item.Focused && item.Selected {
			return GetTheme().ListTextSelectedColor, GetTheme().MainColor
		} else if item.Focused {
			return GetTheme().ListTextSelectedColor, GetTheme().MainColor
		} else if item.Selected {
			return GetTheme().HintInfoColor, GetTheme().PrimaryAccentColor
		}
		return GetTheme().ListTextColor, sdl.Color{}
	}

	if item.Focused && item.Selected {
		return GetTheme().ListTextSelectedColor, GetTheme().MainColor
	} else if item.Focused {
		return GetTheme().ListTextSelectedColor, GetTheme().MainColor
	} else if item.Selected {
		return GetTheme().ListTextSelectedColor, GetTheme().PrimaryAccentColor
	}

	return GetTheme().ListTextColor, sdl.Color{}
}

func formatItemText(item MenuItem, multiSelect bool) string {
	if !multiSelect || item.NotMultiSelectable {
		return item.Text
	}

	if item.Selected {
		return "☑ " + item.Text
	}
	return "☐ " + item.Text
}

func renderScrollingText(renderer *sdl.Renderer, texture *sdl.Texture, textHeight, maxWidth, marginLeft,
	itemY, vertOffset, scrollOffset int32) {

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
	currentTime := time.Now()

	if lc.titleScrollData.needsScrolling {
		lc.updateTextScrolling(lc.titleScrollData, currentTime, lc.Settings.ScrollSpeed, lc.Settings.ScrollPauseTime)
	}

	screenWidth, _, err := GetWindow().Renderer.GetOutputSize()
	if err != nil {
		screenWidth = 768
	}

	maxTextWidth := screenWidth - lc.Settings.Margins.Left - lc.Settings.Margins.Right - 30

	if lc.Settings.EnableImages {
		maxTextWidth = maxTextWidth - 200
	}

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

func (lc *listController) updateTextScrolling(scrollData *textScrollData, currentTime time.Time, scrollSpeed float32, pauseTime int) {
	if !scrollData.needsScrolling {
		return
	}

	// Check if we need to pause at the edges
	if scrollData.lastDirectionChange != nil {
		if currentTime.Sub(*scrollData.lastDirectionChange) < time.Duration(pauseTime)*time.Millisecond {
			return
		}
	}

	// Calculate scroll increment based on speed
	scrollIncrement := int32(scrollSpeed)
	if scrollIncrement < 1 {
		scrollIncrement = 1
	}

	// Update scroll position
	scrollData.scrollOffset += int32(scrollData.direction) * scrollIncrement

	// Check bounds and reverse direction if needed
	maxOffset := scrollData.textWidth - scrollData.containerWidth
	if maxOffset < 0 {
		maxOffset = 0
	}

	if scrollData.scrollOffset <= 0 {
		scrollData.scrollOffset = 0
		if scrollData.direction < 0 {
			scrollData.direction = 1
			now := currentTime
			scrollData.lastDirectionChange = &now
		}
	} else if scrollData.scrollOffset >= maxOffset {
		scrollData.scrollOffset = maxOffset
		if scrollData.direction > 0 {
			scrollData.direction = -1
			now := currentTime
			scrollData.lastDirectionChange = &now
		}
	}
}

func (lc *listController) createScrollDataForItem(idx int, item MenuItem, maxWidth int32) *textScrollData {
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
	textSurface, err := fonts.smallFont.RenderUTF8Blended(
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

func drawScrollableTitle(renderer *sdl.Renderer, font *ttf.Font, title string, align TextAlign,
	startY, marginLeft int32, controller *listController) int32 {
	if title == "" {
		return startY
	}

	textSurface, err := font.RenderUTF8Blended(title, GetTheme().ListTextColor)
	if err != nil {
		return startY + 40
	}
	defer textSurface.Free()

	textTexture, err := renderer.CreateTextureFromSurface(textSurface)
	if err != nil {
		return startY + 40
	}
	defer textTexture.Destroy()

	screenWidth, _, err := renderer.GetOutputSize()
	if err != nil {
		screenWidth = 768
	}

	textWidth := textSurface.W
	textHeight := textSurface.H

	// Calculate available width for title (with margins)
	availableWidth := screenWidth - (marginLeft * 2)

	// Check if title needs scrolling
	needsScrolling := textWidth > availableWidth

	if needsScrolling {
		// Initialize scroll data if needed
		if !controller.titleScrollData.needsScrolling {
			controller.titleScrollData.needsScrolling = true
			controller.titleScrollData.textWidth = textWidth
			controller.titleScrollData.containerWidth = availableWidth
			controller.titleScrollData.scrollOffset = 0
			controller.titleScrollData.direction = 1
			controller.titleScrollData.lastDirectionChange = nil
		}

		// Update container width in case screen size changed
		controller.titleScrollData.containerWidth = availableWidth

		// Render scrolling title
		var titleX int32
		switch align {
		case AlignLeft:
			titleX = marginLeft
		case AlignCenter:
			titleX = marginLeft
		case AlignRight:
			titleX = marginLeft
		default:
			titleX = marginLeft
		}

		renderScrollingTitle(renderer, textTexture, textHeight, availableWidth, titleX,
			startY, controller.titleScrollData.scrollOffset)
	} else {
		// Reset scroll data if title no longer needs scrolling
		controller.titleScrollData.needsScrolling = false

		// Render static title
		var titleX int32
		switch align {
		case AlignLeft:
			titleX = marginLeft
		case AlignCenter:
			titleX = (screenWidth - textWidth) / 2
		case AlignRight:
			titleX = screenWidth - textWidth - marginLeft
		default:
			titleX = marginLeft
		}

		titleRect := sdl.Rect{
			X: titleX,
			Y: startY,
			W: textWidth,
			H: textHeight,
		}

		renderer.Copy(textTexture, nil, &titleRect)
	}

	return startY + textHeight
}

func renderScrollingTitle(renderer *sdl.Renderer, texture *sdl.Texture, textHeight, maxWidth,
	titleX, titleY, scrollOffset int32) {

	// Get the full texture size
	_, _, fullWidth, _, err := texture.Query()
	if err != nil {
		return
	}

	// Ensure the scrollOffset is within bounds
	if scrollOffset < 0 {
		scrollOffset = 0
	}

	maxOffset := fullWidth - maxWidth
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
		W: min(maxWidth, fullWidth-scrollOffset),
		H: textHeight,
	}

	titleRect := sdl.Rect{
		X: titleX,
		Y: titleY,
		W: clipRect.W,
		H: textHeight,
	}

	renderer.Copy(texture, clipRect, &titleRect)
}
