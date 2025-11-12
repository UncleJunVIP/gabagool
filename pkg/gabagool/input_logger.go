package gabagool

import (
	"fmt"
	"sync"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type buttonConfig struct {
	internalButton InternalButton
	displayName    string
}

type inputLoggerController struct {
	lastInput         string
	lastButtonName    string
	font              *ttf.Font
	textColor         sdl.Color
	currentButtonIdx  int
	mappedButtons     map[InternalButton]int
	currentSource     InputSource
	mutex             sync.Mutex
	buttonSequence    []buttonConfig
	lastInputTime     time.Time
	debounceDelay     time.Duration
	waitingForRelease bool
}

func newInputLogger() *inputLoggerController {
	return &inputLoggerController{
		lastInput:         "",
		lastButtonName:    "",
		font:              fonts.largeFont,
		textColor:         sdl.Color{R: 200, G: 100, B: 255, A: 255},
		mappedButtons:     make(map[InternalButton]int),
		currentButtonIdx:  0,
		debounceDelay:     500 * time.Millisecond, // Wait 500ms before accepting next input
		waitingForRelease: false,
		buttonSequence: []buttonConfig{
			{InternalButtonA, "A Button"},
			{InternalButtonB, "B Button"},
			{InternalButtonX, "X Button"},
			{InternalButtonY, "Y Button"},
			{InternalButtonUp, "D-Pad Up"},
			{InternalButtonDown, "D-Pad Down"},
			{InternalButtonLeft, "D-Pad Left"},
			{InternalButtonRight, "D-Pad Right"},
			{InternalButtonStart, "Start"},
			{InternalButtonSelect, "Select"},
			{InternalButtonL1, "L1"},
			{InternalButtonL2, "L2"},
			{InternalButtonR1, "R1"},
			{InternalButtonR2, "R2"},
			{InternalButtonMenu, "Menu"},
			//{InternalButtonF1, "F1"},
			//{InternalButtonF2, "F2"},
			//{InternalButtonVolumeUp, "Volume Up"},
			//{InternalButtonVolumeDown, "Volume Down"},
			//{InternalButtonPower, "Power"},
		},
	}
}

func InputLogger() *InternalInputMapping {
	logger := newInputLogger()

	running := true

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
			default:
				running = logger.handleEvent(event)
			}
		}

		logger.render()
		sdl.Delay(16) // ~60 FPS
	}

	return logger.buildMapping()
}

func (il *inputLoggerController) handleEvent(event sdl.Event) bool {
	il.mutex.Lock()
	defer il.mutex.Unlock()

	// Check if we're done with all buttons
	if il.currentButtonIdx >= len(il.buttonSequence) {
		return false
	}

	// Check debounce - if we're waiting for the button to be released or debounce time hasn't passed, ignore
	if il.waitingForRelease {
		// Wait for button release
		switch e := event.(type) {
		case *sdl.KeyboardEvent:
			if e.State == sdl.RELEASED {
				il.waitingForRelease = false
			}
		case *sdl.JoyButtonEvent:
			if e.State == sdl.RELEASED {
				il.waitingForRelease = false
			}
		case *sdl.ControllerButtonEvent:
			if e.State == sdl.RELEASED {
				il.waitingForRelease = false
			}
		case *sdl.JoyAxisEvent:
			// Axis has to go back to center
			if abs(int(e.Value)) < 5000 {
				il.waitingForRelease = false
			}
		case *sdl.JoyHatEvent:
			if e.Value == sdl.HAT_CENTERED {
				il.waitingForRelease = false
			}
		}
		return true
	}

	currentButton := il.buttonSequence[il.currentButtonIdx]

	switch e := event.(type) {
	case *sdl.KeyboardEvent:
		if e.State == sdl.PRESSED {
			// Exit on ESC key
			if e.Keysym.Scancode == sdl.SCANCODE_ESCAPE {
				return false
			}
			// Store the KEYCODE (Sym), not the scancode
			il.lastInput = fmt.Sprintf("Keyboard: %d", int(e.Keysym.Sym))
			il.lastButtonName = "Registered!"
			il.mappedButtons[currentButton.internalButton] = int(e.Keysym.Sym)
			il.currentSource = InputSourceKeyboard
			il.lastInputTime = time.Now()
			il.waitingForRelease = true
			GetLoggerInstance().Debug("Registered keyboard button",
				"button", currentButton.displayName,
				"keycode", e.Keysym.Sym,
				"scancode", e.Keysym.Scancode)
			il.advanceToNextButton()
		}
	case *sdl.JoyButtonEvent:
		if e.State == sdl.PRESSED {
			il.lastInput = fmt.Sprintf("Joystick Button: %d", int(e.Button))
			il.lastButtonName = "Registered!"
			il.mappedButtons[currentButton.internalButton] = int(e.Button)
			il.currentSource = InputSourceJoystick
			il.lastInputTime = time.Now()
			il.waitingForRelease = true
			il.advanceToNextButton()
		}
	case *sdl.ControllerButtonEvent:
		if e.State == sdl.PRESSED {
			il.lastInput = fmt.Sprintf("Game Controller: %d", int(e.Button))
			il.lastButtonName = "Registered!"
			il.mappedButtons[currentButton.internalButton] = int(e.Button)
			il.currentSource = InputSourceController
			il.lastInputTime = time.Now()
			il.waitingForRelease = true
			il.advanceToNextButton()
		}
	case *sdl.JoyAxisEvent:
		// Only log significant axis movements
		if abs(int(e.Value)) > 16000 {
			if e.Value > 16000 {
				il.lastInput = fmt.Sprintf("Joystick Axis %d (Positive)", int(e.Axis))
			} else {
				il.lastInput = fmt.Sprintf("Joystick Axis %d (Negative)", int(e.Axis))
			}
			il.lastButtonName = "Registered!"
			il.mappedButtons[currentButton.internalButton] = int(e.Axis)
			il.currentSource = InputSourceJoystick
			il.lastInputTime = time.Now()
			il.waitingForRelease = true
			il.advanceToNextButton()
		}
	case *sdl.JoyHatEvent:
		if e.Value != sdl.HAT_CENTERED {
			hatName := ""
			switch e.Value {
			case sdl.HAT_UP:
				hatName = "UP"
			case sdl.HAT_DOWN:
				hatName = "DOWN"
			case sdl.HAT_LEFT:
				hatName = "LEFT"
			case sdl.HAT_RIGHT:
				hatName = "RIGHT"
			}
			il.lastInput = fmt.Sprintf("Hat Switch: %s", hatName)
			il.lastButtonName = "Registered!"
			il.mappedButtons[currentButton.internalButton] = int(e.Hat)
			il.currentSource = InputSourceHatSwitch
			il.lastInputTime = time.Now()
			il.waitingForRelease = true
			il.advanceToNextButton()
		}
	}
	return true
}

func (il *inputLoggerController) advanceToNextButton() {
	il.currentButtonIdx++
	il.lastInput = ""
	il.lastButtonName = ""
}

func (il *inputLoggerController) render() {
	renderer := GetWindow().Renderer

	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.Clear()

	il.mutex.Lock()
	defer il.mutex.Unlock()

	title := "Gabagool Input Configuration"
	il.renderText(renderer, title, GetWindow().GetWidth()/2, 50, true)

	if il.currentButtonIdx >= len(il.buttonSequence) {
		completionText := "Configuration Complete!"
		il.renderText(renderer, completionText, GetWindow().GetWidth()/2, GetWindow().GetHeight()/2-40, true)
		instructionText := "Press any button to exit"
		il.renderText(renderer, instructionText, GetWindow().GetWidth()/2, GetWindow().GetHeight()/2+40, true)
	} else {
		currentButton := il.buttonSequence[il.currentButtonIdx]
		progressText := fmt.Sprintf("Button %d of %d", il.currentButtonIdx+1, len(il.buttonSequence))
		il.renderText(renderer, progressText, GetWindow().GetWidth()/2, 120, true)

		buttonPrompt := fmt.Sprintf("Press: %s", currentButton.displayName)
		il.renderText(renderer, buttonPrompt, GetWindow().GetWidth()/2, GetWindow().GetHeight()/2-60, true)

		if il.lastInput != "" {
			il.renderText(renderer, il.lastInput, GetWindow().GetWidth()/2, GetWindow().GetHeight()/2-20, true)
			if il.waitingForRelease {
				il.renderText(renderer, "Release button...", GetWindow().GetWidth()/2, GetWindow().GetHeight()/2+20, true)
			} else {
				il.renderText(renderer, il.lastButtonName, GetWindow().GetWidth()/2, GetWindow().GetHeight()/2+20, true)
			}
		}

		escapeHint := "Press ESC to cancel"
		il.renderText(renderer, escapeHint, GetWindow().GetWidth()/2, GetWindow().GetHeight()-80, true)
	}

	renderer.Present()
}

func (il *inputLoggerController) buildMapping() *InternalInputMapping {
	il.mutex.Lock()
	defer il.mutex.Unlock()

	mapping := &InternalInputMapping{
		KeyboardMap:         make(map[sdl.Keycode]InternalButton),
		ControllerButtonMap: make(map[sdl.GameControllerButton]InternalButton),
		ControllerHatMap:    make(map[uint8]InternalButton),
		JoystickAxisMap:     make(map[uint8]JoystickAxisMapping),
		JoystickButtonMap:   make(map[uint8]InternalButton),
		JoystickHatMap:      make(map[uint8]InternalButton),
	}

	// Populate the mapping based on the current source and mapped buttons
	for button, code := range il.mappedButtons {
		switch il.currentSource {
		case InputSourceKeyboard:
			mapping.KeyboardMap[sdl.Keycode(code)] = button
		case InputSourceController:
			mapping.ControllerButtonMap[sdl.GameControllerButton(code)] = button
		case InputSourceJoystick:
			mapping.JoystickButtonMap[uint8(code)] = button
		case InputSourceHatSwitch:
			mapping.JoystickHatMap[uint8(code)] = button
		}
	}

	return mapping
}

func (il *inputLoggerController) renderText(renderer *sdl.Renderer, text string, x, y int32, centered bool) {
	surface, err := il.font.RenderUTF8Blended(text, il.textColor)
	if err != nil {
		return
	}
	defer surface.Free()

	texture, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return
	}
	defer texture.Destroy()

	_, _, w, h, err := texture.Query()
	if err != nil {
		return
	}

	finalX := x
	if centered {
		finalX = x - w/2
	}

	dstRect := sdl.Rect{X: finalX, Y: y, W: w, H: h}
	renderer.Copy(texture, nil, &dstRect)
}
