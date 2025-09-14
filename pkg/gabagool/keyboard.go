package gabagool

import (
	"time"

	"github.com/patrickhuber/go-types"
	"github.com/patrickhuber/go-types/option"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
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
	symbolsMode
)

type virtualKeyboard struct {
	Keys             []key
	TextBuffer       string
	CurrentState     keyboardState
	ShiftPressed     bool
	SymbolPressed    bool
	BackspaceRect    sdl.Rect
	EnterRect        sdl.Rect
	SpaceRect        sdl.Rect
	ShiftRect        sdl.Rect
	SymbolRect       sdl.Rect
	TextInputRect    sdl.Rect
	KeyboardRect     sdl.Rect
	SelectedKeyIndex int
	SelectedSpecial  int
	CursorPosition   int
	CursorVisible    bool
	LastCursorBlink  time.Time
	CursorBlinkRate  time.Duration
	helpOverlay      *helpOverlay
	ShowingHelp      bool
	EnterPressed     bool
}

var defaultKeyboardHelpLines = []string{
	"• D-Pad: Navigate between keys",
	"• A: Type the selected key",
	"• B: backspace",
	"• X: Space",
	"• L1 / R1: Move cursor within text",
	"• Select: toggle Shift (uppercase/symbols)",
	"• L2: toggle Symbols mode",
	"• Y: Exit keyboard without saving",
	"• Start: Enter (confirm input)",
}

type keyLayout struct {
	rows [][]interface{}
}

func createKeyLayout() *keyLayout {
	return &keyLayout{
		rows: [][]interface{}{
			// Row 1: numbers + backspace
			{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, "backspace"},
			// Row 2: qwerty row
			{10, 11, 12, 13, 14, 15, 16, 17, 18, 19},
			// Row 3: asdf row + enter
			{20, 21, 22, 23, 24, 25, 26, 27, 28, "enter"},
			// Row 4: shift + zxcv row + symbol
			{"shift", 29, 30, 31, 32, 33, 34, 35, "symbol"},
			// Row 5: space only
			{"space"},
		},
	}
}

func createKeyboard(windowWidth, windowHeight int32) *virtualKeyboard {
	kb := &virtualKeyboard{
		Keys:             createKeys(),
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

	kb.helpOverlay = newHelpOverlay("Keyboard Help", defaultKeyboardHelpLines)
	setupKeyboardRects(kb, windowWidth, windowHeight)

	return kb
}

func createKeys() []key {
	keys := make([]key, 36) // Total number of regular keys

	// Numbers row
	numbers := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0"}
	numberSymbols := []string{"!", "@", "#", "$", "%", "^", "&", "*", "(", ")"}

	for i, num := range numbers {
		keys[i] = key{
			LowerValue:  num,
			UpperValue:  num,
			SymbolValue: numberSymbols[i],
		}
	}

	// QWERTY row
	qwerty := "qwertyuiop"
	qwertySymbols := []string{"`", "~", "[", "]", "\\", "|", "{", "}", ";", ":"}
	for i, char := range qwerty {
		keys[10+i] = key{
			LowerValue:  string(char),
			UpperValue:  string(char - 32),
			SymbolValue: qwertySymbols[i],
		}
	}

	// ASDF row
	asdf := "asdfghjkl"
	asdfSymbols := []string{"'", "\"", "<", ">", "?", "/", "+", "=", "_"}
	for i, char := range asdf {
		keys[20+i] = key{
			LowerValue:  string(char),
			UpperValue:  string(char - 32),
			SymbolValue: asdfSymbols[i],
		}
	}

	// ZXCV row - avoiding symbols already used
	zxcv := "zxcvbnm"
	zxcvSymbols := []string{",", ".", "-", "€", "£", "¥", "¢"}
	for i, char := range zxcv {
		keys[29+i] = key{
			LowerValue:  string(char),
			UpperValue:  string(char - 32),
			SymbolValue: zxcvSymbols[i],
		}
	}

	return keys
}

func setupKeyboardRects(kb *virtualKeyboard, windowWidth, windowHeight int32) {
	keyboardWidth := (windowWidth * 85) / 100
	keyboardHeight := (windowHeight * 85) / 100
	textInputHeight := windowHeight / 10
	keyboardHeight = keyboardHeight - textInputHeight - 20
	startX := (windowWidth - keyboardWidth) / 2
	textInputY := (windowHeight - keyboardHeight - textInputHeight - 20) / 2
	keyboardStartY := textInputY + textInputHeight + 20

	kb.KeyboardRect = sdl.Rect{X: startX, Y: keyboardStartY, W: keyboardWidth, H: keyboardHeight}
	kb.TextInputRect = sdl.Rect{X: startX, Y: textInputY, W: keyboardWidth, H: textInputHeight}

	keyWidth := keyboardWidth / 12
	keyHeight := keyboardHeight / 6
	keySpacing := int32(3)

	// Define consistent key widths for special keys
	backspaceWidth := keyWidth * 2
	shiftWidth := keyWidth * 2
	symbolWidth := keyWidth * 2
	enterWidth := keyWidth + keyWidth/2
	spaceWidth := keyWidth * 8

	// Calculate the maximum row width to determine consistent left margin
	// Row 1: 10 regular keys + backspace
	row1Width := keyWidth*10 + keySpacing*9 + backspaceWidth + keySpacing
	// Row 2: 10 regular keys
	row2Width := keyWidth*10 + keySpacing*9
	// Row 3: 9 regular keys + enter
	row3Width := keyWidth*9 + keySpacing*8 + enterWidth + keySpacing
	// Row 4: shift + 7 regular keys + symbol
	row4Width := shiftWidth + keySpacing + keyWidth*7 + keySpacing*6 + symbolWidth + keySpacing
	// Row 5: space
	row5Width := spaceWidth

	// Find the maximum width to align all rows consistently
	maxRowWidth := row1Width
	if row2Width > maxRowWidth {
		maxRowWidth = row2Width
	}
	if row3Width > maxRowWidth {
		maxRowWidth = row3Width
	}
	if row4Width > maxRowWidth {
		maxRowWidth = row4Width
	}

	// Calculate consistent left margin for all rows
	leftMargin := startX + (keyboardWidth-maxRowWidth)/2

	y := keyboardStartY + keySpacing

	// Row 1: Numbers + Backspace
	x := leftMargin
	for i := 0; i < 10; i++ {
		kb.Keys[i].Rect = sdl.Rect{X: x, Y: y, W: keyWidth, H: keyHeight}
		x += keyWidth + keySpacing
	}
	kb.BackspaceRect = sdl.Rect{X: x, Y: y, W: backspaceWidth, H: keyHeight}

	// Row 2: QWERTY
	y += keyHeight + keySpacing
	x = leftMargin + (maxRowWidth-row2Width)/2 // Center this row within max width
	for i := 10; i < 20; i++ {
		kb.Keys[i].Rect = sdl.Rect{X: x, Y: y, W: keyWidth, H: keyHeight}
		x += keyWidth + keySpacing
	}

	// Row 3: ASDF + Enter
	y += keyHeight + keySpacing
	x = leftMargin + (maxRowWidth-row3Width)/2 // Center this row within max width
	for i := 20; i < 29; i++ {
		kb.Keys[i].Rect = sdl.Rect{X: x, Y: y, W: keyWidth, H: keyHeight}
		x += keyWidth + keySpacing
	}
	kb.EnterRect = sdl.Rect{X: x, Y: y, W: enterWidth, H: keyHeight}

	// Row 4: Shift + ZXCV + Symbol
	y += keyHeight + keySpacing
	x = leftMargin + (maxRowWidth-row4Width)/2 // Center this row within max width

	kb.ShiftRect = sdl.Rect{X: x, Y: y, W: shiftWidth, H: keyHeight}
	x += shiftWidth + keySpacing

	for i := 29; i < 36; i++ {
		kb.Keys[i].Rect = sdl.Rect{X: x, Y: y, W: keyWidth, H: keyHeight}
		x += keyWidth + keySpacing
	}

	kb.SymbolRect = sdl.Rect{X: x, Y: y, W: symbolWidth, H: keyHeight}

	// Row 5: Space
	y += keyHeight + keySpacing
	x = leftMargin + (maxRowWidth-row5Width)/2 // Center space bar within max width
	kb.SpaceRect = sdl.Rect{X: x, Y: y, W: spaceWidth, H: keyHeight}
}

func Keyboard(initialText string) (types.Option[string], error) {
	window := GetWindow()
	renderer := window.Renderer
	font := fonts.mediumFont

	kb := createKeyboard(window.Width, window.Height)
	if initialText != "" {
		kb.TextBuffer = initialText
		kb.CursorPosition = len(initialText)
	}

	for {
		if kb.handleEvents() {
			break
		}

		kb.updateCursorBlink()
		kb.render(renderer, font)
		sdl.Delay(16)
	}

	if kb.EnterPressed {
		return option.Some(kb.TextBuffer), nil
	}
	return option.None[string](), nil
}

func (kb *virtualKeyboard) handleEvents() bool {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.QuitEvent:
			return true

		case *sdl.KeyboardEvent:
			if e.Type == sdl.KEYDOWN {
				if kb.handleKeyInput(e.Keysym.Sym) {
					return true
				}
			}

		case *sdl.ControllerButtonEvent:
			if e.Type == sdl.CONTROLLERBUTTONDOWN {
				if kb.handleButtonInput(Button(e.Button)) {
					return true
				}
			}
		}
	}
	return false
}

func (kb *virtualKeyboard) handleKeyInput(key sdl.Keycode) bool {
	if key == sdl.K_h {
		kb.toggleHelp()
		return false
	}

	if kb.ShowingHelp {
		return kb.handleHelpInput(key)
	}

	switch key {
	case sdl.K_UP, sdl.K_DOWN, sdl.K_LEFT, sdl.K_RIGHT:
		kb.navigate(key)
	case sdl.K_a:
		kb.processSelection()
	case sdl.K_b:
		kb.backspace()
	case sdl.K_LSHIFT, sdl.K_RSHIFT:
		kb.toggleShift()
	case sdl.K_q: // Using Q key as symbol toggle for keyboard input
		kb.toggleSymbols()
	case sdl.K_RETURN:
		kb.EnterPressed = true
		return true
	case sdl.K_ESCAPE:
		return true
	}
	return false
}

func (kb *virtualKeyboard) handleButtonInput(button Button) bool {
	if button == ButtonMenu {
		kb.toggleHelp()
		return false
	}

	if kb.ShowingHelp {
		return kb.handleHelpButton(button)
	}

	switch button {
	case ButtonUp, ButtonDown, ButtonLeft, ButtonRight:
		kb.navigateWithButton(button)
	case ButtonA:
		kb.processSelection()
	case ButtonB:
		kb.backspace()
	case ButtonX:
		kb.insertSpace()
	case ButtonSelect:
		kb.toggleShift()
	case ButtonL2:
		kb.toggleSymbols()
	case ButtonL1:
		kb.moveCursor(-1)
	case ButtonR1:
		kb.moveCursor(1)
	case ButtonStart:
		kb.EnterPressed = true
		return true
	case ButtonY:
		return true
	}
	return false
}

func (kb *virtualKeyboard) handleHelpInput(key sdl.Keycode) bool {
	switch key {
	case sdl.K_UP:
		kb.scrollHelpOverlay(-1)
	case sdl.K_DOWN:
		kb.scrollHelpOverlay(1)
	default:
		kb.ShowingHelp = false
	}
	return false
}

func (kb *virtualKeyboard) handleHelpButton(button Button) bool {
	switch button {
	case ButtonUp:
		kb.scrollHelpOverlay(-1)
	case ButtonDown:
		kb.scrollHelpOverlay(1)
	default:
		kb.ShowingHelp = false
	}
	return false
}

func (kb *virtualKeyboard) navigate(key sdl.Keycode) {
	layout := createKeyLayout()
	currentRow, currentCol := kb.findCurrentPosition(layout)

	var newRow, newCol int
	switch key {
	case sdl.K_UP:
		newRow, newCol = kb.moveUp(layout, currentRow, currentCol)
	case sdl.K_DOWN:
		newRow, newCol = kb.moveDown(layout, currentRow, currentCol)
	case sdl.K_LEFT:
		newRow, newCol = kb.moveLeft(layout, currentRow, currentCol)
	case sdl.K_RIGHT:
		newRow, newCol = kb.moveRight(layout, currentRow, currentCol)
	}

	kb.setSelection(layout, newRow, newCol)
}

func (kb *virtualKeyboard) navigateWithButton(button Button) {
	layout := createKeyLayout()
	currentRow, currentCol := kb.findCurrentPosition(layout)

	var newRow, newCol int
	switch button {
	case ButtonUp:
		newRow, newCol = kb.moveUp(layout, currentRow, currentCol)
	case ButtonDown:
		newRow, newCol = kb.moveDown(layout, currentRow, currentCol)
	case ButtonLeft:
		newRow, newCol = kb.moveLeft(layout, currentRow, currentCol)
	case ButtonRight:
		newRow, newCol = kb.moveRight(layout, currentRow, currentCol)
	}

	kb.setSelection(layout, newRow, newCol)
}

func (kb *virtualKeyboard) findCurrentPosition(layout *keyLayout) (int, int) {
	specialKeys := map[int]string{1: "backspace", 2: "enter", 3: "space", 4: "shift", 5: "symbol"}

	if kb.SelectedSpecial > 0 {
		targetKey := specialKeys[kb.SelectedSpecial]
		for r, row := range layout.rows {
			for c, key := range row {
				if str, ok := key.(string); ok && str == targetKey {
					return r, c
				}
			}
		}
	}

	for r, row := range layout.rows {
		for c, key := range row {
			if idx, ok := key.(int); ok && idx == kb.SelectedKeyIndex {
				return r, c
			}
		}
	}

	return 0, 0 // Default position
}

func (kb *virtualKeyboard) moveUp(layout *keyLayout, row, col int) (int, int) {
	newRow := row - 1
	if newRow < 0 {
		newRow = len(layout.rows) - 1
	}
	if col >= len(layout.rows[newRow]) {
		col = len(layout.rows[newRow]) - 1
	}
	return newRow, col
}

func (kb *virtualKeyboard) moveDown(layout *keyLayout, row, col int) (int, int) {
	newRow := row + 1
	if newRow >= len(layout.rows) {
		newRow = 0
	}
	if col >= len(layout.rows[newRow]) {
		col = len(layout.rows[newRow]) - 1
	}
	return newRow, col
}

func (kb *virtualKeyboard) moveLeft(layout *keyLayout, row, col int) (int, int) {
	newCol := col - 1
	if newCol < 0 {
		newCol = len(layout.rows[row]) - 1
	}
	return row, newCol
}

func (kb *virtualKeyboard) moveRight(layout *keyLayout, row, col int) (int, int) {
	newCol := col + 1
	if newCol >= len(layout.rows[row]) {
		newCol = 0
	}
	return row, newCol
}

func (kb *virtualKeyboard) setSelection(layout *keyLayout, row, col int) {
	kb.resetPressedKeys()

	selectedKey := layout.rows[row][col]
	if idx, ok := selectedKey.(int); ok {
		kb.SelectedKeyIndex = idx
		kb.SelectedSpecial = 0
		kb.Keys[kb.SelectedKeyIndex].IsPressed = true
	} else if str, ok := selectedKey.(string); ok {
		kb.SelectedKeyIndex = -1
		specialMap := map[string]int{"backspace": 1, "enter": 2, "space": 3, "shift": 4, "symbol": 5}
		kb.SelectedSpecial = specialMap[str]
	}
}

func (kb *virtualKeyboard) processSelection() {
	if kb.SelectedKeyIndex >= 0 && kb.SelectedKeyIndex < len(kb.Keys) {
		keyValue := kb.getKeyValue(kb.SelectedKeyIndex)
		kb.insertText(keyValue)
	} else {
		kb.handleSpecialKey()
	}

	kb.CursorVisible = true
	kb.LastCursorBlink = time.Now()
}

func (kb *virtualKeyboard) getKeyValue(index int) string {
	key := kb.Keys[index]
	if kb.CurrentState == symbolsMode {
		return key.SymbolValue
	} else if index < 10 && kb.ShiftPressed {
		return key.SymbolValue
	} else if kb.CurrentState == upperCase {
		return key.UpperValue
	}
	return key.LowerValue
}

func (kb *virtualKeyboard) insertText(text string) {
	if kb.CursorPosition == len(kb.TextBuffer) {
		kb.TextBuffer += text
	} else {
		textRunes := []rune(kb.TextBuffer)
		before := string(textRunes[:kb.CursorPosition])
		after := string(textRunes[kb.CursorPosition:])
		kb.TextBuffer = before + text + after
	}
	kb.CursorPosition += len([]rune(text))
}

func (kb *virtualKeyboard) handleSpecialKey() {
	switch kb.SelectedSpecial {
	case 1: // backspace
		kb.backspace()
	case 2: // enter
		kb.EnterPressed = true
	case 3: // space
		kb.insertSpace()
	case 4: // shift
		kb.toggleShift()
	case 5: // symbol
		kb.toggleSymbols()
	}
}

func (kb *virtualKeyboard) backspace() {
	if kb.CursorPosition > 0 {
		textRunes := []rune(kb.TextBuffer)
		before := string(textRunes[:kb.CursorPosition-1])
		after := string(textRunes[kb.CursorPosition:])
		kb.TextBuffer = before + after
		kb.CursorPosition--
	}
}

func (kb *virtualKeyboard) insertSpace() {
	kb.insertText(" ")
}

func (kb *virtualKeyboard) toggleShift() {
	if kb.CurrentState == symbolsMode {
		// If in symbols mode, shift just toggles the shift flag
		kb.ShiftPressed = !kb.ShiftPressed
	} else {
		// Normal shift behavior for upper/lower case
		kb.ShiftPressed = !kb.ShiftPressed
		if kb.ShiftPressed {
			kb.CurrentState = upperCase
		} else {
			kb.CurrentState = lowerCase
		}
	}
}

func (kb *virtualKeyboard) toggleSymbols() {
	kb.SymbolPressed = !kb.SymbolPressed
	if kb.SymbolPressed {
		kb.CurrentState = symbolsMode
	} else {
		if kb.ShiftPressed {
			kb.CurrentState = upperCase
		} else {
			kb.CurrentState = lowerCase
		}
	}
}

func (kb *virtualKeyboard) moveCursor(direction int) {
	if direction > 0 && kb.CursorPosition < len(kb.TextBuffer) {
		kb.CursorPosition++
	} else if direction < 0 && kb.CursorPosition > 0 {
		kb.CursorPosition--
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

func (kb *virtualKeyboard) toggleHelp() {
	if kb.helpOverlay == nil {
		kb.helpOverlay = newHelpOverlay("Keyboard Help", defaultKeyboardHelpLines)
	}
	kb.helpOverlay.toggle()
	kb.ShowingHelp = kb.helpOverlay.ShowingHelp
}

func (kb *virtualKeyboard) scrollHelpOverlay(direction int) {
	if kb.helpOverlay != nil {
		kb.helpOverlay.scroll(direction)
	}
}

func (kb *virtualKeyboard) render(renderer *sdl.Renderer, font *ttf.Font) {
	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.Clear()

	window := GetWindow()
	window.RenderBackground()

	if !kb.ShowingHelp {
		kb.renderTextInput(renderer, font)
		kb.renderKeys(renderer, font)
		kb.renderSpecialKeys(renderer)
		kb.renderFooter(renderer)
	}

	if kb.ShowingHelp && kb.helpOverlay != nil {
		kb.helpOverlay.render(renderer, fonts.smallFont)
	}

	renderer.Present()
}

func (kb *virtualKeyboard) renderTextInput(renderer *sdl.Renderer, font *ttf.Font) {
	// Background
	renderer.SetDrawColor(50, 50, 50, 255)
	renderer.FillRect(&kb.TextInputRect)
	renderer.SetDrawColor(200, 200, 200, 255)
	renderer.DrawRect(&kb.TextInputRect)

	padding := int32(10)
	if kb.TextBuffer != "" {
		kb.renderTextWithCursor(renderer, font, padding)
	} else if kb.CursorVisible {
		kb.renderEmptyCursor(renderer, font, padding)
	}
}

func (kb *virtualKeyboard) renderTextWithCursor(renderer *sdl.Renderer, font *ttf.Font, padding int32) {
	textColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
	textSurface, err := font.RenderUTF8Blended(kb.TextBuffer, textColor)
	if err != nil {
		return
	}
	defer textSurface.Free()

	textTexture, err := renderer.CreateTextureFromSurface(textSurface)
	if err != nil {
		return
	}
	defer textTexture.Destroy()

	// Calculate cursor position and scrolling
	cursorX := kb.calculateCursorX(font)
	visibleWidth := kb.TextInputRect.W - (padding * 2)
	offsetX := kb.calculateScrollOffset(cursorX, visibleWidth, textSurface.W, padding)

	// Render text
	srcRect := &sdl.Rect{X: offsetX, Y: 0, W: visibleWidth, H: textSurface.H}
	if textSurface.W < visibleWidth {
		srcRect.W = textSurface.W
	}

	textRect := sdl.Rect{
		X: kb.TextInputRect.X + padding,
		Y: kb.TextInputRect.Y + (kb.TextInputRect.H-textSurface.H)/2,
		W: srcRect.W,
		H: textSurface.H,
	}
	renderer.Copy(textTexture, srcRect, &textRect)

	// Render cursor
	if kb.CursorVisible {
		cursorRect := sdl.Rect{
			X: kb.TextInputRect.X + padding + cursorX - offsetX,
			Y: textRect.Y,
			W: 2,
			H: textSurface.H,
		}
		if cursorRect.X >= kb.TextInputRect.X+padding && cursorRect.X <= kb.TextInputRect.X+padding+visibleWidth {
			renderer.SetDrawColor(255, 255, 255, 255)
			renderer.FillRect(&cursorRect)
		}
	}
}

func (kb *virtualKeyboard) renderEmptyCursor(renderer *sdl.Renderer, font *ttf.Font, padding int32) {
	// Get font height for consistent cursor size
	fontHeight := font.Height()

	cursorRect := sdl.Rect{
		X: kb.TextInputRect.X + padding,
		Y: kb.TextInputRect.Y + (kb.TextInputRect.H - int32(fontHeight)),
		W: 2,
		H: int32(fontHeight),
	}
	renderer.SetDrawColor(255, 255, 255, 255)
	renderer.FillRect(&cursorRect)
}

func (kb *virtualKeyboard) calculateCursorX(font *ttf.Font) int32 {
	if kb.CursorPosition == 0 {
		return 0
	}

	cursorText := kb.TextBuffer[:kb.CursorPosition]
	textColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
	cursorSurface, err := font.RenderUTF8Blended(cursorText, textColor)
	if err != nil {
		return 0
	}
	defer cursorSurface.Free()

	return cursorSurface.W
}

func (kb *virtualKeyboard) calculateScrollOffset(cursorX, visibleWidth, textWidth, padding int32) int32 {
	offsetX := int32(0)
	if cursorX > visibleWidth {
		offsetX = cursorX - visibleWidth + padding
	}

	maxOffset := textWidth - visibleWidth
	if maxOffset < 0 {
		maxOffset = 0
	}
	if offsetX > maxOffset {
		offsetX = maxOffset
	}

	return offsetX
}

func (kb *virtualKeyboard) renderKeys(renderer *sdl.Renderer, font *ttf.Font) {
	for i, key := range kb.Keys {
		kb.renderSingleKey(renderer, font, i, key)
	}
}

func (kb *virtualKeyboard) renderSingleKey(renderer *sdl.Renderer, font *ttf.Font, index int, key key) {
	// Background
	bgColor := sdl.Color{R: 50, G: 50, B: 60, A: 255}
	if index == kb.SelectedKeyIndex {
		bgColor = sdl.Color{R: 100, G: 100, B: 240, A: 255}
	} else if key.IsPressed {
		bgColor = sdl.Color{R: 80, G: 80, B: 120, A: 255}
	}

	renderer.SetDrawColor(bgColor.R, bgColor.G, bgColor.B, bgColor.A)
	renderer.FillRect(&key.Rect)
	renderer.SetDrawColor(70, 70, 80, 255)
	renderer.DrawRect(&key.Rect)

	// Text
	keyValue := kb.getKeyValue(index)
	kb.renderKeyText(renderer, font, keyValue, key.Rect)
}

func (kb *virtualKeyboard) renderKeyText(renderer *sdl.Renderer, font *ttf.Font, text string, rect sdl.Rect) {
	textColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
	textSurface, err := font.RenderUTF8Blended(text, textColor)
	if err != nil {
		return
	}
	defer textSurface.Free()

	textTexture, err := renderer.CreateTextureFromSurface(textSurface)
	if err != nil {
		return
	}
	defer textTexture.Destroy()

	textRect := sdl.Rect{
		X: rect.X + (rect.W-textSurface.W)/2,
		Y: rect.Y + (rect.H-textSurface.H)/2,
		W: textSurface.W,
		H: textSurface.H,
	}
	renderer.Copy(textTexture, nil, &textRect)
}

func (kb *virtualKeyboard) renderSpecialKeys(renderer *sdl.Renderer) {
	kb.renderSpecialKey(renderer, kb.BackspaceRect, "←", kb.SelectedSpecial == 1)
	kb.renderSpecialKey(renderer, kb.EnterRect, "↵", kb.SelectedSpecial == 2)
	kb.renderSpecialKey(renderer, kb.ShiftRect, "⇧", kb.SelectedSpecial == 4 || kb.CurrentState == upperCase)
	kb.renderSpecialKey(renderer, kb.SymbolRect, "sym", kb.SelectedSpecial == 5 || kb.CurrentState == symbolsMode)
	kb.renderSpaceKey(renderer)
}

func (kb *virtualKeyboard) renderSpecialKey(renderer *sdl.Renderer, rect sdl.Rect, symbol string, isSelected bool) {
	bgColor := sdl.Color{R: 50, G: 50, B: 60, A: 255}
	if isSelected {
		bgColor = sdl.Color{R: 100, G: 100, B: 240, A: 255}
	}

	renderer.SetDrawColor(bgColor.R, bgColor.G, bgColor.B, bgColor.A)
	renderer.FillRect(&rect)
	renderer.SetDrawColor(70, 70, 80, 255)
	renderer.DrawRect(&rect)

	textColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}
	textSurface, err := fonts.largeSymbolFont.RenderUTF8Blended(symbol, textColor)
	if err != nil {
		return
	}
	defer textSurface.Free()

	textTexture, err := renderer.CreateTextureFromSurface(textSurface)
	if err != nil {
		return
	}
	defer textTexture.Destroy()

	textRect := sdl.Rect{
		X: rect.X + (rect.W-textSurface.W)/2,
		Y: rect.Y + (rect.H-textSurface.H)/2,
		W: textSurface.W,
		H: textSurface.H,
	}
	renderer.Copy(textTexture, nil, &textRect)
}

func (kb *virtualKeyboard) renderSpaceKey(renderer *sdl.Renderer) {
	bgColor := sdl.Color{R: 50, G: 50, B: 60, A: 255}
	if kb.SelectedSpecial == 3 {
		bgColor = sdl.Color{R: 100, G: 100, B: 240, A: 255}
	}

	renderer.SetDrawColor(bgColor.R, bgColor.G, bgColor.B, bgColor.A)
	renderer.FillRect(&kb.SpaceRect)
	renderer.SetDrawColor(70, 70, 80, 255)
	renderer.DrawRect(&kb.SpaceRect)

	// Draw space bar indicator
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
}

func (kb *virtualKeyboard) renderFooter(renderer *sdl.Renderer) {
	renderFooter(
		renderer,
		fonts.smallFont,
		[]FooterHelpItem{
			{ButtonName: "Menu", HelpText: "Help"},
		},
		20,
		true,
	)
}
