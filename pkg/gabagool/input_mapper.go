package gabagool

import (
	"github.com/veandco/go-sdl2/sdl"
)

type InternalButton int

const (
	InternalButtonUp InternalButton = iota
	InternalButtonDown
	InternalButtonLeft
	InternalButtonRight
	InternalButtonA
	InternalButtonB
	InternalButtonX
	InternalButtonY
	InternalButtonL1
	InternalButtonL2
	InternalButtonR1
	InternalButtonR2
	InternalButtonStart
	InternalButtonSelect
	InternalButtonMenu
	InternalButtonUnassigned
)

type InputSource int

const (
	InputSourceKeyboard InputSource = iota
	InputSourceController
	InputSourceJoystick
)

type InputEvent struct {
	Button  InternalButton
	Pressed bool
	Source  InputSource
	RawCode int
}

type InputMapping struct {
	KeyboardMap map[sdl.Keycode]InternalButton

	ControllerButtonMap map[sdl.GameControllerButton]InternalButton

	JoystickAxisMap map[uint8]struct {
		PositiveButton InternalButton
		NegativeButton InternalButton
		Threshold      int16
	}

	JoystickButtonMap map[uint8]InternalButton
}

func DefaultInputMapping() *InputMapping {
	return &InputMapping{
		KeyboardMap: map[sdl.Keycode]InternalButton{
			sdl.K_UP:     InternalButtonUp,
			sdl.K_DOWN:   InternalButtonDown,
			sdl.K_LEFT:   InternalButtonLeft,
			sdl.K_RIGHT:  InternalButtonRight,
			sdl.K_a:      InternalButtonA,
			sdl.K_b:      InternalButtonB,
			sdl.K_x:      InternalButtonX,
			sdl.K_y:      InternalButtonY,
			sdl.K_RETURN: InternalButtonStart,
			sdl.K_SPACE:  InternalButtonSelect,
			sdl.K_h:      InternalButtonMenu,
		},
		ControllerButtonMap: map[sdl.GameControllerButton]InternalButton{
			sdl.CONTROLLER_BUTTON_DPAD_UP:    InternalButtonUp,
			sdl.CONTROLLER_BUTTON_DPAD_DOWN:  InternalButtonDown,
			sdl.CONTROLLER_BUTTON_DPAD_LEFT:  InternalButtonLeft,
			sdl.CONTROLLER_BUTTON_DPAD_RIGHT: InternalButtonRight,
			sdl.CONTROLLER_BUTTON_A:          InternalButtonA,
			sdl.CONTROLLER_BUTTON_B:          InternalButtonB,
			sdl.CONTROLLER_BUTTON_X:          InternalButtonX,
			sdl.CONTROLLER_BUTTON_Y:          InternalButtonY,
			sdl.CONTROLLER_BUTTON_START:      InternalButtonStart,
			sdl.CONTROLLER_BUTTON_BACK:       InternalButtonSelect,
			sdl.CONTROLLER_BUTTON_GUIDE:      InternalButtonMenu,
		},
		JoystickAxisMap: map[uint8]struct {
			PositiveButton InternalButton
			NegativeButton InternalButton
			Threshold      int16
		}{
			0: {InternalButtonRight, InternalButtonLeft, 16000}, // X axis
			1: {InternalButtonDown, InternalButtonUp, 16000},    // Y axis
		},
		JoystickButtonMap: map[uint8]InternalButton{
			0: InternalButtonA,
			1: InternalButtonB,
			2: InternalButtonX,
			3: InternalButtonY,
			4: InternalButtonSelect,
			5: InternalButtonStart,
		},
	}
}

type InputProcessor struct {
	mapping *InputMapping
}

func NewInputProcessor() *InputProcessor {
	return &InputProcessor{
		mapping: DefaultInputMapping(),
	}
}

func (ip *InputProcessor) SetMapping(mapping *InputMapping) {
	ip.mapping = mapping
}

func (ip *InputProcessor) ProcessSDLEvent(event sdl.Event) *InputEvent {
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
		if button, exists := ip.mapping.JoystickButtonMap[e.Button]; exists {
			return &InputEvent{
				Button:  button,
				Pressed: e.Type == sdl.JOYBUTTONDOWN,
				Source:  InputSourceJoystick,
				RawCode: int(e.Button),
			}
		}
	case *sdl.JoyAxisEvent:
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
