package gabagool

import (
	"github.com/veandco/go-sdl2/sdl"
)

type InternalButton int

const (
	InternalButtonUnassigned InternalButton = iota
	InternalButtonUp
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
)

type InputSource int

const (
	InputSourceKeyboard InputSource = iota
	InputSourceController
	InputSourceJoystick
	InputSourceHatSwitch
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

	ControllerHatMap map[uint8]InternalButton

	JoystickAxisMap map[uint8]struct {
		PositiveButton InternalButton
		NegativeButton InternalButton
		Threshold      int16
	}

	JoystickButtonMap map[uint8]InternalButton

	JoystickHatMap map[uint8]InternalButton
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
			3:  InternalButtonA,
			4:  InternalButtonB,
			6:  InternalButtonX,
			5:  InternalButtonY,
			9:  InternalButtonSelect,
			10: InternalButtonStart,
			14: InternalButtonMenu,
			7:  InternalButtonL1,
			8:  InternalButtonR1,
			12: InternalButtonL2,
			13: InternalButtonR2,
		},
		JoystickHatMap: map[uint8]InternalButton{
			sdl.HAT_UP:    InternalButtonUp,
			sdl.HAT_DOWN:  InternalButtonDown,
			sdl.HAT_LEFT:  InternalButtonLeft,
			sdl.HAT_RIGHT: InternalButtonRight,
		},
	}
}
