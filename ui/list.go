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

const (
	scrollDirectionRight = 1
	scrollDirectionLeft  = -1
)

type ListOptions struct {
	FooterHelpItems   []FooterHelpItem
	EnableAction      bool
	EnableMultiSelect bool
	EnableReordering  bool
	SelectedIndex     int
}

type textScrollData struct {
	needsScrolling bool
	scrollOffset   int32
	textWidth      int32
	containerWidth int32
	direction      int
	lastUpdateTime time.Time
	pauseCounter   int
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
	Items         []models.MenuItem
	SelectedIndex int
	SelectedItems map[int]bool
	MultiSelect   bool
	ReorderMode   bool
	Settings      listSettings
	StartY        int32
	lastInputTime time.Time
	OnSelect      func(index int, item *models.MenuItem)

	VisibleStartIndex int
	MaxVisibleItems   int
	OnReorder         func(from, to int)

	EnableAction bool

	HelpEnabled bool
	helpOverlay *helpOverlay
	ShowingHelp bool

	itemScrollData map[int]*textScrollData
}

var defaultListHelpLines = []string{
	"Navigation Controls:",
	"• Up / Down: Navigate through items",
	"• A: Select current item",
	"• B: Cancel and exit",
}

func defaultListSettings(title string) listSettings {
	return listSettings{
		Margins:         models.UniformPadding(20),
		ItemSpacing:     internal.DefaultMenuSpacing,
		InputDelay:      internal.DefaultInputDelay,
		Title:           title,
		TitleAlign:      internal.AlignLeft,
		TitleSpacing:    internal.DefaultTitleSpacing,
		ScrollSpeed:     150.0,
		ScrollPauseTime: 25,
		FooterText:      "",
		FooterTextColor: sdl.Color{R: 180, G: 180, B: 180, A: 255},
		FooterHelpItems: []FooterHelpItem{},
	}
}

func newListController(title string, items []models.MenuItem) *listController {
	selectedItems := make(map[int]bool)
	selectedIndex := 0

	for i, item := range items {
		if item.Selected {
			selectedIndex = i
			break
		}
	}

	for i := range items {
		items[i].Selected = i == selectedIndex
		if items[i].Selected {
			selectedItems[i] = true
		}
	}

	return &listController{
		Items:          items,
		SelectedIndex:  selectedIndex,
		SelectedItems:  selectedItems,
		MultiSelect:    false,
		Settings:       defaultListSettings(title),
		StartY:         20,
		lastInputTime:  time.Now(),
		itemScrollData: make(map[int]*textScrollData),
	}
}

func List(title string, items []models.MenuItem, listOptions ListOptions) (types.Option[models.ListReturn], error) {
	window := internal.GetWindow()
	renderer := window.Renderer

	listController := newListController(title, items)

	listController.SelectedIndex = listOptions.SelectedIndex
	listController.MaxVisibleItems = 8
	listController.EnableAction = listOptions.EnableAction
	listController.Settings.FooterHelpItems = listOptions.FooterHelpItems

	if listOptions.EnableMultiSelect {
		listController.Settings.MultiSelectKey = sdl.K_SPACE
		listController.Settings.MultiSelectButton = BrickButton_SELECT
	}

	if listOptions.EnableReordering {
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
					if indices := listController.getSelectedItems(); len(indices) > 0 {
						result.PopulateMultiSelection(indices, items)
						result.Cancelled = false
					}

				case e.Keysym.Sym == sdl.K_a && listController.MultiSelect:
					listController.handleEvent(event)

				case e.Keysym.Sym == sdl.K_a && !listController.MultiSelect:
					running = false
					result.PopulateSingleSelection(listController.SelectedIndex, items)
					result.Cancelled = false

				case e.Keysym.Sym == sdl.K_b:
					running = false
					result.SelectedIndex = -1
					result.Cancelled = true

				default:
					listController.handleEvent(event)
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
					if indices := listController.getSelectedItems(); len(indices) > 0 {
						result.PopulateMultiSelection(indices, items)
						result.Cancelled = false
					}

				case e.Button == BrickButton_A && listController.MultiSelect:
					listController.handleEvent(event)

				case e.Button == BrickButton_A && !listController.MultiSelect:
					running = false
					result.PopulateSingleSelection(listController.SelectedIndex, items)
					result.Cancelled = false

				case e.Button == BrickButton_B:
					result.SelectedIndex = -1
					running = false

				default:
					listController.handleEvent(event)
				}
			}
		}

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		listController.render(renderer)

		renderer.Present()

		sdl.Delay(16)
	}

	if err != nil || result.Cancelled {
		return option.None[models.ListReturn](), err
	}

	return option.Some(result), nil
}

func (lc *listController) toggleMultiSelect() {
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

func (lc *listController) handleEvent(event sdl.Event) bool {
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

func (lc *listController) handleKeyDown(key sdl.Keycode) bool {
	lc.lastInputTime = time.Now()

	if key == sdl.K_h {
		lc.toggleHelp()
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

func (lc *listController) handleHelpScreenInput(key sdl.Keycode) bool {
	switch key {
	case sdl.K_UP:
		lc.scrollHelpOverlay(-1)
		return true
	case sdl.K_DOWN:
		lc.scrollHelpOverlay(1)
		return true
	default:
		lc.ShowingHelp = false
		return true
	}
}

func (lc *listController) handleReorderModeInput(key sdl.Keycode) bool {
	switch key {
	case sdl.K_UP:
		return lc.moveItemUp()
	case sdl.K_DOWN:
		return lc.moveItemDown()
	case sdl.K_ESCAPE, sdl.K_RETURN:
		lc.ReorderMode = false
		return true
	default:
		return false
	}
}

func (lc *listController) handleNormalModeInput(key sdl.Keycode) bool {
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
		lc.toggleMultiSelect()
		return true
	case sdl.K_a:
		if lc.MultiSelect {
			lc.toggleSelection(lc.SelectedIndex)
		}
		if lc.OnSelect != nil {
			lc.OnSelect(lc.SelectedIndex, &lc.Items[lc.SelectedIndex])
		}
		return true
	case sdl.K_z:
		if lc.MultiSelect {
			lc.toggleSelection(lc.SelectedIndex)
			if lc.OnSelect != nil {
				lc.OnSelect(lc.SelectedIndex, &lc.Items[lc.SelectedIndex])
			}
			return true
		}
	case lc.Settings.ReorderKey:
		lc.toggleReorderMode()
		return true
	}
	return false
}

func (lc *listController) handleButtonPress(button uint8) bool {
	lc.lastInputTime = time.Now()

	if button == BrickButton_MENU {
		lc.toggleHelp()
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

func (lc *listController) handleHelpScreenButtonInput(button uint8) bool {
	switch button {
	case BrickButton_UP:
		lc.scrollHelpOverlay(-1)
		return true
	case BrickButton_DOWN:
		lc.scrollHelpOverlay(1)
		return true
	default:
		return true
	}
}

func (lc *listController) handleReorderModeButtonInput(button uint8) bool {
	switch button {
	case BrickButton_UP:
		return lc.moveItemUp()
	case BrickButton_DOWN:
		return lc.moveItemDown()
	case BrickButton_B, BrickButton_A:
		lc.ReorderMode = false
		return true
	default:
		return false
	}
}

func (lc *listController) handleNormalModeButtonInput(button uint8) bool {
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
			lc.toggleSelection(lc.SelectedIndex)
		}
		if lc.OnSelect != nil {
			lc.OnSelect(lc.SelectedIndex, &lc.Items[lc.SelectedIndex])
		}
		return true
	case lc.Settings.MultiSelectButton:
		lc.toggleMultiSelect()
		return true
	case lc.Settings.ReorderButton:
		lc.toggleReorderMode()
		return true
	default:
		return false
	}
}

func (lc *listController) moveSelection(direction int) {
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

	drawScrollableMenu(renderer, internal.GetFont(), visibleItems, lc.StartY, lc.Settings, lc.MultiSelect, lc)

	lc.Settings.Title = originalTitle
	lc.Settings.TitleAlign = originalAlign

	if lc.ShowingHelp && lc.helpOverlay != nil {
		lc.helpOverlay.render(renderer, internal.GetSmallFont())
	}
}

func (lc *listController) toggleHelp() {
	if !lc.HelpEnabled {
		return
	}

	if lc.helpOverlay == nil {
		lc.helpOverlay = newHelpOverlay(defaultListHelpLines)
	}

	lc.helpOverlay.toggle()
	lc.ShowingHelp = lc.helpOverlay.ShowingHelp
}

func (lc *listController) scrollHelpOverlay(direction int) {
	if lc.helpOverlay != nil {
		lc.helpOverlay.scroll(direction)
	}
}

func (lc *listController) measureTextForScrolling(idx int, item models.MenuItem, maxWidth int32) *textScrollData {

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

func (lc *listController) updateItemScrollAnimation(data *textScrollData) {

	currentTime := time.Now()
	elapsed := currentTime.Sub(data.lastUpdateTime).Seconds()
	data.lastUpdateTime = currentTime

	if data.pauseCounter > 0 {
		data.pauseCounter--
		return
	}

	pixelsToScroll := max(int32(float32(elapsed)*lc.Settings.ScrollSpeed), 1)

	if data.direction == scrollDirectionRight {

		data.scrollOffset += pixelsToScroll

		if data.scrollOffset >= data.textWidth-data.containerWidth {
			data.scrollOffset = data.textWidth - data.containerWidth
			data.direction = scrollDirectionLeft
			data.pauseCounter = lc.Settings.ScrollPauseTime
		}
	} else {

		data.scrollOffset -= pixelsToScroll

		if data.scrollOffset <= 0 {
			data.scrollOffset = 0
			data.direction = scrollDirectionRight
			data.pauseCounter = lc.Settings.ScrollPauseTime
		}
	}
}

func drawScrollableMenu(renderer *sdl.Renderer, font *ttf.Font, visibleItems []models.MenuItem,
	startY int32, settings listSettings, multiSelect bool, controller *listController) {

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
		screenWidth = 768
	}

	maxTextWidth := screenWidth - settings.Margins.Left - settings.Margins.Right - 15

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

		pillWidth := textWidth + 10
		if needsScrolling {
			pillWidth = maxTextWidth + 10
		}

		if item.Selected || item.Focused {
			pillRect := sdl.Rect{
				X: settings.Margins.Left,
				Y: itemY,
				W: pillWidth,
				H: pillHeight,
			}
			drawRoundedRect(renderer, &pillRect, 12, bgColor)
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

	renderListFooter(renderer, settings, controller.MultiSelect)
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

func renderListFooter(renderer *sdl.Renderer, settings listSettings, isMultiSelect bool) {

	if settings.FooterText == "" && len(settings.FooterHelpItems) == 0 && !isMultiSelect {
		return
	}

	_, screenHeight, err := renderer.GetOutputSize()
	if err != nil {
		return
	}

	if isMultiSelect {
		pillHeight := int32(40)
		pillPadding := int32(12)
		pillSpacing := int32(20)
		font := internal.GetSmallFont()

		whitePillRadius := int32(8)
		blackPillRadius := int32(6)

		buttonPaddingX := int32(12)

		multiSelectHelpItems := []struct {
			ButtonName string
			HelpText   string
		}{
			{ButtonName: "A", HelpText: "Add / Remove"},
			{ButtonName: "Select", HelpText: "Cancel"},
			{ButtonName: "Start", HelpText: "Confirm"},
		}

		var totalWidth int32 = 0
		var pillInfos []struct {
			buttonName     string
			helpText       string
			buttonWidth    int32
			buttonHeight   int32
			helpWidth      int32
			blackPillWidth int32
			whitePillWidth int32
		}

		for _, item := range multiSelectHelpItems {

			buttonSurface, err := font.RenderUTF8Blended(item.ButtonName, sdl.Color{R: 255, G: 255, B: 255, A: 255})
			if err != nil {
				continue
			}
			buttonWidth := buttonSurface.W
			buttonHeight := buttonSurface.H
			buttonSurface.Free()

			blackPillWidth := buttonWidth + buttonPaddingX*2

			helpSurface, err := font.RenderUTF8Blended(item.HelpText, sdl.Color{R: 50, G: 50, B: 50, A: 255})
			if err != nil {
				continue
			}
			helpWidth := helpSurface.W
			helpSurface.Free()

			whitePillWidth := blackPillWidth + helpWidth + pillPadding*3

			pillInfos = append(pillInfos, struct {
				buttonName     string
				helpText       string
				buttonWidth    int32
				buttonHeight   int32
				helpWidth      int32
				blackPillWidth int32
				whitePillWidth int32
			}{
				buttonName:     item.ButtonName,
				helpText:       item.HelpText,
				buttonWidth:    buttonWidth,
				buttonHeight:   buttonHeight,
				helpWidth:      helpWidth,
				blackPillWidth: blackPillWidth,
				whitePillWidth: whitePillWidth,
			})

			totalWidth += whitePillWidth + pillSpacing
		}

		screenWidth, _, err := renderer.GetOutputSize()
		if err != nil {
			screenWidth = 768
		}

		startX := (screenWidth - totalWidth + pillSpacing) / 2
		currentX := startX

		for _, info := range pillInfos {

			outerPillRect := &sdl.Rect{
				X: currentX,
				Y: screenHeight - pillHeight - settings.Margins.Bottom,
				W: info.whitePillWidth,
				H: pillHeight,
			}

			whiteColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
			drawRoundedRect(renderer, outerPillRect, whitePillRadius, whiteColor)

			blackPillRect := &sdl.Rect{
				X: currentX + pillPadding,
				Y: screenHeight - pillHeight - settings.Margins.Bottom + pillPadding/2,
				W: info.blackPillWidth,
				H: pillHeight - pillPadding,
			}
			blackColor := sdl.Color{R: 0, G: 0, B: 0, A: 255}
			drawRoundedRect(renderer, blackPillRect, blackPillRadius, blackColor)

			blackPillCenterX := blackPillRect.X + blackPillRect.W/2
			blackPillCenterY := blackPillRect.Y + blackPillRect.H/2

			buttonSurface, err := font.RenderUTF8Blended(info.buttonName, sdl.Color{R: 255, G: 255, B: 255, A: 255})
			if err == nil {

				buttonWidth := buttonSurface.W
				buttonHeight := buttonSurface.H

				buttonTexture, err := renderer.CreateTextureFromSurface(buttonSurface)
				if err == nil {

					buttonRect := &sdl.Rect{
						X: blackPillCenterX - buttonWidth/2,
						Y: blackPillCenterY - buttonHeight/2,
						W: buttonWidth,
						H: buttonHeight,
					}
					renderer.Copy(buttonTexture, nil, buttonRect)
					buttonTexture.Destroy()
				}
				buttonSurface.Free()
			}

			helpSurface, err := font.RenderUTF8Blended(info.helpText, sdl.Color{R: 50, G: 50, B: 50, A: 255})
			if err == nil {
				helpTexture, err := renderer.CreateTextureFromSurface(helpSurface)
				if err == nil {
					helpRect := &sdl.Rect{
						X: currentX + info.blackPillWidth + pillPadding*2,
						Y: screenHeight - pillHeight - settings.Margins.Bottom + (pillHeight-helpSurface.H)/2,
						W: helpSurface.W,
						H: helpSurface.H,
					}
					renderer.Copy(helpTexture, nil, helpRect)
					helpTexture.Destroy()
				}
				helpSurface.Free()
			}

			currentX += info.whitePillWidth + pillSpacing
		}

		return
	}

	if settings.FooterText != "" {
		footerSurface, err := internal.GetSmallFont().RenderUTF8Blended(
			settings.FooterText,
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
		return
	}

	if len(settings.FooterHelpItems) > 0 {
		pillHeight := int32(40)
		pillPadding := int32(12)
		pillSpacing := int32(20)
		font := internal.GetSmallFont()

		whitePillRadius := int32(8)
		blackPillRadius := int32(6)

		buttonPaddingX := int32(12)

		var totalWidth int32 = 0
		var pillInfos []struct {
			buttonName     string
			helpText       string
			buttonWidth    int32
			buttonHeight   int32
			helpWidth      int32
			blackPillWidth int32
			whitePillWidth int32
		}

		for _, item := range settings.FooterHelpItems {

			buttonSurface, err := font.RenderUTF8Blended(item.ButtonName, sdl.Color{R: 255, G: 255, B: 255, A: 255})
			if err != nil {
				continue
			}
			buttonWidth := buttonSurface.W
			buttonHeight := buttonSurface.H
			buttonSurface.Free()

			blackPillWidth := buttonWidth + buttonPaddingX*2

			helpSurface, err := font.RenderUTF8Blended(item.HelpText, sdl.Color{R: 50, G: 50, B: 50, A: 255})
			if err != nil {
				continue
			}
			helpWidth := helpSurface.W
			helpSurface.Free()

			whitePillWidth := blackPillWidth + helpWidth + pillPadding*3

			pillInfos = append(pillInfos, struct {
				buttonName     string
				helpText       string
				buttonWidth    int32
				buttonHeight   int32
				helpWidth      int32
				blackPillWidth int32
				whitePillWidth int32
			}{
				buttonName:     item.ButtonName,
				helpText:       item.HelpText,
				buttonWidth:    buttonWidth,
				buttonHeight:   buttonHeight,
				helpWidth:      helpWidth,
				blackPillWidth: blackPillWidth,
				whitePillWidth: whitePillWidth,
			})

			totalWidth += whitePillWidth + pillSpacing
		}

		screenWidth, _, err := renderer.GetOutputSize()
		if err != nil {
			screenWidth = 768
		}

		startX := (screenWidth - totalWidth + pillSpacing) / 2
		currentX := startX

		for _, info := range pillInfos {

			outerPillRect := &sdl.Rect{
				X: currentX,
				Y: screenHeight - pillHeight - settings.Margins.Bottom,
				W: info.whitePillWidth,
				H: pillHeight,
			}

			whiteColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
			drawRoundedRect(renderer, outerPillRect, whitePillRadius, whiteColor)

			blackPillRect := &sdl.Rect{
				X: currentX + pillPadding,
				Y: screenHeight - pillHeight - settings.Margins.Bottom + pillPadding/2,
				W: info.blackPillWidth,
				H: pillHeight - pillPadding,
			}
			blackColor := sdl.Color{R: 0, G: 0, B: 0, A: 255}
			drawRoundedRect(renderer, blackPillRect, blackPillRadius, blackColor)

			blackPillCenterX := blackPillRect.X + blackPillRect.W/2
			blackPillCenterY := blackPillRect.Y + blackPillRect.H/2

			buttonSurface, err := font.RenderUTF8Blended(info.buttonName, sdl.Color{R: 255, G: 255, B: 255, A: 255})
			if err == nil {

				buttonWidth := buttonSurface.W
				buttonHeight := buttonSurface.H

				buttonTexture, err := renderer.CreateTextureFromSurface(buttonSurface)
				if err == nil {

					buttonRect := &sdl.Rect{
						X: blackPillCenterX - buttonWidth/2,
						Y: blackPillCenterY - buttonHeight/2,
						W: buttonWidth,
						H: buttonHeight,
					}
					renderer.Copy(buttonTexture, nil, buttonRect)
					buttonTexture.Destroy()
				}
				buttonSurface.Free()
			}

			helpSurface, err := font.RenderUTF8Blended(info.helpText, sdl.Color{R: 50, G: 50, B: 50, A: 255})
			if err == nil {
				helpTexture, err := renderer.CreateTextureFromSurface(helpSurface)
				if err == nil {
					helpRect := &sdl.Rect{
						X: currentX + info.blackPillWidth + pillPadding*2,
						Y: screenHeight - pillHeight - settings.Margins.Bottom + (pillHeight-helpSurface.H)/2,
						W: helpSurface.W,
						H: helpSurface.H,
					}
					renderer.Copy(helpTexture, nil, helpRect)
					helpTexture.Destroy()
				}
				helpSurface.Free()
			}

			currentX += info.whitePillWidth + pillSpacing
		}
	}
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

	maxTextWidth := screenWidth - lc.Settings.Margins.Left - lc.Settings.Margins.Right - 15

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
		direction:      1,
		scrollOffset:   0,
		lastUpdateTime: time.Now(),
		pauseCounter:   lc.Settings.ScrollPauseTime,
	}
}

func (lc *listController) updateScrollAnimation(data *textScrollData) {

	currentTime := time.Now()
	elapsed := currentTime.Sub(data.lastUpdateTime).Seconds()
	data.lastUpdateTime = currentTime

	if data.pauseCounter > 0 {
		data.pauseCounter--
		return
	}

	pixelsToScroll := max(int32(float32(elapsed)*lc.Settings.ScrollSpeed), 1)

	if data.direction > 0 {

		data.scrollOffset += pixelsToScroll

		if data.scrollOffset >= data.textWidth-data.containerWidth {
			data.scrollOffset = data.textWidth - data.containerWidth
			data.direction = -1
			data.pauseCounter = lc.Settings.ScrollPauseTime
		}
	} else {

		data.scrollOffset -= pixelsToScroll

		if data.scrollOffset <= 0 {
			data.scrollOffset = 0
			data.direction = 1
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
