package gabagool

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type inputLoggerController struct {
	lastInput      string
	lastButtonName string
	font           *ttf.Font
	textColor      sdl.Color
}

func newInputLogger() *inputLoggerController {
	return &inputLoggerController{
		lastInput:      "",
		lastButtonName: "",
		font:           fonts.largeFont,
		textColor:      sdl.Color{R: 200, G: 100, B: 255, A: 255},
	}
}

func InputLogger() {
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
}

func (il *inputLoggerController) handleEvent(event sdl.Event) bool {
	switch e := event.(type) {
	case *sdl.KeyboardEvent:
		if e.State == sdl.PRESSED {
			// Exit on ESC key
			if e.Keysym.Scancode == sdl.SCANCODE_ESCAPE {
				return false
			}
			il.lastInput = fmt.Sprintf("Keyboard: %d", int(e.Keysym.Scancode))
			il.lastButtonName = ""
		}
	case *sdl.JoyButtonEvent:
		// This handles raw joystick button events (like RG35XXSP muOS-Keys)
		if e.State == sdl.PRESSED {
			il.lastInput = fmt.Sprintf("Joystick Button: %d", int(e.Button))
			il.lastButtonName = getButtonNameFromCode(e.Button)
			GetLoggerInstance().Debug("Raw joystick button pressed", "button", e.Button, "mapped", il.lastButtonName)
		}
	case *sdl.ControllerButtonEvent:
		// This handles standardized game controller button events
		if e.State == sdl.PRESSED {
			il.lastInput = fmt.Sprintf("Game Controller: %d", int(e.Button))
			il.lastButtonName = getButtonNameFromCode(e.Button)
		}
	case *sdl.JoyAxisEvent:
		// Only log significant axis movements
		if abs(int(e.Value)) > 16000 {
			il.lastInput = fmt.Sprintf("Joystick Axis: %d", int(e.Axis))
			il.lastButtonName = ""
			GetLoggerInstance().Debug("Joystick axis motion", "axis", e.Axis, "value", e.Value)
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
			il.lastButtonName = ""
			GetLoggerInstance().Debug("Joystick hat motion", "hat", e.Hat, "value", hatName)
		}
	}
	return true
}

func (il *inputLoggerController) render() {
	renderer := GetWindow().Renderer

	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.Clear()

	title := "Gabagool Input Tester"
	il.renderText(renderer, title, GetWindow().GetWidth()/2, 50, true)

	if il.lastInput == "" {
		instructionText := "Press any button or key"
		il.renderText(renderer, instructionText, GetWindow().GetWidth()/2, GetWindow().GetHeight()/2, true)
	} else {
		il.renderText(renderer, il.lastInput, GetWindow().GetWidth()/2, GetWindow().GetWidth()/2-40, true)
		if il.lastButtonName != "" {
			il.renderText(renderer, il.lastButtonName, GetWindow().GetWidth()/2, GetWindow().GetHeight()/2+40, true)
		}
	}

	renderer.Present()
}

func getButtonNameFromCode(code uint8) string {
	// Check against current button mappings
	mapping := GetCurrentButtonMapping()
	for buttonName, buttonCode := range mapping {
		if buttonCode == Button(code) {
			return buttonName
		}
	}
	return "Unknown"
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
