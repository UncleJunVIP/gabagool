package gabagool

import (
	"fmt"
	"strings"
	"time"

	"github.com/UncleJunVIP/gabagool/pkg/gabagool/core"
	"github.com/patrickhuber/go-types"
	"github.com/patrickhuber/go-types/option"
	"github.com/veandco/go-sdl2/sdl"
)

type OptionType int

const (
	OptionTypeStandard OptionType = iota
	OptionTypeKeyboard
	OptionTypeClickable
	OptionTypeColorPicker // New option type for the color picker
)

// Option represents a single option for a menu item.
// DisplayName is the text that will be displayed to the user.
// Value is the value that will be returned when the option is submitted.
// Type controls the option's behavior. There are four types:
//   - Standard: A standard option that will be displayed to the user.
//   - Keyboard: A keyboard option that will be displayed to the user.
//   - Clickable: A clickable option that will be displayed to the user.
//   - ColorPicker: A hexagonal color picker for selecting colors.
//
// KeyboardPrompt is the text that will be displayed to the user when the option is a keyboard option.
// For ColorPicker type, Value should be an sdl.Color.
type Option struct {
	DisplayName    string
	Value          interface{}
	Type           OptionType
	KeyboardPrompt string
	Masked         bool
	OnUpdate       func(newValue interface{})
}

// ItemWithOptions represents a menu item with multiple choices.
// Item is the menu item itself.
// Options is the list of options for the menu item.
// SelectedOption is the index of the currently selected option.
type ItemWithOptions struct {
	Item           MenuItem
	Options        []Option
	SelectedOption int
	colorPicker    *ColorPicker // New field to store the color picker instance
}

func (iow *ItemWithOptions) Value() interface{} {
	if iow.Options[iow.SelectedOption].Value == nil {
		return ""
	}

	return fmt.Sprintf("%s", iow.Options[iow.SelectedOption].Value)
}

// OptionsListReturn represents the return value of the OptionsList function.
// Items is the entire list of menu items that were selected.
// SelectedIndex is the index of the selected item.
// SelectedItem is the selected item.
// Canceled is true if the user canceled the OptionsList.
type OptionsListReturn struct {
	Items         []ItemWithOptions
	SelectedIndex int
	SelectedItem  *ItemWithOptions
	Canceled      bool
}
type optionsListSettings struct {
	Margins         padding
	ItemSpacing     int32
	InputDelay      time.Duration
	Title           string
	TitleAlign      TextAlign
	TitleSpacing    int32
	ScrollSpeed     float32
	ScrollPauseTime int
	FooterHelpItems []FooterHelpItem
	FooterTextColor sdl.Color
}

type optionsListController struct {
	Items         []ItemWithOptions
	SelectedIndex int
	Settings      optionsListSettings
	StartY        int32
	lastInputTime time.Time
	OnSelect      func(index int, item *ItemWithOptions)

	VisibleStartIndex int
	MaxVisibleItems   int

	HelpEnabled bool
	helpOverlay *helpOverlay
	ShowingHelp bool

	itemScrollData       map[int]*textScrollData
	showingColorPicker   bool
	activeColorPickerIdx int
}

func defaultOptionsListSettings(title string) optionsListSettings {
	return optionsListSettings{
		Margins:         uniformPadding(20),
		ItemSpacing:     60,
		InputDelay:      DefaultInputDelay,
		Title:           title,
		TitleAlign:      TextAlignLeft,
		TitleSpacing:    DefaultTitleSpacing,
		ScrollSpeed:     150.0,
		ScrollPauseTime: 25,
		FooterTextColor: sdl.Color{R: 180, G: 180, B: 180, A: 255},
		FooterHelpItems: []FooterHelpItem{},
	}
}

func newOptionsListController(title string, items []ItemWithOptions) *optionsListController {
	selectedIndex := 0

	for i, item := range items {
		if item.Item.Selected {
			selectedIndex = i
			break
		}
	}

	for i := range items {
		items[i].Item.Selected = i == selectedIndex
	}

	// Initialize color pickers for any color picker options
	for i := range items {
		for j, opt := range items[i].Options {
			if opt.Type == OptionTypeColorPicker {
				// Initialize with default color if not already set
				if opt.Value == nil {
					items[i].Options[j].Value = sdl.Color{R: 255, G: 255, B: 255, A: 255}
				}

				// Create the color picker
				window := GetWindow()
				items[i].colorPicker = NewHexColorPicker(window)

				// Initialize with current color value if it's an sdl.Color
				if color, ok := opt.Value.(sdl.Color); ok {
					colorFound := false
					for idx, pickerColor := range items[i].colorPicker.Colors {
						if pickerColor.R == color.R && pickerColor.G == color.G && pickerColor.B == color.B {
							items[i].colorPicker.SelectedIndex = idx
							colorFound = true
							break
						}
					}
					// If color not found in the predefined list, we could add it
					if !colorFound {
						// Optional: Add custom color to the list or leave as is
					}
				}

				// Set visibility to false initially
				items[i].colorPicker.SetVisible(false)

				// Set the callback for when a color is selected
				items[i].colorPicker.SetOnColorSelected(func(color sdl.Color) {
					items[i].Options[j].Value = color
					items[i].Options[j].DisplayName = fmt.Sprintf("#%02X%02X%02X", color.R, color.G, color.B)

					if items[i].Options[j].OnUpdate != nil {
						items[i].Options[j].OnUpdate(color)
					}
				})

				break // Only need one color picker per item
			}
		}
	}

	return &optionsListController{
		Items:                items,
		SelectedIndex:        selectedIndex,
		Settings:             defaultOptionsListSettings(title),
		StartY:               20,
		lastInputTime:        time.Now(),
		itemScrollData:       make(map[int]*textScrollData),
		showingColorPicker:   false,
		activeColorPickerIdx: -1,
	}
}

// OptionsList presents a list of options to the user.
// This blocks until a selection is made or the user cancels.
func OptionsList(title string, items []ItemWithOptions, footerHelpItems []FooterHelpItem) (types.Option[OptionsListReturn], error) {
	window := GetWindow()
	renderer := window.Renderer
	processor := GetInputProcessor()

	optionsListController := newOptionsListController(title, items)

	// Calculate MaxVisibleItems based on window height
	optionsListController.MaxVisibleItems = int(optionsListController.calculateMaxVisibleItems(window))
	optionsListController.Settings.FooterHelpItems = footerHelpItems

	running := true
	result := OptionsListReturn{
		Items: items,
	}

	var err error

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
				err = sdl.GetError()

			case *sdl.KeyboardEvent, *sdl.ControllerButtonEvent, *sdl.ControllerAxisEvent, *sdl.JoyButtonEvent, *sdl.JoyAxisEvent, *sdl.JoyHatEvent:
				inputEvent := processor.ProcessSDLEvent(event.(sdl.Event))
				if inputEvent == nil {
					continue
				}

				if optionsListController.showingColorPicker {
					optionsListController.handleColorPickerInput(inputEvent)
				} else {
					optionsListController.handleOptionsInput(inputEvent, &running, &result)
				}
			}
		}

		if window.Background != nil {
			window.RenderBackground()
		} else {
			renderer.SetDrawColor(0, 0, 0, 255)
			renderer.Clear()
		}

		// If showing color picker, draw it; otherwise draw the options list
		if optionsListController.showingColorPicker &&
			optionsListController.activeColorPickerIdx >= 0 &&
			optionsListController.activeColorPickerIdx < len(optionsListController.Items) {
			item := &optionsListController.Items[optionsListController.activeColorPickerIdx]
			if item.colorPicker != nil {
				item.colorPicker.Draw(renderer)
			}
		} else {
			optionsListController.render(renderer)
		}

		renderer.Present()

		sdl.Delay(16)
	}

	if err != nil || result.Canceled {
		return option.None[OptionsListReturn](), err
	}

	return option.Some(result), nil
}

func (olc *optionsListController) calculateMaxVisibleItems(window *Window) int32 {
	scaleFactor := GetScaleFactor()

	itemSpacing := int32(float32(60) * scaleFactor)

	_, screenHeight, _ := window.Renderer.GetOutputSize()

	var titleHeight int32 = 0
	if olc.Settings.Title != "" {
		titleHeight = int32(float32(60) * scaleFactor)
		titleHeight += olc.Settings.TitleSpacing
	}

	footerHeight := int32(float32(50) * scaleFactor)

	availableHeight := screenHeight - titleHeight - footerHeight - olc.StartY

	maxItems := availableHeight / itemSpacing

	if maxItems < 1 {
		maxItems = 1
	}

	return maxItems
}

func (olc *optionsListController) handleColorPickerInput(inputEvent *InputEvent) {
	if !inputEvent.Pressed {
		return
	}

	if olc.activeColorPickerIdx < 0 || olc.activeColorPickerIdx >= len(olc.Items) {
		return
	}

	item := &olc.Items[olc.activeColorPickerIdx]
	if item.colorPicker == nil {
		return
	}

	switch inputEvent.Button {
	case InternalButtonB:
		olc.hideColorPicker()
	case InternalButtonA:
		selectedColor := item.colorPicker.GetSelectedColor()
		for j := range item.Options {
			if item.Options[j].Type == OptionTypeColorPicker {
				item.Options[j].Value = selectedColor
				item.Options[j].DisplayName = fmt.Sprintf("#%02X%02X%02X",
					selectedColor.R, selectedColor.G, selectedColor.B)
				if item.Options[j].OnUpdate != nil {
					item.Options[j].OnUpdate(selectedColor)
				}
				break
			}
		}
		olc.hideColorPicker()
	case InternalButtonLeft, InternalButtonRight, InternalButtonUp, InternalButtonDown:
		// Convert internal button to keycode for color picker
		var keycode sdl.Keycode
		switch inputEvent.Button {
		case InternalButtonLeft:
			keycode = sdl.K_LEFT
		case InternalButtonRight:
			keycode = sdl.K_RIGHT
		case InternalButtonUp:
			keycode = sdl.K_UP
		case InternalButtonDown:
			keycode = sdl.K_DOWN
		}
		item.colorPicker.handleKeyPress(keycode)

		selectedColor := item.colorPicker.GetSelectedColor()
		for j := range item.Options {
			if item.Options[j].Type == OptionTypeColorPicker && item.Options[j].OnUpdate != nil {
				item.Options[j].OnUpdate(selectedColor)
				break
			}
		}
	}
}

func (olc *optionsListController) handleOptionsInput(inputEvent *InputEvent, running *bool, result *OptionsListReturn) {
	if !inputEvent.Pressed {
		return
	}

	currentTime := time.Now()
	if currentTime.Sub(olc.lastInputTime) < olc.Settings.InputDelay {
		return
	}

	switch inputEvent.Button {
	case InternalButtonMenu:
		olc.toggleHelp()
		olc.lastInputTime = time.Now()

	case InternalButtonB:
		if olc.ShowingHelp {
			olc.ShowingHelp = false
		} else {
			*running = false
			result.SelectedIndex = -1
			result.Canceled = true
		}
		olc.lastInputTime = time.Now()

	case InternalButtonA:
		if olc.ShowingHelp {
			olc.ShowingHelp = false
		} else {
			olc.handleAButton(running, result)
		}
		olc.lastInputTime = time.Now()

	case InternalButtonStart:
		if !olc.ShowingHelp && olc.SelectedIndex >= 0 && olc.SelectedIndex < len(olc.Items) {
			*running = false
			result.SelectedIndex = olc.SelectedIndex
			result.SelectedItem = &olc.Items[olc.SelectedIndex]
			result.Canceled = false
		}
		olc.lastInputTime = time.Now()

	case InternalButtonLeft:
		if !olc.ShowingHelp {
			olc.cycleOptionLeft()
		}
		olc.lastInputTime = time.Now()

	case InternalButtonRight:
		if !olc.ShowingHelp {
			olc.cycleOptionRight()
		}
		olc.lastInputTime = time.Now()

	case InternalButtonUp:
		if olc.ShowingHelp {
			olc.scrollHelpOverlay(-1)
		} else {
			olc.moveSelection(-1)
		}
		olc.lastInputTime = time.Now()

	case InternalButtonDown:
		if olc.ShowingHelp {
			olc.scrollHelpOverlay(1)
		} else {
			olc.moveSelection(1)
		}
		olc.lastInputTime = time.Now()
	}
}

func (olc *optionsListController) handleAButton(running *bool, result *OptionsListReturn) {
	if olc.SelectedIndex >= 0 && olc.SelectedIndex < len(olc.Items) {
		item := &olc.Items[olc.SelectedIndex]
		if len(item.Options) > 0 && item.SelectedOption < len(item.Options) {
			option := item.Options[item.SelectedOption]
			switch option.Type {
			case OptionTypeKeyboard:
				prompt := option.KeyboardPrompt
				keyboardResult, err := Keyboard(prompt)
				if err == nil && keyboardResult.IsSome() {
					enteredText := keyboardResult.Unwrap()
					item.Options[item.SelectedOption] = Option{
						DisplayName:    enteredText,
						Value:          enteredText,
						Type:           OptionTypeKeyboard,
						KeyboardPrompt: enteredText,
						Masked:         option.Masked,
					}
				}
			case OptionTypeColorPicker:
				olc.showColorPicker(olc.SelectedIndex)
			case OptionTypeClickable:
				*running = false
				result.SelectedIndex = olc.SelectedIndex
				result.SelectedItem = &olc.Items[olc.SelectedIndex]
				result.Canceled = false
			}
		}
	}
}

func (olc *optionsListController) moveSelection(direction int) {
	if len(olc.Items) == 0 {
		return
	}

	olc.Items[olc.SelectedIndex].Item.Selected = false

	if direction > 0 {
		olc.SelectedIndex++
		if olc.SelectedIndex >= len(olc.Items) {
			olc.SelectedIndex = 0
			olc.VisibleStartIndex = 0
		}
	} else {
		olc.SelectedIndex--
		if olc.SelectedIndex < 0 {
			olc.SelectedIndex = len(olc.Items) - 1
			if len(olc.Items) > olc.MaxVisibleItems {
				olc.VisibleStartIndex = len(olc.Items) - olc.MaxVisibleItems
			} else {
				olc.VisibleStartIndex = 0
			}
		}
	}

	olc.Items[olc.SelectedIndex].Item.Selected = true
	olc.scrollTo(olc.SelectedIndex)

	if olc.OnSelect != nil {
		olc.OnSelect(olc.SelectedIndex, &olc.Items[olc.SelectedIndex])
	}
}

func (olc *optionsListController) showColorPicker(itemIndex int) {
	if itemIndex < 0 || itemIndex >= len(olc.Items) {
		return
	}

	item := &olc.Items[itemIndex]
	if item.colorPicker != nil {
		item.colorPicker.SetVisible(true)
		olc.showingColorPicker = true
		olc.activeColorPickerIdx = itemIndex
	}
}

func (olc *optionsListController) hideColorPicker() {
	if olc.activeColorPickerIdx >= 0 && olc.activeColorPickerIdx < len(olc.Items) {
		item := &olc.Items[olc.activeColorPickerIdx]
		if item.colorPicker != nil {
			item.colorPicker.SetVisible(false)
		}
	}
	olc.showingColorPicker = false
	olc.activeColorPickerIdx = -1
}

func (olc *optionsListController) cycleOptionLeft() {
	if olc.SelectedIndex < 0 || olc.SelectedIndex >= len(olc.Items) {
		return
	}

	item := &olc.Items[olc.SelectedIndex]
	if len(item.Options) == 0 {
		return
	}

	if item.Options[item.SelectedOption].Type == OptionTypeClickable {
		return
	}

	item.SelectedOption--
	if item.SelectedOption < 0 {
		item.SelectedOption = len(item.Options) - 1
	}

	currentOption := item.Options[item.SelectedOption]
	if currentOption.OnUpdate != nil {
		currentOption.OnUpdate(currentOption.Value)
	}
}

func (olc *optionsListController) cycleOptionRight() {
	if olc.SelectedIndex < 0 || olc.SelectedIndex >= len(olc.Items) {
		return
	}

	item := &olc.Items[olc.SelectedIndex]
	if len(item.Options) == 0 {
		return
	}

	if item.Options[item.SelectedOption].Type == OptionTypeClickable {
		return
	}

	item.SelectedOption++
	if item.SelectedOption >= len(item.Options) {
		item.SelectedOption = 0
	}

	currentOption := item.Options[item.SelectedOption]
	if currentOption.OnUpdate != nil {
		currentOption.OnUpdate(currentOption.Value)
	}
}

func (olc *optionsListController) scrollTo(index int) {
	if index < 0 || index >= len(olc.Items) {
		return
	}

	if index >= olc.VisibleStartIndex && index < olc.VisibleStartIndex+olc.MaxVisibleItems {
		return
	}

	if index < olc.VisibleStartIndex {
		olc.VisibleStartIndex = index
	} else {
		olc.VisibleStartIndex = index - olc.MaxVisibleItems + 1
		if olc.VisibleStartIndex < 0 {
			olc.VisibleStartIndex = 0
		}
	}
}

func (olc *optionsListController) toggleHelp() {
	if !olc.HelpEnabled {
		return
	}

	olc.ShowingHelp = !olc.ShowingHelp
	if olc.ShowingHelp && olc.helpOverlay == nil {
		helpLines := []string{
			"Navigation Controls:",
			"• Up / Down: Navigate through items",
			"• Left / Right: Change option for current item",
			"• A: Select or input text for keyboard options",
			"• B: Cancel and exit",
		}
		olc.helpOverlay = newHelpOverlay(fmt.Sprintf("%s Help", olc.Settings.Title), helpLines)
	}
}

func (olc *optionsListController) scrollHelpOverlay(direction int) {
	if olc.helpOverlay == nil {
		return
	}
	olc.helpOverlay.scroll(direction)
}

func (olc *optionsListController) render(renderer *sdl.Renderer) {
	if olc.ShowingHelp && olc.helpOverlay != nil {
		olc.helpOverlay.render(renderer, fonts.smallFont)
		return
	}

	scaleFactor := GetScaleFactor()
	window := GetWindow()
	titleFont := fonts.largeSymbolFont
	font := fonts.smallFont

	itemSpacing := int32(float32(60) * scaleFactor)
	selectionRectHeight := int32(float32(60) * scaleFactor)
	cornerRadius := int32(float32(20) * scaleFactor)

	if olc.Settings.Title != "" {
		titleSurface, _ := titleFont.RenderUTF8Blended(olc.Settings.Title, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		if titleSurface != nil {
			defer titleSurface.Free()
			titleTexture, _ := renderer.CreateTextureFromSurface(titleSurface)
			if titleTexture != nil {
				defer titleTexture.Destroy()

				var titleX int32
				switch olc.Settings.TitleAlign {
				case TextAlignLeft:
					titleX = olc.Settings.Margins.Left
				case TextAlignCenter:
					titleX = (window.GetWidth() - titleSurface.W) / 2
				case TextAlignRight:
					titleX = window.GetWidth() - olc.Settings.Margins.Right - titleSurface.W
				}

				renderer.Copy(titleTexture, nil, &sdl.Rect{
					X: titleX,
					Y: olc.Settings.Margins.Top,
					W: titleSurface.W,
					H: titleSurface.H,
				})

				olc.StartY = olc.Settings.Margins.Top + titleSurface.H + olc.Settings.TitleSpacing
			}
		}
	}

	olc.MaxVisibleItems = int(olc.calculateMaxVisibleItems(window))
	visibleCount := min(olc.MaxVisibleItems, len(olc.Items)-olc.VisibleStartIndex)

	for i := 0; i < visibleCount; i++ {
		itemIndex := i + olc.VisibleStartIndex
		item := olc.Items[itemIndex]

		textColor := core.GetTheme().ListTextColor
		bgColor := sdl.Color{R: 0, G: 0, B: 0, A: 0}

		if item.Item.Selected {
			textColor = core.GetTheme().ListTextSelectedColor
			bgColor = core.GetTheme().MainColor
		}

		itemY := olc.StartY + (int32(i) * itemSpacing)

		if item.Item.Selected {
			selectionRect := &sdl.Rect{
				X: olc.Settings.Margins.Left - 10,
				Y: itemY - 5,
				W: window.GetWidth() - olc.Settings.Margins.Left - olc.Settings.Margins.Right + 20,
				H: selectionRectHeight,
			}
			drawRoundedRect(renderer, selectionRect, cornerRadius, sdl.Color{R: bgColor.R, G: bgColor.G, B: bgColor.B, A: bgColor.A})
		}

		itemSurface, _ := font.RenderUTF8Blended(item.Item.Text, textColor)
		if itemSurface != nil {
			defer itemSurface.Free()
			itemTexture, _ := renderer.CreateTextureFromSurface(itemSurface)
			if itemTexture != nil {
				defer itemTexture.Destroy()
				renderer.Copy(itemTexture, nil, &sdl.Rect{
					X: olc.Settings.Margins.Left,
					Y: itemY,
					W: itemSurface.W,
					H: itemSurface.H,
				})
			}
		}

		if len(item.Options) > 0 {
			selectedOption := item.Options[item.SelectedOption]

			if selectedOption.Type == OptionTypeKeyboard {
				var indicatorText string

				if selectedOption.Masked {
					indicatorText = strings.Repeat("*", len(selectedOption.DisplayName))
				} else {
					indicatorText = selectedOption.DisplayName
				}

				optionSurface, _ := font.RenderUTF8Blended(indicatorText, textColor)
				if optionSurface != nil {
					defer optionSurface.Free()
					optionTexture, _ := renderer.CreateTextureFromSurface(optionSurface)
					if optionTexture != nil {
						defer optionTexture.Destroy()

						renderer.Copy(optionTexture, nil, &sdl.Rect{
							X: window.GetWidth() - olc.Settings.Margins.Right - optionSurface.W,
							Y: itemY,
							W: optionSurface.W,
							H: optionSurface.H,
						})
					}
				}
			} else if selectedOption.Type == OptionTypeClickable {
				indicatorText := selectedOption.DisplayName

				optionSurface, _ := font.RenderUTF8Blended(indicatorText, textColor)
				if optionSurface != nil {
					defer optionSurface.Free()
					optionTexture, _ := renderer.CreateTextureFromSurface(optionSurface)
					if optionTexture != nil {
						defer optionTexture.Destroy()

						renderer.Copy(optionTexture, nil, &sdl.Rect{
							X: window.GetWidth() - olc.Settings.Margins.Right - optionSurface.W,
							Y: itemY,
							W: optionSurface.W,
							H: optionSurface.H,
						})
					}
				}
			} else if selectedOption.Type == OptionTypeColorPicker {
				// For color picker option, display the color swatch and hex value
				indicatorText := selectedOption.DisplayName
				if indicatorText == "" {
					if color, ok := selectedOption.Value.(sdl.Color); ok {
						indicatorText = fmt.Sprintf("#%02X%02X%02X", color.R, color.G, color.B)
					} else {
						indicatorText = ""
					}
				}

				optionSurface, _ := font.RenderUTF8Blended(indicatorText, textColor)
				if optionSurface != nil {
					defer optionSurface.Free()
					optionTexture, _ := renderer.CreateTextureFromSurface(optionSurface)
					if optionTexture != nil {
						defer optionTexture.Destroy()

						// Make the swatch slightly smaller than text height
						swatchHeight := int32(float32(optionSurface.H) * 0.8) // 80% of text height
						swatchWidth := swatchHeight                           // Keep it square
						swatchSpacing := int32(float32(10) * scaleFactor)     // Scale spacing

						// Position swatch on the right
						swatchX := window.GetWidth() - olc.Settings.Margins.Right - swatchWidth

						// Position the text to the left of the swatch
						textX := swatchX - optionSurface.W - swatchSpacing

						// Center the swatch vertically with the text
						textCenterY := itemY + (optionSurface.H / 2)
						swatchY := textCenterY - (swatchHeight / 2)

						// Draw the text on the left
						renderer.Copy(optionTexture, nil, &sdl.Rect{
							X: textX,
							Y: itemY,
							W: optionSurface.W,
							H: optionSurface.H,
						})

						// Draw color swatch on the right
						if color, ok := selectedOption.Value.(sdl.Color); ok {
							swatchRect := &sdl.Rect{
								X: swatchX,
								Y: swatchY, // Centered vertically
								W: swatchWidth,
								H: swatchHeight,
							}

							// Save current color
							r, g, b, a, _ := renderer.GetDrawColor()

							// Draw color swatch
							renderer.SetDrawColor(color.R, color.G, color.B, color.A)
							renderer.FillRect(swatchRect)

							// Draw swatch border
							renderer.SetDrawColor(255, 255, 255, 255)
							renderer.DrawRect(swatchRect)

							// Restore previous color
							renderer.SetDrawColor(r, g, b, a)
						}
					}
				}
			} else {
				optionSurface, _ := font.RenderUTF8Blended(selectedOption.DisplayName, textColor)
				if optionSurface != nil {
					defer optionSurface.Free()
					optionTexture, _ := renderer.CreateTextureFromSurface(optionSurface)
					if optionTexture != nil {
						defer optionTexture.Destroy()

						renderer.Copy(optionTexture, nil, &sdl.Rect{
							X: window.GetWidth() - olc.Settings.Margins.Right - optionSurface.W,
							Y: itemY,
							W: optionSurface.W,
							H: optionSurface.H,
						})
					}
				}
			}
		}
	}

	renderFooter(
		renderer,
		fonts.smallFont,
		olc.Settings.FooterHelpItems,
		olc.Settings.Margins.Bottom,
		true,
	)
}
