package gabagool

import (
	"log/slog"

	"github.com/veandco/go-sdl2/sdl"
)

var globalInputProcessor *InputProcessor

func InitInputProcessor() {
	globalInputProcessor = NewInputProcessor()
}

func GetInputProcessor() *InputProcessor {
	if globalInputProcessor == nil {
		InitInputProcessor()
	}
	return globalInputProcessor
}

type InputProcessor struct {
	mapping                       *InputMapping
	gameControllerJoystickIndices map[int]bool // Track which joystick indices are game controllers
}

func NewInputProcessor() *InputProcessor {
	return &InputProcessor{
		mapping:                       DefaultInputMapping(),
		gameControllerJoystickIndices: make(map[int]bool),
	}
}

func (ip *InputProcessor) SetMapping(mapping *InputMapping) {
	ip.mapping = mapping
}

// RegisterGameControllerJoystickIndex marks a joystick index as being handled by a game controller
func (ip *InputProcessor) RegisterGameControllerJoystickIndex(joystickIndex int) {
	ip.gameControllerJoystickIndices[joystickIndex] = true
}

// IsGameControllerJoystick checks if this joystick index is already handled as a game controller
func (ip *InputProcessor) IsGameControllerJoystick(joystickIndex int) bool {
	return ip.gameControllerJoystickIndices[joystickIndex]
}

func (ip *InputProcessor) ProcessSDLEvent(event sdl.Event) *InputEvent {
	slog.Debug("Event detected", "event", event)
	switch e := event.(type) {
	case *sdl.KeyboardEvent:
		if button, exists := ip.mapping.KeyboardMap[e.Keysym.Sym]; exists {
			return &InputEvent{
				Button:  button,
				Pressed: e.Type == sdl.KEYDOWN,
				Source:  InputSourceKeyboard,
				RawCode: int(e.Keysym.Sym),
			}
		}
	case *sdl.ControllerButtonEvent:
		if button, exists := ip.mapping.ControllerButtonMap[sdl.GameControllerButton(e.Button)]; exists {
			return &InputEvent{
				Button:  button,
				Pressed: e.Type == sdl.CONTROLLERBUTTONDOWN,
				Source:  InputSourceController,
				RawCode: int(e.Button),
			}
		}
	case *sdl.JoyHatEvent:
		if button, exists := ip.mapping.JoystickHatMap[e.Value]; exists {
			return &InputEvent{
				Button:  button,
				Pressed: true,
				Source:  InputSourceHatSwitch,
				RawCode: int(e.Value),
			}
		}
	case *sdl.ControllerAxisEvent:
		if axisConfig, exists := ip.mapping.JoystickAxisMap[e.Axis]; exists {
			if e.Value > axisConfig.Threshold {
				return &InputEvent{
					Button:  axisConfig.PositiveButton,
					Pressed: true,
					Source:  InputSourceController,
					RawCode: int(e.Axis),
				}
			}
			if e.Value < -axisConfig.Threshold {
				return &InputEvent{
					Button:  axisConfig.NegativeButton,
					Pressed: true,
					Source:  InputSourceController,
					RawCode: int(e.Axis),
				}
			}
		}
	case *sdl.JoyButtonEvent:
		// Skip if this joystick index is already handled as a game controller
		if ip.IsGameControllerJoystick(int(e.Which)) {
			return nil
		}
		if button, exists := ip.mapping.JoystickButtonMap[e.Button]; exists {
			return &InputEvent{
				Button:  button,
				Pressed: e.Type == sdl.JOYBUTTONDOWN,
				Source:  InputSourceJoystick,
				RawCode: int(e.Button),
			}
		}
	case *sdl.JoyAxisEvent:
		// Skip if this joystick index is already handled as a game controller
		if ip.IsGameControllerJoystick(int(e.Which)) {
			return nil
		}
		if axisConfig, exists := ip.mapping.JoystickAxisMap[e.Axis]; exists {
			if e.Value > axisConfig.Threshold {
				return &InputEvent{
					Button:  axisConfig.PositiveButton,
					Pressed: true,
					Source:  InputSourceJoystick,
					RawCode: int(e.Axis),
				}
			}
			if e.Value < -axisConfig.Threshold {
				return &InputEvent{
					Button:  axisConfig.NegativeButton,
					Pressed: true,
					Source:  InputSourceJoystick,
					RawCode: int(e.Axis),
				}
			}
		}
	}
	return nil
}
