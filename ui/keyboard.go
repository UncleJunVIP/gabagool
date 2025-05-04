package ui

import (
	"github.com/UncleJunVIP/gabagool/internal"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"time"
)

type Key struct {
	Rect        sdl.Rect
	LowerValue  string
	UpperValue  string
	SymbolValue string // Will now be accessed when shift is pressed for the top row
	IsPressed   bool
}

type KeyboardState int

const (
	LowerCase KeyboardState = iota
	UpperCase
)

type VirtualKeyboard struct {
	Keys             []Key
	TextBuffer       string
	CurrentState     KeyboardState
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

	HelpLines        []string
	ShowingHelp      bool
	HelpScrollOffset int32
	MaxHelpScroll    int32
}

func CreateKeyboard(windowWidth, windowHeight int32) *VirtualKeyboard {
	kb := &VirtualKeyboard{
		Keys:             make([]Key, 0),
		TextBuffer:       "",
		CurrentState:     LowerCase,
		SelectedKeyIndex: 0,
		SelectedSpecial:  0,
		CursorPosition:   0,
		CursorVisible:    true,
		LastCursorBlink:  time.Now(),
		CursorBlinkRate:  500 * time.Millisecond,
		HelpLines: []string{
			"Navigation: D-Pad",
			"Move Cursor: L1/R1 Buttons",
			"Select / Type: A Button",
			"Backspace: B Button",
			"Shift: Select Button",
			"Space: X Button",
			"Enter: Start Button",
			"Exit Keyboard: Y Button",
			"Show / Hide Help: Menu Button",
		},
		ShowingHelp:      false,
		HelpScrollOffset: 0,
		MaxHelpScroll:    0,
	}

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
		kb.Keys = append(kb.Keys, Key{
			Rect:        sdl.Rect{X: x, Y: y, W: keyWidth, H: keyHeight},
			LowerValue:  keyVal,
			UpperValue:  keyVal,        // Same as lower when not shifted
			SymbolValue: rowSymbols[i], // Symbol value for when shift is pressed
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
		kb.Keys = append(kb.Keys, Key{
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
		kb.Keys = append(kb.Keys, Key{
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
		kb.Keys = append(kb.Keys, Key{
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

func NewBlockingKeyboard(initialText string) (string, error) {
	window := internal.GetWindow()
	renderer := window.Renderer
	font := internal.GetFont()

	kb := CreateKeyboard(window.Width, window.Height)

	if initialText != "" {
		kb.TextBuffer = initialText
		kb.CursorPosition = len(initialText)
	}

	running := true
	var result string
	var err error

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
				err = sdl.GetError()

			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {
					kb.HandleKeyDown(e.Keysym.Sym)

					if e.Keysym.Sym == sdl.K_RETURN {
						running = false
						result = kb.TextBuffer
						break
					} else if e.Keysym.Sym == sdl.K_ESCAPE {
						running = false
						result = initialText
						break
					}
				}

			case *sdl.ControllerButtonEvent:
				if e.Type == sdl.CONTROLLERBUTTONDOWN {
					kb.HandleButtonPress(e.Button)

					if e.Button == BrickButton_START {
						running = false
						result = kb.TextBuffer
						break
					} else if e.Button == BrickButton_Y {
						running = false
						result = ""
						break
					}
				}

			}
		}

		kb.UpdateCursorBlink()

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		kb.Render(renderer, font)

		renderer.Present()

		sdl.Delay(16)
	}

	return result, err
}

func (kb *VirtualKeyboard) ToggleHelp() {
	kb.ShowingHelp = !kb.ShowingHelp
	kb.HelpScrollOffset = 0
}

func (kb *VirtualKeyboard) RenderHelpPrompt(renderer *sdl.Renderer, font *ttf.Font) {
	_, screenHeight, err := renderer.GetOutputSize()
	if err != nil {
		return
	}

	if !kb.ShowingHelp {
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
}

func (kb *VirtualKeyboard) ScrollHelpOverlay(direction int32) {
	newOffset := kb.HelpScrollOffset + direction

	if newOffset < 0 {
		newOffset = 0
	}

	if newOffset > kb.MaxHelpScroll {
		newOffset = kb.MaxHelpScroll
	}

	kb.HelpScrollOffset = newOffset
}

func (kb *VirtualKeyboard) ProcessNavigation(direction int) {
	kb.ResetPressedKeys()

	var keyGrid [][]interface{}

	row1 := make([]interface{}, 0)

	numKeys := make([]Key, 0)
	for i := range kb.Keys {
		// Assume number keys are in the top row (lowest Y value)
		if i < 10 { // First 10 keys are typically 1-0
			numKeys = append(numKeys, kb.Keys[i])
		}
	}

	for i := range numKeys {
		row1 = append(row1, i) // Store index of key
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
			// Fallback to first key
			kb.SelectedKeyIndex = 0
			kb.SelectedSpecial = 0
			kb.Keys[kb.SelectedKeyIndex].IsPressed = true
			return
		}
	}

	newRow := currentRow
	newCol := currentCol

	switch direction {
	case 1: // Right
		newCol++
		// Loop around if needed
		if newCol >= len(keyGrid[currentRow]) {
			newCol = 0
		}

	case 2: // Left
		newCol--
		// Loop around if needed
		if newCol < 0 {
			newCol = len(keyGrid[currentRow]) - 1
		}

	case 3: // Up
		newRow--
		// Loop around if needed
		if newRow < 0 {
			newRow = len(keyGrid) - 1
		}

		// Adjust column to nearest button in the row
		if newCol >= len(keyGrid[newRow]) {
			// Find the closest column
			newCol = len(keyGrid[newRow]) - 1
		}

	case 4: // Down
		newRow++
		// Loop around if needed
		if newRow >= len(keyGrid) {
			newRow = 0
		}

		// Adjust column to nearest button in the row
		if newCol >= len(keyGrid[newRow]) {
			// Find the closest column
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

func (kb *VirtualKeyboard) MoveCursor(direction int) {
	if direction > 0 { // Move right
		if kb.CursorPosition < len(kb.TextBuffer) {
			kb.CursorPosition++
		}
	} else { // Move left
		if kb.CursorPosition > 0 {
			kb.CursorPosition--
		}
	}

	kb.CursorVisible = true
	kb.LastCursorBlink = time.Now()
}

func (kb *VirtualKeyboard) UpdateCursorBlink() {
	if time.Since(kb.LastCursorBlink) > kb.CursorBlinkRate {
		kb.CursorVisible = !kb.CursorVisible
		kb.LastCursorBlink = time.Now()
	}
}

func (kb *VirtualKeyboard) ResetPressedKeys() {
	for i := range kb.Keys {
		kb.Keys[i].IsPressed = false
	}
}

func (kb *VirtualKeyboard) HandleKeyDown(keyCode sdl.Keycode) {
	if keyCode == sdl.K_h || keyCode == sdl.K_QUESTION {
		kb.ToggleHelp()
		return
	}

	if kb.ShowingHelp {
		if keyCode == sdl.K_UP {
			kb.ScrollHelpOverlay(-1) // Scroll up
			return
		}
		if keyCode == sdl.K_DOWN {
			kb.ScrollHelpOverlay(1) // Scroll down
			return
		}

		if keyCode != sdl.K_UP && keyCode != sdl.K_DOWN {
			kb.ShowingHelp = false
		}
		return
	}

	switch keyCode {
	case sdl.K_BACKSPACE:
		if len(kb.TextBuffer) > 0 && kb.CursorPosition > 0 {
			textRunes := []rune(kb.TextBuffer)
			before := string(textRunes[:kb.CursorPosition-1])
			after := string(textRunes[kb.CursorPosition:])
			kb.TextBuffer = before + after
			kb.CursorPosition--
		}
		return

	case sdl.K_RETURN:
		if kb.SelectedKeyIndex >= 0 || kb.SelectedSpecial > 0 {
			kb.ProcessSelection()
		} else {
			if kb.CursorPosition == len(kb.TextBuffer) {
				kb.TextBuffer += "\n"
			} else {
				textRunes := []rune(kb.TextBuffer)
				before := string(textRunes[:kb.CursorPosition])
				after := string(textRunes[kb.CursorPosition:])
				kb.TextBuffer = before + "\n" + after
			}
			kb.CursorPosition++
		}
		return

	case sdl.K_SPACE:
		if kb.SelectedKeyIndex >= 0 || kb.SelectedSpecial > 0 {
			kb.ProcessSelection()
		} else {
			if kb.CursorPosition == len(kb.TextBuffer) {
				kb.TextBuffer += " "
			} else {
				textRunes := []rune(kb.TextBuffer)
				before := string(textRunes[:kb.CursorPosition])
				after := string(textRunes[kb.CursorPosition:])
				kb.TextBuffer = before + " " + after
			}
			kb.CursorPosition++
		}
		return

	case sdl.K_LSHIFT, sdl.K_RSHIFT:
		kb.ShiftPressed = !kb.ShiftPressed
		if kb.ShiftPressed {
			kb.CurrentState = UpperCase
		} else {
			kb.CurrentState = LowerCase
		}
		return

	case sdl.K_UP:
		kb.ProcessNavigation(3)
		return
	case sdl.K_DOWN:
		kb.ProcessNavigation(4)
		return
	case sdl.K_LEFT:
		kb.ProcessNavigation(2)
		return
	case sdl.K_RIGHT:
		kb.ProcessNavigation(1)
		return

	case sdl.K_HOME:
		kb.CursorPosition = 0
		return
	case sdl.K_END:
		kb.CursorPosition = len(kb.TextBuffer)
		return
	}

	var r rune
	switch {
	case keyCode >= sdl.K_a && keyCode <= sdl.K_z:
		if kb.CurrentState == UpperCase {
			r = rune(keyCode - sdl.K_a + 'A')
		} else {
			r = rune(keyCode - sdl.K_a + 'a')
		}

		if kb.CursorPosition == len(kb.TextBuffer) {
			kb.TextBuffer += string(r)
		} else {
			textRunes := []rune(kb.TextBuffer)
			before := string(textRunes[:kb.CursorPosition])
			after := string(textRunes[kb.CursorPosition:])
			kb.TextBuffer = before + string(r) + after
		}
		kb.CursorPosition++

	case keyCode >= sdl.K_0 && keyCode <= sdl.K_9:
		if kb.ShiftPressed {
			symbols := []string{")", "!", "@", "#", "$", "%", "^", "&", "*", "("}
			idx := int(keyCode - sdl.K_0)
			if idx == 0 {
				idx = 9
			} else {
				idx = idx - 1
			}

			if kb.CursorPosition == len(kb.TextBuffer) {
				kb.TextBuffer += symbols[idx]
			} else {
				textRunes := []rune(kb.TextBuffer)
				before := string(textRunes[:kb.CursorPosition])
				after := string(textRunes[kb.CursorPosition:])
				kb.TextBuffer = before + symbols[idx] + after
			}
			kb.CursorPosition++
		} else {
			r = rune(keyCode - sdl.K_0 + '0')
			if kb.CursorPosition == len(kb.TextBuffer) {
				kb.TextBuffer += string(r)
			} else {
				textRunes := []rune(kb.TextBuffer)
				before := string(textRunes[:kb.CursorPosition])
				after := string(textRunes[kb.CursorPosition:])
				kb.TextBuffer = before + string(r) + after
			}
			kb.CursorPosition++
		}
	}

	kb.CursorVisible = true
	kb.LastCursorBlink = time.Now()
}

func (kb *VirtualKeyboard) HandleButtonPress(button uint8) {
	if button == BrickButton_MENU {
		kb.ToggleHelp()
		return
	}

	if kb.ShowingHelp {
		if button == BrickButton_UP {
			kb.ScrollHelpOverlay(-1)
			return
		}
		if button == BrickButton_DOWN {
			kb.ScrollHelpOverlay(1)
			return
		}

		kb.ShowingHelp = false
		return
	}

	switch button {
	case BrickButton_UP:
		kb.ProcessNavigation(3)
	case BrickButton_DOWN:
		kb.ProcessNavigation(4)
	case BrickButton_LEFT:
		kb.ProcessNavigation(2)
	case BrickButton_RIGHT:
		kb.ProcessNavigation(1)
	case BrickButton_A:
		kb.ProcessSelection()
	case BrickButton_B:
		// Fix: Delete at cursor position instead of at the end
		if len(kb.TextBuffer) > 0 && kb.CursorPosition > 0 {
			textRunes := []rune(kb.TextBuffer)
			before := string(textRunes[:kb.CursorPosition-1])
			after := string(textRunes[kb.CursorPosition:])
			kb.TextBuffer = before + after
			kb.CursorPosition--

			// Reset cursor blink when deleting
			kb.CursorVisible = true
			kb.LastCursorBlink = time.Now()
		}

	case BrickButton_L1:
		kb.MoveCursor(-1)
	case BrickButton_R1:
		kb.MoveCursor(1)
	case BrickButton_SELECT:
		kb.ShiftPressed = !kb.ShiftPressed
		if kb.ShiftPressed {
			kb.CurrentState = UpperCase
		} else {
			kb.CurrentState = LowerCase
		}
	case BrickButton_X:
		if kb.CursorPosition == len(kb.TextBuffer) {
			kb.TextBuffer += " "
		} else {
			textRunes := []rune(kb.TextBuffer)
			before := string(textRunes[:kb.CursorPosition])
			after := string(textRunes[kb.CursorPosition:])
			kb.TextBuffer = before + " " + after
		}
		kb.CursorPosition++

		kb.CursorVisible = true
		kb.LastCursorBlink = time.Now()

	}
}

func (kb *VirtualKeyboard) ProcessSelection() {
	totalKeys := len(kb.Keys)

	if kb.SelectedKeyIndex >= 0 && kb.SelectedKeyIndex < totalKeys {
		// Regular key
		var keyValue string

		// For number keys (0-9) with shift pressed, use the symbol value
		if kb.SelectedKeyIndex < 10 && kb.ShiftPressed {
			keyValue = kb.Keys[kb.SelectedKeyIndex].SymbolValue
		} else if kb.CurrentState == UpperCase {
			keyValue = kb.Keys[kb.SelectedKeyIndex].UpperValue
		} else {
			keyValue = kb.Keys[kb.SelectedKeyIndex].LowerValue
		}

		// Insert at cursor position instead of appending
		if kb.CursorPosition == len(kb.TextBuffer) {
			kb.TextBuffer += keyValue
		} else {
			// Split text at cursor position and insert the new character
			textRunes := []rune(kb.TextBuffer)
			before := string(textRunes[:kb.CursorPosition])
			after := string(textRunes[kb.CursorPosition:])
			kb.TextBuffer = before + keyValue + after
		}
		kb.CursorPosition += len([]rune(keyValue))
	} else {
		// Special key
		switch kb.SelectedSpecial {
		case 1: // Backspace
			if len(kb.TextBuffer) > 0 && kb.CursorPosition > 0 {
				textRunes := []rune(kb.TextBuffer)
				before := string(textRunes[:kb.CursorPosition-1])
				after := string(textRunes[kb.CursorPosition:])
				kb.TextBuffer = before + after
				kb.CursorPosition--
			}
		case 2: // Enter
			if kb.CursorPosition == len(kb.TextBuffer) {
				kb.TextBuffer += "\n"
			} else {
				textRunes := []rune(kb.TextBuffer)
				before := string(textRunes[:kb.CursorPosition])
				after := string(textRunes[kb.CursorPosition:])
				kb.TextBuffer = before + "\n" + after
			}
			kb.CursorPosition++
		case 3: // Space
			if kb.CursorPosition == len(kb.TextBuffer) {
				kb.TextBuffer += " "
			} else {
				textRunes := []rune(kb.TextBuffer)
				before := string(textRunes[:kb.CursorPosition])
				after := string(textRunes[kb.CursorPosition:])
				kb.TextBuffer = before + " " + after
			}
			kb.CursorPosition++
		case 4: // Shift
			kb.ShiftPressed = !kb.ShiftPressed
			if kb.ShiftPressed {
				kb.CurrentState = UpperCase
			} else {
				kb.CurrentState = LowerCase
			}
		}
	}

	kb.CursorVisible = true
	kb.LastCursorBlink = time.Now()
}

func (kb *VirtualKeyboard) Render(renderer *sdl.Renderer, font *ttf.Font) {
	kb.UpdateCursorBlink()

	renderer.SetDrawColor(40, 40, 40, 255)
	renderer.FillRect(&kb.TextInputRect)
	renderer.SetDrawColor(80, 80, 80, 255)
	renderer.DrawRect(&kb.TextInputRect)

	textColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}

	if kb.TextBuffer != "" {
		textSurface, err := font.RenderUTF8BlendedWrapped(kb.TextBuffer, textColor, int(kb.TextInputRect.W-20))
		if err == nil && textSurface != nil {
			textTexture, _ := renderer.CreateTextureFromSurface(textSurface)
			textRect := sdl.Rect{
				X: kb.TextInputRect.X + 10,
				Y: kb.TextInputRect.Y + 10,
				W: textSurface.W,
				H: textSurface.H,
			}
			renderer.Copy(textTexture, nil, &textRect)

			if kb.CursorVisible {
				cursorX, cursorY := kb.TextInputRect.X+10, kb.TextInputRect.Y+10

				if kb.CursorPosition > 0 {
					cursorText := string([]rune(kb.TextBuffer)[:kb.CursorPosition])
					measureSurface, _ := font.RenderUTF8Blended(cursorText, textColor)
					if measureSurface != nil {
						lineHeight := int32(font.Height())
						fullLines := 0
						lastLinebreak := 0

						for i, char := range cursorText {
							if char == '\n' {
								fullLines++
								lastLinebreak = i + 1
							}
						}

						lineText := cursorText
						if lastLinebreak > 0 {
							lineText = cursorText[lastLinebreak:]
						}

						lineSurface, _ := font.RenderUTF8Blended(lineText, textColor)
						if lineSurface != nil {
							cursorX = kb.TextInputRect.X + 10 + lineSurface.W
							cursorY = kb.TextInputRect.Y + 10 + int32(fullLines)*lineHeight
							lineSurface.Free()
						}

						measureSurface.Free()
					}
				}

				renderer.SetDrawColor(255, 255, 255, 255)
				cursorRect := sdl.Rect{
					X: cursorX,
					Y: cursorY,
					W: 2,
					H: int32(font.Height()),
				}
				renderer.FillRect(&cursorRect)
			}

			textSurface.Free()
			textTexture.Destroy()
		}
	} else {
		if kb.CursorVisible {
			renderer.SetDrawColor(255, 255, 255, 255)
			cursorRect := sdl.Rect{
				X: kb.TextInputRect.X + 10,
				Y: kb.TextInputRect.Y + 10,
				W: 2,
				H: int32(font.Height()),
			}
			renderer.FillRect(&cursorRect)
		}
	}

	for i, key := range kb.Keys {
		if key.IsPressed || i == kb.SelectedKeyIndex {
			renderer.SetDrawColor(100, 100, 200, 255)
		} else {
			renderer.SetDrawColor(60, 60, 60, 255)
		}
		renderer.FillRect(&key.Rect)

		renderer.SetDrawColor(120, 120, 120, 255)
		renderer.DrawRect(&key.Rect)

		var keyText string

		if i < 10 && kb.ShiftPressed {
			keyText = key.SymbolValue
		} else if kb.CurrentState == UpperCase {
			keyText = key.UpperValue
		} else {
			keyText = key.LowerValue
		}

		textSurface, _ := font.RenderUTF8Blended(keyText, textColor)
		if textSurface != nil {
			textTexture, _ := renderer.CreateTextureFromSurface(textSurface)
			textRect := sdl.Rect{
				X: key.Rect.X + (key.Rect.W-textSurface.W)/2,
				Y: key.Rect.Y + (key.Rect.H-textSurface.H)/2,
				W: textSurface.W,
				H: textSurface.H,
			}
			renderer.Copy(textTexture, nil, &textRect)
			textSurface.Free()
			textTexture.Destroy()
		}
	}

	if kb.SelectedSpecial == 1 {
		renderer.SetDrawColor(100, 100, 200, 255)
	} else {
		renderer.SetDrawColor(60, 60, 60, 255)
	}
	renderer.FillRect(&kb.BackspaceRect)
	renderer.SetDrawColor(120, 120, 120, 255)
	renderer.DrawRect(&kb.BackspaceRect)

	textSurface, _ := font.RenderUTF8Blended("⌫", textColor)
	if textSurface != nil {
		textTexture, _ := renderer.CreateTextureFromSurface(textSurface)
		textRect := sdl.Rect{
			X: kb.BackspaceRect.X + (kb.BackspaceRect.W-textSurface.W)/2,
			Y: kb.BackspaceRect.Y + (kb.BackspaceRect.H-textSurface.H)/2,
			W: textSurface.W,
			H: textSurface.H,
		}
		renderer.Copy(textTexture, nil, &textRect)
		textSurface.Free()
		textTexture.Destroy()
	}

	if kb.SelectedSpecial == 2 {
		renderer.SetDrawColor(100, 100, 200, 255)
	} else {
		renderer.SetDrawColor(60, 60, 60, 255)
	}
	renderer.FillRect(&kb.EnterRect)
	renderer.SetDrawColor(120, 120, 120, 255)
	renderer.DrawRect(&kb.EnterRect)

	textSurface, _ = font.RenderUTF8Blended("⏎", textColor)
	if textSurface != nil {
		textTexture, _ := renderer.CreateTextureFromSurface(textSurface)
		textRect := sdl.Rect{
			X: kb.EnterRect.X + (kb.EnterRect.W-textSurface.W)/2,
			Y: kb.EnterRect.Y + (kb.EnterRect.H-textSurface.H)/2,
			W: textSurface.W,
			H: textSurface.H,
		}
		renderer.Copy(textTexture, nil, &textRect)
		textSurface.Free()
		textTexture.Destroy()
	}

	if kb.SelectedSpecial == 3 {
		renderer.SetDrawColor(100, 100, 200, 255)
	} else {
		renderer.SetDrawColor(60, 60, 60, 255)
	}
	renderer.FillRect(&kb.SpaceRect)
	renderer.SetDrawColor(120, 120, 120, 255)
	renderer.DrawRect(&kb.SpaceRect)

	if kb.ShiftPressed || kb.SelectedSpecial == 4 {
		renderer.SetDrawColor(100, 100, 200, 255)
	} else {
		renderer.SetDrawColor(60, 60, 60, 255)
	}
	renderer.FillRect(&kb.ShiftRect)
	renderer.SetDrawColor(120, 120, 120, 255)
	renderer.DrawRect(&kb.ShiftRect)

	textSurface, _ = font.RenderUTF8Blended("⇧", textColor)
	if textSurface != nil {
		textTexture, _ := renderer.CreateTextureFromSurface(textSurface)
		textRect := sdl.Rect{
			X: kb.ShiftRect.X + (kb.ShiftRect.W-textSurface.W)/2,
			Y: kb.ShiftRect.Y + (kb.ShiftRect.H-textSurface.H)/2,
			W: textSurface.W,
			H: textSurface.H,
		}
		renderer.Copy(textTexture, nil, &textRect)
		textSurface.Free()
		textTexture.Destroy()
	}

	kb.RenderHelpPrompt(renderer, internal.GetSmallFont())

	if kb.ShowingHelp {
		overlayRect := sdl.Rect{
			X: 0,
			Y: 0,
			W: renderer.GetViewport().W,
			H: renderer.GetViewport().H,
		}
		renderer.SetDrawColor(0, 0, 0, 200) // Black with alpha
		renderer.FillRect(&overlayRect)

		helpWidth := overlayRect.W * 80 / 100
		helpHeight := overlayRect.H * 80 / 100
		helpX := (overlayRect.W - helpWidth) / 2
		helpY := (overlayRect.H - helpHeight) / 2
		helpRect := sdl.Rect{X: helpX, Y: helpY, W: helpWidth, H: helpHeight}

		renderer.SetDrawColor(40, 40, 40, 255)
		renderer.FillRect(&helpRect)

		renderer.SetDrawColor(120, 120, 120, 255)
		renderer.DrawRect(&helpRect)

		titleText := "Keyboard Help"
		titleColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
		titleSurface, err := font.RenderUTF8Blended(titleText, titleColor)
		if err == nil && titleSurface != nil {
			titleTexture, _ := renderer.CreateTextureFromSurface(titleSurface)
			titleRect := sdl.Rect{
				X: helpX + (helpWidth-titleSurface.W)/2,
				Y: helpY + 20,
				W: titleSurface.W,
				H: titleSurface.H,
			}
			renderer.Copy(titleTexture, nil, &titleRect)
			titleSurface.Free()
			titleTexture.Destroy()
		}

		contentY := helpY + 80
		lineHeight := int32(font.Height() + 10)

		totalContentHeight := lineHeight * int32(len(kb.HelpLines))
		visibleContentHeight := helpHeight - 120 // Account for padding and title
		kb.MaxHelpScroll = int32(0)

		if totalContentHeight > visibleContentHeight {
			kb.MaxHelpScroll = (totalContentHeight - visibleContentHeight) / lineHeight
		}

		textColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
		startLine := kb.HelpScrollOffset
		endLine := startLine + (visibleContentHeight / lineHeight)

		if endLine > int32(len(kb.HelpLines)) {
			endLine = int32(len(kb.HelpLines))
		}

		for i := startLine; i < endLine; i++ {
			if i >= 0 && int(i) < len(kb.HelpLines) {
				lineSurface, err := font.RenderUTF8Blended(kb.HelpLines[i], textColor)
				if err == nil && lineSurface != nil {
					lineTexture, _ := renderer.CreateTextureFromSurface(lineSurface)
					lineRect := sdl.Rect{
						X: helpX + 30,
						Y: contentY + (i-startLine)*lineHeight,
						W: lineSurface.W,
						H: lineSurface.H,
					}
					renderer.Copy(lineTexture, nil, &lineRect)
					lineSurface.Free()
					lineTexture.Destroy()
				}
			}
		}

		if kb.MaxHelpScroll > 0 {
			if kb.HelpScrollOffset > 0 {
				upArrow := "▲ More"
				arrowSurface, _ := font.RenderUTF8Blended(upArrow, textColor)
				if arrowSurface != nil {
					arrowTexture, _ := renderer.CreateTextureFromSurface(arrowSurface)
					arrowRect := sdl.Rect{
						X: helpX + helpWidth - arrowSurface.W - 20,
						Y: helpY + 30,
						W: arrowSurface.W,
						H: arrowSurface.H,
					}
					renderer.Copy(arrowTexture, nil, &arrowRect)
					arrowSurface.Free()
					arrowTexture.Destroy()
				}
			}

			if kb.HelpScrollOffset < kb.MaxHelpScroll {
				downArrow := "▼ More"
				arrowSurface, _ := font.RenderUTF8Blended(downArrow, textColor)
				if arrowSurface != nil {
					arrowTexture, _ := renderer.CreateTextureFromSurface(arrowSurface)
					arrowRect := sdl.Rect{
						X: helpX + helpWidth - arrowSurface.W - 20,
						Y: helpY + helpHeight - arrowSurface.H - 30,
						W: arrowSurface.W,
						H: arrowSurface.H,
					}
					renderer.Copy(arrowTexture, nil, &arrowRect)
					arrowSurface.Free()
					arrowTexture.Destroy()
				}
			}
		}
	}
}
