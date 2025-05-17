package ui

import (
	"github.com/UncleJunVIP/gabagool/internal"
	"github.com/patrickhuber/go-types"
	"github.com/patrickhuber/go-types/option"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"time"
)

type key struct {
	Rect        sdl.Rect
	LowerValue  string
	UpperValue  string
	SymbolValue string
	IsPressed   bool
}

type keyboardState int

const (
	lowerCase keyboardState = iota
	upperCase
)

type virtualKeyboard struct {
	Keys             []key
	TextBuffer       string
	CurrentState     keyboardState
	ShiftPressed     bool
	BackspaceRect    sdl.Rect
	EnterRect        sdl.Rect
	SpaceRect        sdl.Rect
	ShiftRect        sdl.Rect
	TextInputRect    sdl.Rect
	KeyboardRect     sdl.Rect
	SelectedKeyIndex int
	SelectedSpecial  int

	CursorPosition  int
	CursorVisible   bool
	LastCursorBlink time.Time
	CursorBlinkRate time.Duration

	helpOverlay  *helpOverlay
	ShowingHelp  bool
	EnterPressed bool
}

var defaultKeyboardHelpLines = []string{
	"Keyboard Controls:",
	"• D-Pad: Navigate between keys",
	"• A: Type the selected key",
	"• B: backspace",
	"• X: Space",
	"• L1 / R1: Move cursor within text",
	"• Select: toggle Shift (uppercase/symbols)",
	"• Y: Exit keyboard without saving",
	"• Start: Enter (confirm input)",
}

func createKeyboard(windowWidth, windowHeight int32) *virtualKeyboard {
	kb := &virtualKeyboard{
		Keys:             make([]key, 0),
		TextBuffer:       "",
		CurrentState:     lowerCase,
		SelectedKeyIndex: 0,
		SelectedSpecial:  0,
		CursorPosition:   0,
		CursorVisible:    true,
		LastCursorBlink:  time.Now(),
		CursorBlinkRate:  500 * time.Millisecond,
		ShowingHelp:      false,
	}

	kb.helpOverlay = newHelpOverlay(defaultKeyboardHelpLines)

	keyboardWidth := (windowWidth * 85) / 100
	keyboardHeight := (windowHeight * 85) / 100

	textInputHeight := windowHeight / 10

	keyboardHeight = keyboardHeight - textInputHeight - 20

	startX := (windowWidth - keyboardWidth) / 2

	textInputY := (windowHeight - keyboardHeight - textInputHeight - 20) / 2
	keyboardStartY := textInputY + textInputHeight + 20

	kb.KeyboardRect = sdl.Rect{
		X: startX,
		Y: keyboardStartY,
		W: keyboardWidth,
		H: keyboardHeight,
	}

	keyWidth := keyboardWidth / 12
	keyHeight := keyboardHeight / 6
	keySpacing := int32(6)

	kb.TextInputRect = sdl.Rect{
		X: startX,
		Y: textInputY,
		W: keyboardWidth,
		H: textInputHeight,
	}

	rowKeys := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
	rowSymbols := []string{"!", "@", "#", "$", "%", "^", "&", "*", "(", ")"}

	rowWidth := (keyWidth * int32(len(rowKeys))) + (keySpacing * int32(len(rowKeys)-1))
	rowStartX := startX + (keyboardWidth-rowWidth-(keyWidth*2)-keySpacing)/2

	x := rowStartX
	y := keyboardStartY + keySpacing

	for i, keyVal := range rowKeys {
		kb.Keys = append(kb.Keys, key{
			Rect:        sdl.Rect{X: x, Y: y, W: keyWidth, H: keyHeight},
			LowerValue:  keyVal,
			UpperValue:  keyVal,
			SymbolValue: rowSymbols[i],
			IsPressed:   false,
		})
		x += keyWidth + keySpacing
	}

	kb.BackspaceRect = sdl.Rect{
		X: x,
		Y: y,
		W: keyWidth * 2,
		H: keyHeight,
	}

	rowKeys = []string{"q", "w", "e", "r", "t", "y", "u", "i", "o", "p"}

	rowWidth = (keyWidth * int32(len(rowKeys))) + (keySpacing * int32(len(rowKeys)-1))
	rowStartX = startX + (keyboardWidth-rowWidth)/2

	x = rowStartX
	y += keyHeight + keySpacing

	for _, keyVal := range rowKeys {
		kb.Keys = append(kb.Keys, key{
			Rect:       sdl.Rect{X: x, Y: y, W: keyWidth, H: keyHeight},
			LowerValue: keyVal,
			UpperValue: string([]rune(keyVal)[0] - 32),
			IsPressed:  false,
		})
		x += keyWidth + keySpacing
	}

	rowKeys = []string{"a", "s", "d", "f", "g", "h", "j", "k", "l"}

	rowWidth = (keyWidth * int32(len(rowKeys))) + (keySpacing * int32(len(rowKeys)-1))
	rowStartX = startX + (keyboardWidth-rowWidth)/2

	x = rowStartX
	y += keyHeight + keySpacing

	for _, keyVal := range rowKeys {
		kb.Keys = append(kb.Keys, key{
			Rect:       sdl.Rect{X: x, Y: y, W: keyWidth, H: keyHeight},
			LowerValue: keyVal,
			UpperValue: string([]rune(keyVal)[0] - 32),
			IsPressed:  false,
		})
		x += keyWidth + keySpacing
	}

	rowKeys = []string{"z", "x", "c", "v", "b", "n", "m"}

	shiftKeyWidth := keyWidth * 2

	regularKeysWidth := (keyWidth * int32(len(rowKeys))) + (keySpacing * int32(len(rowKeys)-1))

	enterKeyWidth := keyWidth + keyWidth/2

	totalFourthRowWidth := shiftKeyWidth + regularKeysWidth + enterKeyWidth + keySpacing*2

	fourthRowStartX := startX + (keyboardWidth-totalFourthRowWidth)/2

	kb.ShiftRect = sdl.Rect{
		X: fourthRowStartX,
		Y: y + keyHeight + keySpacing,
		W: shiftKeyWidth,
		H: keyHeight,
	}

	x = kb.ShiftRect.X + kb.ShiftRect.W + keySpacing

	for _, keyVal := range rowKeys {
		kb.Keys = append(kb.Keys, key{
			Rect:       sdl.Rect{X: x, Y: y + keyHeight + keySpacing, W: keyWidth, H: keyHeight},
			LowerValue: keyVal,
			UpperValue: string([]rune(keyVal)[0] - 32),
			IsPressed:  false,
		})
		x += keyWidth + keySpacing
	}

	kb.EnterRect = sdl.Rect{
		X: x,
		Y: y + keyHeight + keySpacing,
		W: enterKeyWidth,
		H: keyHeight,
	}

	spaceBarWidth := keyWidth * 6
	spaceBarX := startX + (keyboardWidth-spaceBarWidth)/2

	kb.SpaceRect = sdl.Rect{
		X: spaceBarX,
		Y: y + keyHeight*2 + keySpacing*2,
		W: spaceBarWidth,
		H: keyHeight,
	}

	return kb
}

// Keyboard creates a virtual keyboard that can be used to enter text.
// The keyboard is centered on screen and features both on-screen controls and button shortcuts.
// This returns an Option[string] that can be used to get the text entered by the user.
// If the user hits enter, the Option[string] will contain the text entered by the user.
// If the user exits, the Option[string] will be empty.
func Keyboard(initialText string) (types.Option[string], error) {
	window := internal.GetWindow()
	renderer := window.Renderer
	font := internal.GetMediumFont()

	kb := createKeyboard(window.Width, window.Height)

	if initialText != "" {
		kb.TextBuffer = initialText
		kb.CursorPosition = len(initialText)
	}

	running := true
	var result string
	var exitType int
	var err error

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
				err = sdl.GetError()

			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {
					kb.handleKeyDown(e.Keysym.Sym)

					if kb.EnterPressed {
						running = false
						result = kb.TextBuffer
						exitType = 0
						break
					}

					if e.Keysym.Sym == sdl.K_RETURN && !kb.ShowingHelp {
						running = false
						result = kb.TextBuffer
						exitType = 0
						break
					} else if e.Keysym.Sym == sdl.K_ESCAPE && !kb.ShowingHelp {
						running = false
						result = initialText
						exitType = 1
						break
					}
				}

			case *sdl.ControllerButtonEvent:
				if e.Type == sdl.CONTROLLERBUTTONDOWN {
					kb.handleButtonPress(e.Button)

					if kb.EnterPressed {
						running = false
						result = kb.TextBuffer
						exitType = 0
						break
					}

					if e.Button == BrickButton_START && !kb.ShowingHelp {
						running = false
						result = kb.TextBuffer
						exitType = 0
						break
					} else if e.Button == BrickButton_Y && !kb.ShowingHelp {
						running = false
						result = ""
						exitType = 1
						break
					}
				}

			}
		}

		kb.updateCursorBlink()

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		kb.render(renderer, font)

		renderer.Present()

		sdl.Delay(16)
	}

	if err != nil || exitType == 1 {
		return option.None[string](), err
	}

	return option.Some[string](result), nil
}

func (kb *virtualKeyboard) toggleHelp() {
	if kb.helpOverlay == nil {
		kb.helpOverlay = newHelpOverlay(defaultKeyboardHelpLines)
	}

	kb.helpOverlay.toggle()
	kb.ShowingHelp = kb.helpOverlay.ShowingHelp
}

func (kb *virtualKeyboard) scrollHelpOverlay(direction int) {
	if kb.helpOverlay != nil {
		kb.helpOverlay.scroll(direction)
	}
}

func (kb *virtualKeyboard) processNavigation(direction int) {
	kb.resetPressedKeys()

	var keyGrid [][]interface{}

	row1 := make([]interface{}, 0)

	numKeys := make([]key, 0)
	for i := range kb.Keys {

		if i < 10 {
			numKeys = append(numKeys, kb.Keys[i])
		}
	}

	for i := range numKeys {
		row1 = append(row1, i)
	}

	row1 = append(row1, "backspace")
	keyGrid = append(keyGrid, row1)

	row2 := make([]interface{}, 0)
	for i := 10; i < 20 && i < len(kb.Keys); i++ {
		row2 = append(row2, i)
	}
	keyGrid = append(keyGrid, row2)

	row3 := make([]interface{}, 0)
	for i := 20; i < 29 && i < len(kb.Keys); i++ {
		row3 = append(row3, i)
	}
	keyGrid = append(keyGrid, row3)

	row4 := make([]interface{}, 0)
	row4 = append(row4, "shift")
	for i := 29; i < 36 && i < len(kb.Keys); i++ {
		row4 = append(row4, i)
	}
	row4 = append(row4, "enter")
	keyGrid = append(keyGrid, row4)

	row5 := make([]interface{}, 0)
	row5 = append(row5, "space")
	keyGrid = append(keyGrid, row5)

	currentRow := -1
	currentCol := -1

	if kb.SelectedSpecial > 0 {
		var specialKeyName string
		switch kb.SelectedSpecial {
		case 1:
			specialKeyName = "backspace"
		case 2:
			specialKeyName = "enter"
		case 3:
			specialKeyName = "space"
		case 4:
			specialKeyName = "shift"
		}

		for r, row := range keyGrid {
			for c, key := range row {
				if str, ok := key.(string); ok && str == specialKeyName {
					currentRow = r
					currentCol = c
					break
				}
			}
			if currentRow >= 0 {
				break
			}
		}
	} else if kb.SelectedKeyIndex >= 0 {
		for r, row := range keyGrid {
			for c, key := range row {
				if idx, ok := key.(int); ok && idx == kb.SelectedKeyIndex {
					currentRow = r
					currentCol = c
					break
				}
			}
			if currentRow >= 0 {
				break
			}
		}
	}

	if currentRow < 0 || currentCol < 0 {
		if len(keyGrid) > 0 && len(keyGrid[0]) > 0 {
			currentRow = 0
			currentCol = 0
		} else {

			kb.SelectedKeyIndex = 0
			kb.SelectedSpecial = 0
			kb.Keys[kb.SelectedKeyIndex].IsPressed = true
			return
		}
	}

	newRow := currentRow
	newCol := currentCol

	switch direction {
	case 1:
		newCol++

		if newCol >= len(keyGrid[currentRow]) {
			newCol = 0
		}

	case 2:
		newCol--

		if newCol < 0 {
			newCol = len(keyGrid[currentRow]) - 1
		}

	case 3:
		newRow--

		if newRow < 0 {
			newRow = len(keyGrid) - 1
		}

		if newCol >= len(keyGrid[newRow]) {

			newCol = len(keyGrid[newRow]) - 1
		}

	case 4:
		newRow++

		if newRow >= len(keyGrid) {
			newRow = 0
		}

		if newCol >= len(keyGrid[newRow]) {

			newCol = len(keyGrid[newRow]) - 1
		}
	}

	if newRow >= 0 && newRow < len(keyGrid) && newCol >= 0 && newCol < len(keyGrid[newRow]) {
		selectedKey := keyGrid[newRow][newCol]

		if idx, ok := selectedKey.(int); ok {
			kb.SelectedKeyIndex = idx
			kb.SelectedSpecial = 0
			kb.Keys[kb.SelectedKeyIndex].IsPressed = true
		} else if str, ok := selectedKey.(string); ok {
			kb.SelectedKeyIndex = -1
			switch str {
			case "backspace":
				kb.SelectedSpecial = 1
			case "enter":
				kb.SelectedSpecial = 2
			case "space":
				kb.SelectedSpecial = 3
			case "shift":
				kb.SelectedSpecial = 4
			}
		}
	}
}

func (kb *virtualKeyboard) backspace() {
	if kb.CursorPosition > 0 {
		if kb.CursorPosition < len(kb.TextBuffer) {
			kb.TextBuffer = kb.TextBuffer[:kb.CursorPosition-1] + kb.TextBuffer[kb.CursorPosition:]
		} else {
			kb.TextBuffer = kb.TextBuffer[:len(kb.TextBuffer)-1]
		}
		kb.CursorPosition--
	}
}

func (kb *virtualKeyboard) insertSpace() {
	if kb.CursorPosition < len(kb.TextBuffer) {
		kb.TextBuffer = kb.TextBuffer[:kb.CursorPosition] + " " + kb.TextBuffer[kb.CursorPosition:]
	} else {
		kb.TextBuffer += " "
	}
	kb.CursorPosition++
}

func (kb *virtualKeyboard) toggleShift() {
	kb.ShiftPressed = !kb.ShiftPressed
	if kb.ShiftPressed {
		kb.CurrentState = upperCase
	} else {
		kb.CurrentState = lowerCase
	}
}

func (kb *virtualKeyboard) moveCursor(direction int) {
	if direction > 0 {
		if kb.CursorPosition < len(kb.TextBuffer) {
			kb.CursorPosition++
		}
	} else {
		if kb.CursorPosition > 0 {
			kb.CursorPosition--
		}
	}

	kb.CursorVisible = true
	kb.LastCursorBlink = time.Now()
}

func (kb *virtualKeyboard) updateCursorBlink() {
	if time.Since(kb.LastCursorBlink) > kb.CursorBlinkRate {
		kb.CursorVisible = !kb.CursorVisible
		kb.LastCursorBlink = time.Now()
	}
}

func (kb *virtualKeyboard) resetPressedKeys() {
	for i := range kb.Keys {
		kb.Keys[i].IsPressed = false
	}
}

func (kb *virtualKeyboard) handleKeyDown(key sdl.Keycode) bool {

	if key == sdl.K_h {
		kb.toggleHelp()
		return true
	}

	if kb.ShowingHelp {
		if key == sdl.K_UP {
			kb.scrollHelpOverlay(-1)
			return true
		}
		if key == sdl.K_DOWN {
			kb.scrollHelpOverlay(1)
			return true
		}

		if key != sdl.K_UP && key != sdl.K_DOWN {
			kb.ShowingHelp = false
			return true
		}
		return true
	}

	switch key {
	case sdl.K_UP, sdl.K_DOWN, sdl.K_LEFT, sdl.K_RIGHT:
		direction := 0
		switch key {
		case sdl.K_UP:
			direction = 3
		case sdl.K_DOWN:
			direction = 4
		case sdl.K_LEFT:
			direction = 2
		case sdl.K_RIGHT:
			direction = 1
		}
		kb.processNavigation(direction)
		return true

	case sdl.K_a:
		kb.processSelection()
		return true

	case sdl.K_b:
		kb.backspace()
		return true

	case sdl.K_LSHIFT, sdl.K_RSHIFT:
		kb.toggleShift()
		return true

	default:
		return false
	}
}

func (kb *virtualKeyboard) handleButtonPress(button uint8) bool {

	if button == BrickButton_MENU {
		kb.toggleHelp()
		return true
	}

	if kb.ShowingHelp {
		if button == BrickButton_UP {
			kb.scrollHelpOverlay(-1)
			return true
		}
		if button == BrickButton_DOWN {
			kb.scrollHelpOverlay(1)
			return true
		}

		kb.ShowingHelp = false
		return true
	}

	switch button {
	case BrickButton_UP, BrickButton_DOWN, BrickButton_LEFT, BrickButton_RIGHT:
		direction := 0
		switch button {
		case BrickButton_UP:
			direction = 3
		case BrickButton_DOWN:
			direction = 4
		case BrickButton_LEFT:
			direction = 2
		case BrickButton_RIGHT:
			direction = 1
		}
		kb.processNavigation(direction)
		return true

	case BrickButton_A:
		kb.processSelection()
		return true

	case BrickButton_B:
		kb.backspace()
		return true

	case BrickButton_X:
		kb.insertSpace()
		return true

	case BrickButton_SELECT:
		kb.toggleShift()
		return true

	case BrickButton_L1:
		kb.moveCursor(-1)
		return true

	case BrickButton_R1:
		kb.moveCursor(1)
		return true

	default:
		return false
	}
}

func (kb *virtualKeyboard) processSelection() {
	totalKeys := len(kb.Keys)

	if kb.SelectedKeyIndex >= 0 && kb.SelectedKeyIndex < totalKeys {

		var keyValue string

		if kb.SelectedKeyIndex < 10 && kb.ShiftPressed {
			keyValue = kb.Keys[kb.SelectedKeyIndex].SymbolValue
		} else if kb.CurrentState == upperCase {
			keyValue = kb.Keys[kb.SelectedKeyIndex].UpperValue
		} else {
			keyValue = kb.Keys[kb.SelectedKeyIndex].LowerValue
		}

		if kb.CursorPosition == len(kb.TextBuffer) {
			kb.TextBuffer += keyValue
		} else {

			textRunes := []rune(kb.TextBuffer)
			before := string(textRunes[:kb.CursorPosition])
			after := string(textRunes[kb.CursorPosition:])
			kb.TextBuffer = before + keyValue + after
		}
		kb.CursorPosition += len([]rune(keyValue))
	} else {

		switch kb.SelectedSpecial {
		case 1:
			if len(kb.TextBuffer) > 0 && kb.CursorPosition > 0 {
				textRunes := []rune(kb.TextBuffer)
				before := string(textRunes[:kb.CursorPosition-1])
				after := string(textRunes[kb.CursorPosition:])
				kb.TextBuffer = before + after
				kb.CursorPosition--
			}
		case 2:
			kb.EnterPressed = true
		case 3:
			if kb.CursorPosition == len(kb.TextBuffer) {
				kb.TextBuffer += " "
			} else {
				textRunes := []rune(kb.TextBuffer)
				before := string(textRunes[:kb.CursorPosition])
				after := string(textRunes[kb.CursorPosition:])
				kb.TextBuffer = before + " " + after
			}
			kb.CursorPosition++
		case 4:
			kb.ShiftPressed = !kb.ShiftPressed
			if kb.ShiftPressed {
				kb.CurrentState = upperCase
			} else {
				kb.CurrentState = lowerCase
			}
		}
	}

	kb.CursorVisible = true
	kb.LastCursorBlink = time.Now()
}

func (kb *virtualKeyboard) render(renderer *sdl.Renderer, font *ttf.Font) {

	if !kb.ShowingHelp {
		kb.renderKeyboard(renderer, font)
	}

	if kb.ShowingHelp && kb.helpOverlay != nil {
		kb.helpOverlay.render(renderer, internal.GetSmallFont())
	} else {

		kb.renderHelpPrompt(renderer, font)
	}
}

func (kb *virtualKeyboard) renderKeyboard(renderer *sdl.Renderer, font *ttf.Font) {

	renderer.SetDrawColor(50, 50, 50, 255)
	renderer.FillRect(&kb.TextInputRect)

	renderer.SetDrawColor(200, 200, 200, 255)
	renderer.DrawRect(&kb.TextInputRect)

	if kb.TextBuffer != "" {
		textColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
		textSurface, err := font.RenderUTF8Blended(kb.TextBuffer, textColor)
		if err != nil {
			return
		}

		textTexture, err := renderer.CreateTextureFromSurface(textSurface)
		if err != nil {
			textSurface.Free()
			return
		}

		var textWidth, textHeight = textSurface.W, textSurface.H

		var cursorX int32
		if kb.CursorPosition > 0 {

			cursorText := kb.TextBuffer[:kb.CursorPosition]
			cursorSurface, err := font.RenderUTF8Blended(cursorText, textColor)
			if err == nil {
				cursorX = cursorSurface.W
				cursorSurface.Free()
			}
		}

		padding := int32(10)
		textRect := sdl.Rect{
			X: kb.TextInputRect.X + padding,
			Y: kb.TextInputRect.Y + (kb.TextInputRect.H-textHeight)/2,
			W: textWidth,
			H: textHeight,
		}

		renderer.Copy(textTexture, nil, &textRect)

		if kb.CursorVisible {
			renderer.SetDrawColor(255, 255, 255, 255)
			cursorRect := sdl.Rect{
				X: textRect.X + cursorX,
				Y: textRect.Y,
				W: 2,
				H: textHeight,
			}
			renderer.FillRect(&cursorRect)
		}

		textTexture.Destroy()
		textSurface.Free()
	} else if kb.CursorVisible {

		padding := int32(10)
		placeholderHeight := int32(20)
		cursorRect := sdl.Rect{
			X: kb.TextInputRect.X + padding,
			Y: kb.TextInputRect.Y + (kb.TextInputRect.H-placeholderHeight)/2,
			W: 2,
			H: placeholderHeight,
		}
		renderer.SetDrawColor(255, 255, 255, 255)
		renderer.FillRect(&cursorRect)
	}

	for i, key := range kb.Keys {
		bgColor := sdl.Color{R: 50, G: 50, B: 60, A: 255}
		textColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}

		if i == kb.SelectedKeyIndex {
			bgColor = sdl.Color{R: 100, G: 100, B: 240, A: 255}
			textColor = sdl.Color{R: 255, G: 255, B: 255, A: 255}
		} else if key.IsPressed {
			bgColor = sdl.Color{R: 80, G: 80, B: 120, A: 255}
		}

		renderer.SetDrawColor(bgColor.R, bgColor.G, bgColor.B, bgColor.A)
		renderer.FillRect(&key.Rect)

		renderer.SetDrawColor(70, 70, 80, 255)
		renderer.DrawRect(&key.Rect)

		keyVal := key.LowerValue
		if kb.CurrentState == upperCase {
			if i < 10 && kb.ShiftPressed {
				keyVal = key.SymbolValue
			} else {
				keyVal = key.UpperValue
			}
		}

		textSurface, err := font.RenderUTF8Blended(keyVal, textColor)
		if err != nil {
			continue
		}

		textTexture, err := renderer.CreateTextureFromSurface(textSurface)
		if err != nil {
			textSurface.Free()
			continue
		}

		textRect := sdl.Rect{
			X: key.Rect.X + (key.Rect.W-textSurface.W)/2,
			Y: key.Rect.Y + (key.Rect.H-textSurface.H)/2,
			W: textSurface.W,
			H: textSurface.H,
		}

		renderer.Copy(textTexture, nil, &textRect)

		textTexture.Destroy()
		textSurface.Free()
	}

	backspaceBgColor := sdl.Color{R: 50, G: 50, B: 60, A: 255}
	if kb.SelectedSpecial == 1 {
		backspaceBgColor = sdl.Color{R: 100, G: 100, B: 240, A: 255}
	}

	renderer.SetDrawColor(backspaceBgColor.R, backspaceBgColor.G, backspaceBgColor.B, backspaceBgColor.A)
	renderer.FillRect(&kb.BackspaceRect)

	renderer.SetDrawColor(70, 70, 80, 255)
	renderer.DrawRect(&kb.BackspaceRect)

	backspaceText := "⌫"
	textColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
	backspaceSurface, err := internal.GetLargeSymbolFont().RenderUTF8Blended(backspaceText, textColor)
	if err == nil {
		backspaceTexture, err := renderer.CreateTextureFromSurface(backspaceSurface)
		if err == nil {
			backspaceRect := sdl.Rect{
				X: kb.BackspaceRect.X + (kb.BackspaceRect.W-backspaceSurface.W)/2,
				Y: kb.BackspaceRect.Y + (kb.BackspaceRect.H-backspaceSurface.H)/2,
				W: backspaceSurface.W,
				H: backspaceSurface.H,
			}
			renderer.Copy(backspaceTexture, nil, &backspaceRect)
			backspaceTexture.Destroy()
		}
		backspaceSurface.Free()
	}

	enterBgColor := sdl.Color{R: 50, G: 50, B: 60, A: 255}
	if kb.SelectedSpecial == 2 {
		enterBgColor = sdl.Color{R: 100, G: 100, B: 240, A: 255}
	}

	renderer.SetDrawColor(enterBgColor.R, enterBgColor.G, enterBgColor.B, enterBgColor.A)
	renderer.FillRect(&kb.EnterRect)

	renderer.SetDrawColor(70, 70, 80, 255)
	renderer.DrawRect(&kb.EnterRect)

	enterText := "↵"
	enterSurface, err := internal.GetLargeSymbolFont().RenderUTF8Blended(enterText, textColor)
	if err == nil {
		enterTexture, err := renderer.CreateTextureFromSurface(enterSurface)
		if err == nil {
			enterRect := sdl.Rect{
				X: kb.EnterRect.X + (kb.EnterRect.W-enterSurface.W)/2,
				Y: kb.EnterRect.Y + (kb.EnterRect.H-enterSurface.H)/2,
				W: enterSurface.W,
				H: enterSurface.H,
			}
			renderer.Copy(enterTexture, nil, &enterRect)
			enterTexture.Destroy()
		}
		enterSurface.Free()
	}

	spaceBgColor := sdl.Color{R: 50, G: 50, B: 60, A: 255}
	if kb.SelectedSpecial == 3 {
		spaceBgColor = sdl.Color{R: 100, G: 100, B: 240, A: 255}
	}

	renderer.SetDrawColor(spaceBgColor.R, spaceBgColor.G, spaceBgColor.B, spaceBgColor.A)
	renderer.FillRect(&kb.SpaceRect)

	renderer.SetDrawColor(70, 70, 80, 255)
	renderer.DrawRect(&kb.SpaceRect)

	lineWidth := kb.SpaceRect.W / 3
	lineHeight := int32(4)
	lineRect := sdl.Rect{
		X: kb.SpaceRect.X + (kb.SpaceRect.W-lineWidth)/2,
		Y: kb.SpaceRect.Y + (kb.SpaceRect.H-lineHeight)/2,
		W: lineWidth,
		H: lineHeight,
	}
	renderer.SetDrawColor(255, 255, 255, 255)
	renderer.FillRect(&lineRect)

	shiftBgColor := sdl.Color{R: 50, G: 50, B: 60, A: 255}
	if kb.SelectedSpecial == 4 || kb.CurrentState == upperCase {
		shiftBgColor = sdl.Color{R: 100, G: 100, B: 240, A: 255}
	}

	renderer.SetDrawColor(shiftBgColor.R, shiftBgColor.G, shiftBgColor.B, shiftBgColor.A)
	renderer.FillRect(&kb.ShiftRect)

	renderer.SetDrawColor(70, 70, 80, 255)
	renderer.DrawRect(&kb.ShiftRect)

	shiftText := "⇧"
	shiftSurface, err := internal.GetLargeSymbolFont().RenderUTF8Blended(shiftText, textColor)
	if err == nil {
		shiftTexture, err := renderer.CreateTextureFromSurface(shiftSurface)
		if err == nil {
			shiftRect := sdl.Rect{
				X: kb.ShiftRect.X + (kb.ShiftRect.W-shiftSurface.W)/2,
				Y: kb.ShiftRect.Y + (kb.ShiftRect.H-shiftSurface.H)/2,
				W: shiftSurface.W,
				H: shiftSurface.H,
			}
			renderer.Copy(shiftTexture, nil, &shiftRect)
			shiftTexture.Destroy()
		}
		shiftSurface.Free()
	}
}

func (kb *virtualKeyboard) renderHelpPrompt(renderer *sdl.Renderer, font *ttf.Font) {
	_, screenHeight, err := renderer.GetOutputSize()
	if err != nil {
		return
	}

	promptText := "Help (Menu)"

	promptColor := sdl.Color{R: 180, G: 180, B: 180, A: 200}
	promptSurface, err := font.RenderUTF8Blended(promptText, promptColor)
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
		X: padding,
		Y: screenHeight - promptSurface.H - padding,
		W: promptSurface.W,
		H: promptSurface.H,
	}

	renderer.Copy(promptTexture, nil, &promptRect)

	promptTexture.Destroy()
	promptSurface.Free()
}
