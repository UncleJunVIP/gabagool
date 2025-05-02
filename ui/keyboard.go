package ui

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// Key represents a single key on the virtual keyboard
type Key struct {
	Rect        sdl.Rect
	LowerValue  string
	UpperValue  string
	SymbolValue string // Will now be accessed when shift is pressed for the top row
	IsPressed   bool
}

// KeyboardState represents the state of the keyboard
type KeyboardState int

const (
	LowerCase KeyboardState = iota
	UpperCase
)

// VirtualKeyboard represents the entire virtual keyboard
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
	KeyboardRect     sdl.Rect // Overall keyboard rectangle
	SelectedKeyIndex int      // Index of currently selected key (-1 for none)
	SelectedSpecial  int      // Special key selection (0=none, 1=backspace, 2=enter, 3=space, 4=shift)
}

// CreateKeyboard initializes a new virtual keyboard
func CreateKeyboard(windowWidth, windowHeight int32) *VirtualKeyboard {
	kb := &VirtualKeyboard{
		Keys:             make([]Key, 0),
		TextBuffer:       "",
		CurrentState:     LowerCase,
		SelectedKeyIndex: 0,
		SelectedSpecial:  0,
	}

	// Calculate 85% of screen dimensions for the usable area (not including text input)
	keyboardWidth := (windowWidth * 85) / 100
	keyboardHeight := (windowHeight * 85) / 100

	// Calculate text input height (smaller than regular keyboard area)
	textInputHeight := int32(windowHeight / 10)

	// Adjust keyboard height to account for text input area
	keyboardHeight = keyboardHeight - textInputHeight - 20 // 20 pixels for spacing

	// Center the keyboard and text input horizontally
	startX := (windowWidth - keyboardWidth) / 2

	// Position keyboard and text input vertically with proper spacing
	textInputY := (windowHeight - keyboardHeight - textInputHeight - 20) / 2
	keyboardStartY := textInputY + textInputHeight + 20

	// Set overall keyboard rectangle (only covers the actual keyboard, not text input)
	kb.KeyboardRect = sdl.Rect{
		X: startX,
		Y: keyboardStartY,
		W: keyboardWidth,
		H: keyboardHeight,
	}

	// Define keyboard dimensions and spacing
	keyWidth := int32(keyboardWidth / 12)
	keyHeight := int32(keyboardHeight / 6) // Adjusted for better key proportions
	keySpacing := int32(6)

	// Define text input box (centered above keyboard)
	kb.TextInputRect = sdl.Rect{
		X: startX,
		Y: textInputY,
		W: keyboardWidth,
		H: textInputHeight,
	}

	// First row: numbers and symbols (1-10)
	rowKeys := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
	// Symbols that typically appear on the number keys of a QWERTY keyboard
	rowSymbols := []string{"!", "@", "#", "$", "%", "^", "&", "*", "(", ")"}

	// Calculate total row width
	rowWidth := (keyWidth * int32(len(rowKeys))) + (keySpacing * int32(len(rowKeys)-1))
	// Center the row
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

	// Add backspace key
	kb.BackspaceRect = sdl.Rect{
		X: x,
		Y: y,
		W: keyWidth * 2,
		H: keyHeight,
	}

	// Second row: QWERTYUIOP
	rowKeys = []string{"q", "w", "e", "r", "t", "y", "u", "i", "o", "p"}

	// Calculate total row width
	rowWidth = (keyWidth * int32(len(rowKeys))) + (keySpacing * int32(len(rowKeys)-1))
	// Center the row
	rowStartX = startX + (keyboardWidth-rowWidth)/2

	x = rowStartX
	y += keyHeight + keySpacing

	for _, keyVal := range rowKeys {
		kb.Keys = append(kb.Keys, Key{
			Rect:        sdl.Rect{X: x, Y: y, W: keyWidth, H: keyHeight},
			LowerValue:  keyVal,
			UpperValue:  string([]rune(keyVal)[0] - 32), // Convert to uppercase
			SymbolValue: "",                             // Not used anymore
			IsPressed:   false,
		})
		x += keyWidth + keySpacing
	}

	// Third row: ASDFGHJKL
	rowKeys = []string{"a", "s", "d", "f", "g", "h", "j", "k", "l"}

	// Calculate total row width
	rowWidth = (keyWidth * int32(len(rowKeys))) + (keySpacing * int32(len(rowKeys)-1))
	// Center the row
	rowStartX = startX + (keyboardWidth-rowWidth)/2

	x = rowStartX
	y += keyHeight + keySpacing

	for _, keyVal := range rowKeys {
		kb.Keys = append(kb.Keys, Key{
			Rect:        sdl.Rect{X: x, Y: y, W: keyWidth, H: keyHeight},
			LowerValue:  keyVal,
			UpperValue:  string([]rune(keyVal)[0] - 32), // Convert to uppercase
			SymbolValue: "",                             // Not used anymore
			IsPressed:   false,
		})
		x += keyWidth + keySpacing
	}

	// Fourth row: ZXCVBNM + special keys
	rowKeys = []string{"z", "x", "c", "v", "b", "n", "m"}

	// Calculate space needed for Shift key on the left
	shiftKeyWidth := keyWidth * 2

	// Calculate total row width for regular keys
	regularKeysWidth := (keyWidth * int32(len(rowKeys))) + (keySpacing * int32(len(rowKeys)-1))

	// Calculate space for Enter key on the right
	enterKeyWidth := keyWidth + keyWidth/2

	// Calculate total fourth row width
	totalFourthRowWidth := shiftKeyWidth + regularKeysWidth + enterKeyWidth + keySpacing*2

	// Center the fourth row
	fourthRowStartX := startX + (keyboardWidth-totalFourthRowWidth)/2

	// Add shift key
	kb.ShiftRect = sdl.Rect{
		X: fourthRowStartX,
		Y: y + keyHeight + keySpacing,
		W: shiftKeyWidth,
		H: keyHeight,
	}

	x = kb.ShiftRect.X + kb.ShiftRect.W + keySpacing

	for _, keyVal := range rowKeys {
		kb.Keys = append(kb.Keys, Key{
			Rect:        sdl.Rect{X: x, Y: y + keyHeight + keySpacing, W: keyWidth, H: keyHeight},
			LowerValue:  keyVal,
			UpperValue:  string([]rune(keyVal)[0] - 32), // Convert to uppercase
			SymbolValue: "",                             // Not used anymore
			IsPressed:   false,
		})
		x += keyWidth + keySpacing
	}

	// Add enter key
	kb.EnterRect = sdl.Rect{
		X: x,
		Y: y + keyHeight + keySpacing,
		W: enterKeyWidth,
		H: keyHeight,
	}

	// Add space bar - center it in the bottom row
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

// ProcessNavigation handles navigation via keyboard or controller
func (kb *VirtualKeyboard) ProcessNavigation(direction int) {
	// Reset all pressed keys to fix lingering highlights
	kb.ResetPressedKeys()

	// Define all keys including special keys as a grid
	// This helps us organize the keyboard layout for easier navigation
	var keyGrid [][]interface{}

	// Row 1: 1-0 and delete
	row1 := make([]interface{}, 0)

	// Find all number keys (row 1)
	numKeys := make([]Key, 0)
	for i := range kb.Keys {
		// Assume number keys are in the top row (lowest Y value)
		if i < 10 { // First 10 keys are typically 1-0
			numKeys = append(numKeys, kb.Keys[i])
		}
	}

	// Add all number keys to row 1
	for i := range numKeys {
		row1 = append(row1, i) // Store index of key
	}

	// Add backspace (special key 1) to row 1
	row1 = append(row1, "backspace")
	keyGrid = append(keyGrid, row1)

	// Row 2: qwertyuiop
	row2 := make([]interface{}, 0)
	for i := 10; i < 20 && i < len(kb.Keys); i++ {
		row2 = append(row2, i)
	}
	keyGrid = append(keyGrid, row2)

	// Row 3: asdfghjkl
	row3 := make([]interface{}, 0)
	for i := 20; i < 29 && i < len(kb.Keys); i++ {
		row3 = append(row3, i)
	}
	keyGrid = append(keyGrid, row3)

	// Row 4: Shift, zxcvbnm, Enter
	row4 := make([]interface{}, 0)
	row4 = append(row4, "shift") // Shift (special key 4)
	for i := 29; i < 36 && i < len(kb.Keys); i++ {
		row4 = append(row4, i)
	}
	row4 = append(row4, "enter") // Enter (special key 2)
	keyGrid = append(keyGrid, row4)

	// Row 5: Space
	row5 := make([]interface{}, 0)
	row5 = append(row5, "space") // Space (special key 3)
	keyGrid = append(keyGrid, row5)

	// Find current position in grid
	currentRow := -1
	currentCol := -1

	// Check if a special key is selected
	if kb.SelectedSpecial > 0 {
		// Convert special key index to name
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

		// Find this special key in the grid
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
		// Find the regular key in the grid
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

	// If no current selection, start at the top-left
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

	// Determine next position based on direction
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

	// Update selected key based on new position
	if newRow >= 0 && newRow < len(keyGrid) && newCol >= 0 && newCol < len(keyGrid[newRow]) {
		selectedKey := keyGrid[newRow][newCol]

		// Update the selection based on the type of key
		if idx, ok := selectedKey.(int); ok {
			// It's a regular key
			kb.SelectedKeyIndex = idx
			kb.SelectedSpecial = 0
			kb.Keys[kb.SelectedKeyIndex].IsPressed = true
		} else if str, ok := selectedKey.(string); ok {
			// It's a special key
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

// ResetPressedKeys resets all pressed keys
func (kb *VirtualKeyboard) ResetPressedKeys() {
	for i := range kb.Keys {
		kb.Keys[i].IsPressed = false
	}
}

// ProcessSelection handles selection via keyboard or controller
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

		kb.TextBuffer += keyValue
	} else {
		// Special key
		switch kb.SelectedSpecial {
		case 1: // Backspace
			if len(kb.TextBuffer) > 0 {
				runeText := []rune(kb.TextBuffer)
				kb.TextBuffer = string(runeText[:len(runeText)-1])
			}
		case 2: // Enter
			kb.TextBuffer += "\n"
		case 3: // Space
			kb.TextBuffer += " "
		case 4: // Shift
			kb.ShiftPressed = !kb.ShiftPressed
			if kb.ShiftPressed {
				kb.CurrentState = UpperCase
			} else {
				kb.CurrentState = LowerCase
			}
		}
	}
}

// ProcessKeyboardInput processes physical keyboard input
func (kb *VirtualKeyboard) ProcessKeyboardInput(keyCode sdl.Keycode) {
	// Handle direct typing/input
	switch keyCode {
	case sdl.K_BACKSPACE:
		if len(kb.TextBuffer) > 0 {
			runeText := []rune(kb.TextBuffer)
			kb.TextBuffer = string(runeText[:len(runeText)-1])
		}
		return

	case sdl.K_RETURN:
		// If something is selected, process the selection
		if kb.SelectedKeyIndex >= 0 || kb.SelectedSpecial > 0 {
			kb.ProcessSelection()
		} else {
			// Otherwise, add a newline
			kb.TextBuffer += "\n"
		}
		return

	case sdl.K_SPACE:
		// If in navigation mode, select the current key
		if kb.SelectedKeyIndex >= 0 || kb.SelectedSpecial > 0 {
			kb.ProcessSelection()
		} else {
			// Otherwise, add a space
			kb.TextBuffer += " "
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

	// Add arrow key navigation
	case sdl.K_UP:
		kb.ProcessNavigation(3) // Up direction
		return
	case sdl.K_DOWN:
		kb.ProcessNavigation(4) // Down direction
		return
	case sdl.K_LEFT:
		kb.ProcessNavigation(2) // Left direction
		return
	case sdl.K_RIGHT:
		kb.ProcessNavigation(1) // Right direction
		return
	}

	// Process alphanumeric keys
	var r rune
	switch {
	case keyCode >= sdl.K_a && keyCode <= sdl.K_z:
		if kb.CurrentState == UpperCase {
			r = rune(keyCode - sdl.K_a + 'A')
		} else {
			r = rune(keyCode - sdl.K_a + 'a')
		}
		kb.TextBuffer += string(r)

	case keyCode >= sdl.K_0 && keyCode <= sdl.K_9:
		if kb.ShiftPressed {
			// Handle symbol input from 0-9 keys when shift is pressed
			symbols := []string{")", "!", "@", "#", "$", "%", "^", "&", "*", "("}
			idx := int(keyCode - sdl.K_0)
			if idx == 0 {
				idx = 9 // Handle the '0' key specially
			} else {
				idx = idx - 1
			}
			kb.TextBuffer += symbols[idx]
		} else {
			r = rune(keyCode - sdl.K_0 + '0')
			kb.TextBuffer += string(r)
		}
	}
}

// Render draws the keyboard on the renderer
func (kb *VirtualKeyboard) Render(renderer *sdl.Renderer, font *ttf.Font) {
	// Draw text input box with slightly darker color
	renderer.SetDrawColor(40, 40, 40, 255)
	renderer.FillRect(&kb.TextInputRect)
	renderer.SetDrawColor(80, 80, 80, 255)
	renderer.DrawRect(&kb.TextInputRect)

	// Draw text in input box
	if kb.TextBuffer != "" {
		textColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
		textSurface, _ := font.RenderUTF8BlendedWrapped(kb.TextBuffer, textColor, int(kb.TextInputRect.W-20))
		if textSurface != nil {
			textTexture, _ := renderer.CreateTextureFromSurface(textSurface)
			textRect := sdl.Rect{
				X: kb.TextInputRect.X + 10,
				Y: kb.TextInputRect.Y + 10,
				W: textSurface.W,
				H: textSurface.H,
			}
			renderer.Copy(textTexture, nil, &textRect)
			textSurface.Free()
			textTexture.Destroy()
		}
	}

	// Draw all keys
	for i, key := range kb.Keys {
		// Key background
		if key.IsPressed || i == kb.SelectedKeyIndex {
			renderer.SetDrawColor(100, 100, 200, 255)
		} else {
			renderer.SetDrawColor(60, 60, 60, 255)
		}
		renderer.FillRect(&key.Rect)

		// Key border
		renderer.SetDrawColor(120, 120, 120, 255)
		renderer.DrawRect(&key.Rect)

		// Key text
		var keyText string

		// For number keys (first row), show symbol when shift is pressed
		if i < 10 && kb.ShiftPressed {
			keyText = key.SymbolValue
		} else if kb.CurrentState == UpperCase {
			keyText = key.UpperValue
		} else {
			keyText = key.LowerValue
		}

		textColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
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

	// Draw special keys with highlight if selected

	// Backspace
	if kb.SelectedSpecial == 1 {
		renderer.SetDrawColor(100, 100, 200, 255)
	} else {
		renderer.SetDrawColor(60, 60, 60, 255)
	}
	renderer.FillRect(&kb.BackspaceRect)
	renderer.SetDrawColor(120, 120, 120, 255)
	renderer.DrawRect(&kb.BackspaceRect)

	textColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
	textSurface, _ := font.RenderUTF8Blended("␡", textColor)
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

	// Enter
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

	// Space
	if kb.SelectedSpecial == 3 {
		renderer.SetDrawColor(100, 100, 200, 255)
	} else {
		renderer.SetDrawColor(60, 60, 60, 255)
	}
	renderer.FillRect(&kb.SpaceRect)
	renderer.SetDrawColor(120, 120, 120, 255)
	renderer.DrawRect(&kb.SpaceRect)

	// Shift
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
}
