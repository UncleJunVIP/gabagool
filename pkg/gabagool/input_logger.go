package gabagool

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type inputLoggerController struct {
	lastInput string
	font      *ttf.Font
	textColor sdl.Color
}

func newInputLogger() *inputLoggerController {
	return &inputLoggerController{
		lastInput: "",
		font:      fonts.largeFont,
		textColor: sdl.Color{R: 200, G: 100, B: 255, A: 255},
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
		}
	case *sdl.JoyButtonEvent:
		if e.State == sdl.PRESSED {
			il.lastInput = fmt.Sprintf("Controller: %d", int(e.Button))
		}
	case *sdl.ControllerButtonEvent:
		if e.State == sdl.PRESSED {
			il.lastInput = fmt.Sprintf("Gamepad: %d", int(e.Button))
		}
	case *sdl.JoyAxisEvent:
		// Only log significant axis movements
		if abs(int(e.Value)) > 16000 {
			il.lastInput = fmt.Sprintf("Axis: %d", int(e.Axis))
		}
	case *sdl.JoyHatEvent:
		if e.Value != sdl.HAT_CENTERED {
			il.lastInput = fmt.Sprintf("Hat: %d", int(e.Value))
		}
	}
	return true
}

func (il *inputLoggerController) render() {
	renderer := GetWindow().Renderer

	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.Clear()

	title := "Gabagool Input Tester"
	il.renderText(renderer, title, GetWindow().Width/2, 50, true)

	if il.lastInput == "" {
		instructionText := "Press any button or key"
		il.renderText(renderer, instructionText, GetWindow().Width/2, GetWindow().Height/2, true)
	} else {
		il.renderText(renderer, il.lastInput, GetWindow().Width/2, GetWindow().Height/2, true)
	}

	renderer.Present()
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
