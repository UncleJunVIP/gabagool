package scenes

import (
	"github.com/veandco/go-sdl2/sdl"
	"nextui-sdl2/ui"
)

type KeyboardScene struct {
	window   *ui.Window
	keyboard *ui.VirtualKeyboard
}

func NewKeyboardScene(window *ui.Window) *KeyboardScene {
	scene := &KeyboardScene{
		window: window,
	}

	scene.keyboard = ui.CreateKeyboard(window.Width, window.Height)

	return scene
}

// Init initializes the scene
func (s *KeyboardScene) Init() error {
	s.keyboard.TextBuffer = ""
	s.keyboard.CurrentState = ui.LowerCase
	s.keyboard.ShiftPressed = false

	s.keyboard.SelectedKeyIndex = 0
	s.keyboard.SelectedSpecial = 0

	if len(s.keyboard.Keys) > 0 {
		s.keyboard.Keys[0].IsPressed = true
	}

	return nil
}

func (s *KeyboardScene) HandleEvent(event sdl.Event) bool {
	switch t := event.(type) {
	case *sdl.KeyboardEvent:
		if t.Type == sdl.KEYDOWN {
			s.keyboard.ProcessKeyboardInput(t.Keysym.Sym)
			return true
		}

	case *sdl.ControllerButtonEvent:
		if t.Type == sdl.CONTROLLERBUTTONDOWN {
			switch t.Button {
			case sdl.CONTROLLER_BUTTON_DPAD_UP:
				s.keyboard.ProcessNavigation(3) // Up
				return true
			case sdl.CONTROLLER_BUTTON_DPAD_DOWN:
				s.keyboard.ProcessNavigation(4) // Down
				return true
			case sdl.CONTROLLER_BUTTON_DPAD_LEFT:
				s.keyboard.ProcessNavigation(2) // Left
				return true
			case sdl.CONTROLLER_BUTTON_DPAD_RIGHT:
				s.keyboard.ProcessNavigation(1) // Right
				return true
			case sdl.CONTROLLER_BUTTON_A:
				s.keyboard.ProcessSelection()
				return true
			case sdl.CONTROLLER_BUTTON_B:
				ui.GetSceneManager().SwitchTo("mainMenu")
				return true
			}
		}

	case *sdl.ControllerAxisEvent:
		if t.Value > 20000 || t.Value < -20000 {
			if t.Axis == 0 { // Left stick horizontal
				if t.Value > 20000 {
					s.keyboard.ProcessNavigation(1) // Right
				} else {
					s.keyboard.ProcessNavigation(2) // Left
				}
				return true
			} else if t.Axis == 1 { // Left stick vertical
				if t.Value > 20000 {
					s.keyboard.ProcessNavigation(4) // Down
				} else {
					s.keyboard.ProcessNavigation(3) // Up
				}
				return true
			}
		}
	}

	return false
}

func (s *KeyboardScene) Update() error {
	return nil
}

// Render draws the scene
func (s *KeyboardScene) Render() error {
	s.window.Renderer.SetDrawColor(20, 20, 20, 255)
	s.window.Renderer.Clear()

	s.keyboard.Render(s.window.Renderer, ui.GetFont())

	return nil
}

func (s *KeyboardScene) Destroy() error {
	return nil
}
