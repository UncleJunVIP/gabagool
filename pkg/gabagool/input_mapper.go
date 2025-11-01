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
			sdl.CONTROLLER_BUTTON_A:          InternalButtonB,
			sdl.CONTROLLER_BUTTON_B:          InternalButtonA,
			sdl.CONTROLLER_BUTTON_X:          InternalButtonY,
			sdl.CONTROLLER_BUTTON_Y:          InternalButtonX,
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
