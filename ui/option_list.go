package ui

import (
	"time"

	"github.com/UncleJunVIP/gabagool/internal"
	"github.com/patrickhuber/go-types"
	"github.com/patrickhuber/go-types/option"
	"github.com/veandco/go-sdl2/sdl"
)

type OptionType int

const (
	OptionTypeStandard OptionType = iota
	OptionTypeKeyboard
	OptionTypeClickable
)

// Option represents a single option for a menu item.
// DisplayName is the text that will be displayed to the user.
// Value is the value that will be returned when the option is submitted.
// Type controls the option's behavior. There are three types:
//   - Standard: A standard option that will be displayed to the user.
//   - Keyboard: A keyboard option that will be displayed to the user.
//   - Clickable: A clickable option that will be displayed to the user.
//
// KeyboardPrompt is the text that will be displayed to the user when the option is a keyboard option.
type Option struct {
	DisplayName    string
	Value          interface{}
	Type           OptionType
	KeyboardPrompt string
}

// ItemWithOptions represents a menu item with multiple choices.
// Item is the menu item itself.
// Options is the list of options for the menu item.
// SelectedOption is the index of the currently selected option.
type ItemWithOptions struct {
	Item           MenuItem
	Options        []Option
	SelectedOption int
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
	Margins         Padding
	ItemSpacing     int32
	InputDelay      time.Duration
	Title           string
	TitleAlign      internal.TextAlignment
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

	itemScrollData map[int]*textScrollData
}

func defaultOptionsListSettings(title string) optionsListSettings {
	return optionsListSettings{
		Margins:         UniformPadding(20),
		ItemSpacing:     60,
		InputDelay:      internal.DefaultInputDelay,
		Title:           title,
		TitleAlign:      internal.AlignLeft,
		TitleSpacing:    internal.DefaultTitleSpacing,
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

	return &optionsListController{
		Items:          items,
		SelectedIndex:  selectedIndex,
		Settings:       defaultOptionsListSettings(title),
		StartY:         20,
		lastInputTime:  time.Now(),
		itemScrollData: make(map[int]*textScrollData),
	}
}

// OptionsList presents a list of options to the user.
// This blocks until a selection is made or the user cancels.
func OptionsList(title string, items []ItemWithOptions, footerHelpItems []FooterHelpItem) (types.Option[OptionsListReturn], error) {
	window := internal.GetWindow()
	renderer := window.Renderer

	optionsListController := newOptionsListController(title, items)

	optionsListController.MaxVisibleItems = 8
	optionsListController.Settings.FooterHelpItems = footerHelpItems

	running := true
	result := OptionsListReturn{
		Items: items,
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

				switch e.Keysym.Sym {
				case sdl.K_b:
					running = false
					result.SelectedIndex = -1
					result.Canceled = true

				case sdl.K_a:

					if optionsListController.SelectedIndex >= 0 && optionsListController.SelectedIndex < len(optionsListController.Items) {
						item := &optionsListController.Items[optionsListController.SelectedIndex]
						if len(item.Options) > 0 && item.SelectedOption < len(item.Options) {
							option := item.Options[item.SelectedOption]
							if option.Type == OptionTypeKeyboard {
								prompt := option.KeyboardPrompt
								if prompt == "" {
									prompt = "Enter value"
								}
								keyboardResult, keyboardErr := Keyboard(prompt)
								if keyboardErr == nil && keyboardResult.IsSome() {
									enteredText := keyboardResult.Unwrap()
									item.Options[item.SelectedOption] = Option{
										DisplayName:    enteredText,
										Value:          enteredText,
										Type:           OptionTypeKeyboard,
										KeyboardPrompt: option.KeyboardPrompt,
									}
								}
								continue
							} else if option.Type == OptionTypeClickable {
								running = false
								result.SelectedIndex = optionsListController.SelectedIndex
								result.SelectedItem = &optionsListController.Items[optionsListController.SelectedIndex]
								result.Canceled = false
							}
						}
					}

				case sdl.K_RETURN:
					running = false
					result.SelectedIndex = optionsListController.SelectedIndex
					result.SelectedItem = &optionsListController.Items[optionsListController.SelectedIndex]
					result.Canceled = false

				case sdl.K_LEFT:
					optionsListController.cycleOptionLeft()

				case sdl.K_RIGHT:
					optionsListController.cycleOptionRight()

				default:
					optionsListController.handleEvent(event)
				}

			case *sdl.ControllerButtonEvent:
				if e.Type != sdl.CONTROLLERBUTTONDOWN {
					continue
				}

				switch e.Button {
				case BrickButton_B:
					result.SelectedIndex = -1
					result.Canceled = true
					running = false

				case BrickButton_A:
					if optionsListController.SelectedIndex >= 0 && optionsListController.SelectedIndex < len(optionsListController.Items) {
						item := &optionsListController.Items[optionsListController.SelectedIndex]
						if len(item.Options) > 0 && item.SelectedOption < len(item.Options) {
							option := item.Options[item.SelectedOption]
							if option.Type == OptionTypeKeyboard {
								prompt := option.KeyboardPrompt
								if prompt == "" {
									prompt = "Enter value"
								}
								keyboardResult, keyboardErr := Keyboard(prompt)
								if keyboardErr == nil && keyboardResult.IsSome() {
									enteredText := keyboardResult.Unwrap()
									item.Options[item.SelectedOption] = Option{
										DisplayName:    enteredText,
										Value:          enteredText,
										Type:           OptionTypeKeyboard,
										KeyboardPrompt: option.KeyboardPrompt,
									}
								}
								continue
							} else if option.Type == OptionTypeClickable {
								running = false
								result.SelectedIndex = optionsListController.SelectedIndex
								result.SelectedItem = &optionsListController.Items[optionsListController.SelectedIndex]
								result.Canceled = false
							}
						}
					}

				case BrickButton_START:
					running = false
					result.SelectedIndex = optionsListController.SelectedIndex
					result.SelectedItem = &optionsListController.Items[optionsListController.SelectedIndex]
					result.Canceled = false

				case BrickButton_LEFT:
					optionsListController.cycleOptionLeft()

				case BrickButton_RIGHT:
					optionsListController.cycleOptionRight()

				default:
					optionsListController.handleEvent(event)
				}
			}
		}

		window.RenderBackground()

		renderer.SetDrawColor(0, 0, 0, 255)

		optionsListController.render(renderer)

		renderer.Present()

		sdl.Delay(16)
	}

	if err != nil || result.Canceled {
		return option.None[OptionsListReturn](), err
	}

	return option.Some(result), nil
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

func (olc *optionsListController) handleEvent(event sdl.Event) bool {
	currentTime := time.Now()
	if currentTime.Sub(olc.lastInputTime) < olc.Settings.InputDelay {
		return false
	}

	switch t := event.(type) {
	case *sdl.KeyboardEvent:
		if t.Type == sdl.KEYDOWN {
			return olc.handleKeyDown(t.Keysym.Sym)
		}
	case *sdl.ControllerButtonEvent:
		if t.Type == sdl.CONTROLLERBUTTONDOWN {
			return olc.handleButtonPress(t.Button)
		}
	}
	return false
}

func (olc *optionsListController) handleKeyDown(key sdl.Keycode) bool {
	olc.lastInputTime = time.Now()

	if key == sdl.K_h {
		olc.toggleHelp()
		return true
	}

	if olc.ShowingHelp {
		return olc.handleHelpScreenInput(key)
	}

	if olc.SelectedIndex >= 0 && olc.SelectedIndex < len(olc.Items) {
		item := &olc.Items[olc.SelectedIndex]
		if len(item.Options) > 0 && item.SelectedOption < len(item.Options) {
			option := item.Options[item.SelectedOption]
			if option.Type == OptionTypeKeyboard {
				if key == sdl.K_a {
					prompt := option.KeyboardPrompt
					if prompt == "" {
						prompt = "Enter value"
					}
					keyboardResult, err := Keyboard(prompt)
					if err == nil && keyboardResult.IsSome() {
						enteredText := keyboardResult.Unwrap()
						item.Options[item.SelectedOption] = Option{
							DisplayName:    enteredText,
							Value:          enteredText,
							Type:           OptionTypeKeyboard,
							KeyboardPrompt: option.KeyboardPrompt,
						}
					}
					return true
				}
			}
		}
	}

	return olc.handleNormalModeInput(key)
}

func (olc *optionsListController) handleHelpScreenInput(key sdl.Keycode) bool {
	switch key {
	case sdl.K_UP:
		olc.scrollHelpOverlay(-1)
		return true
	case sdl.K_DOWN:
		olc.scrollHelpOverlay(1)
		return true
	default:
		olc.ShowingHelp = false
		return true
	}
}

func (olc *optionsListController) handleNormalModeInput(key sdl.Keycode) bool {
	switch key {
	case sdl.K_UP:
		olc.Items[olc.SelectedIndex].Item.Selected = false
		if olc.SelectedIndex > 0 {
			olc.SelectedIndex--
		} else {
			olc.SelectedIndex = len(olc.Items) - 1
		}
		olc.Items[olc.SelectedIndex].Item.Selected = true
		olc.scrollTo(olc.SelectedIndex)
		if olc.OnSelect != nil {
			olc.OnSelect(olc.SelectedIndex, &olc.Items[olc.SelectedIndex])
		}
		return true

	case sdl.K_DOWN:
		olc.Items[olc.SelectedIndex].Item.Selected = false
		if olc.SelectedIndex < len(olc.Items)-1 {
			olc.SelectedIndex++
		} else {
			olc.SelectedIndex = 0
		}
		olc.Items[olc.SelectedIndex].Item.Selected = true
		olc.scrollTo(olc.SelectedIndex)
		if olc.OnSelect != nil {
			olc.OnSelect(olc.SelectedIndex, &olc.Items[olc.SelectedIndex])
		}
		return true

	default:
		return false
	}
}

func (olc *optionsListController) handleButtonPress(button uint8) bool {
	olc.lastInputTime = time.Now()

	if button == BrickButton_MENU {
		olc.toggleHelp()
		return true
	}

	if olc.ShowingHelp {
		return olc.handleHelpScreenButton(button)
	}

	return olc.handleNormalModeButton(button)
}

func (olc *optionsListController) handleHelpScreenButton(button uint8) bool {
	switch button {
	case BrickButton_UP:
		olc.scrollHelpOverlay(-1)
		return true
	case BrickButton_DOWN:
		olc.scrollHelpOverlay(1)
		return true
	default:
		olc.ShowingHelp = false
		return true
	}
}

func (olc *optionsListController) handleNormalModeButton(button uint8) bool {
	switch button {
	case BrickButton_UP:
		olc.Items[olc.SelectedIndex].Item.Selected = false
		if olc.SelectedIndex > 0 {
			olc.SelectedIndex--
		} else {
			olc.SelectedIndex = len(olc.Items) - 1
		}
		olc.Items[olc.SelectedIndex].Item.Selected = true
		olc.scrollTo(olc.SelectedIndex)
		if olc.OnSelect != nil {
			olc.OnSelect(olc.SelectedIndex, &olc.Items[olc.SelectedIndex])
		}
		return true

	case BrickButton_DOWN:
		olc.Items[olc.SelectedIndex].Item.Selected = false
		if olc.SelectedIndex < len(olc.Items)-1 {
			olc.SelectedIndex++
		} else {
			olc.SelectedIndex = 0
		}
		olc.Items[olc.SelectedIndex].Item.Selected = true
		olc.scrollTo(olc.SelectedIndex)
		if olc.OnSelect != nil {
			olc.OnSelect(olc.SelectedIndex, &olc.Items[olc.SelectedIndex])
		}
		return true

	case BrickButton_A:
		if olc.SelectedIndex >= 0 && olc.SelectedIndex < len(olc.Items) {
			item := &olc.Items[olc.SelectedIndex]
			if len(item.Options) > 0 && item.SelectedOption < len(item.Options) {
				option := item.Options[item.SelectedOption]
				if option.Type == OptionTypeKeyboard {
					prompt := option.KeyboardPrompt
					if prompt == "" {
						prompt = "Enter value"
					}
					keyboardResult, err := Keyboard(prompt)
					if err == nil && keyboardResult.IsSome() {
						enteredText := keyboardResult.Unwrap()
						item.Options[item.SelectedOption] = Option{
							DisplayName:    enteredText,
							Value:          enteredText,
							Type:           OptionTypeKeyboard,
							KeyboardPrompt: option.KeyboardPrompt,
						}
					}
					return true
				}
			}
		}

	default:
		return false
	}
	return false
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
		olc.helpOverlay = newHelpOverlay(helpLines)
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
		olc.helpOverlay.render(renderer, internal.GetSmallFont())
		return
	}

	window := internal.GetWindow()
	titleFont := internal.GetLargeFont()
	font := internal.GetSmallFont()

	if olc.Settings.Title != "" {
		titleSurface, _ := titleFont.RenderUTF8Blended(olc.Settings.Title, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		if titleSurface != nil {
			defer titleSurface.Free()
			titleTexture, _ := renderer.CreateTextureFromSurface(titleSurface)
			if titleTexture != nil {
				defer titleTexture.Destroy()

				var titleX int32
				switch olc.Settings.TitleAlign {
				case internal.AlignLeft:
					titleX = olc.Settings.Margins.Left
				case internal.AlignCenter:
					titleX = (window.Width - titleSurface.W) / 2
				case internal.AlignRight:
					titleX = window.Width - olc.Settings.Margins.Right - titleSurface.W
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

	visibleCount := min(olc.MaxVisibleItems, len(olc.Items)-olc.VisibleStartIndex)

	for i := 0; i < visibleCount; i++ {
		itemIndex := i + olc.VisibleStartIndex
		item := olc.Items[itemIndex]

		textColor := internal.GetTheme().ListTextColor
		bgColor := sdl.Color{R: 0, G: 0, B: 0, A: 0}

		if item.Item.Selected {
			textColor = internal.GetTheme().ListTextSelectedColor
			bgColor = internal.GetTheme().MainColor
		}

		itemY := olc.StartY + (int32(i) * olc.Settings.ItemSpacing)

		if item.Item.Selected {
			renderer.SetDrawColor(bgColor.R, bgColor.G, bgColor.B, bgColor.A)
			selectionRect := &sdl.Rect{
				X: olc.Settings.Margins.Left - 10,
				Y: itemY - 5,
				W: window.Width - olc.Settings.Margins.Left - olc.Settings.Margins.Right + 20,
				H: olc.Settings.ItemSpacing,
			}
			drawRoundedRect(renderer, selectionRect, 20)
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
				indicatorText := selectedOption.DisplayName
				if indicatorText == "" {
					indicatorText = "[Press A to input]"
				}

				optionSurface, _ := font.RenderUTF8Blended(indicatorText, textColor)
				if optionSurface != nil {
					defer optionSurface.Free()
					optionTexture, _ := renderer.CreateTextureFromSurface(optionSurface)
					if optionTexture != nil {
						defer optionTexture.Destroy()

						renderer.Copy(optionTexture, nil, &sdl.Rect{
							X: window.Width - olc.Settings.Margins.Right - optionSurface.W,
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
							X: window.Width - olc.Settings.Margins.Right - optionSurface.W,
							Y: itemY,
							W: optionSurface.W,
							H: optionSurface.H,
						})
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
							X: window.Width - olc.Settings.Margins.Right - optionSurface.W,
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
		internal.GetSmallFont(),
		olc.Settings.FooterHelpItems,
		olc.Settings.Margins.Bottom,
	)
}
