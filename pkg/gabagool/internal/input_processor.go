package internal

import (
	"github.com/veandco/go-sdl2/sdl"
)

var globalInputProcessor *Processor
var gameControllers []*sdl.GameController
var rawJoysticks []*sdl.Joystick

func InitInputProcessor() {
	globalInputProcessor = NewInputProcessor()
}

func GetInputProcessor() *Processor {
	if globalInputProcessor == nil {
		InitInputProcessor()
	}
	return globalInputProcessor
}

type Processor struct {
	mapping                       *InternalInputMapping
	gameControllerJoystickIndices map[int]bool
}

func NewInputProcessor() *Processor {
	return &Processor{
		mapping:                       GetInputMapping(),
		gameControllerJoystickIndices: make(map[int]bool),
	}
}

func (ip *Processor) RegisterGameControllerJoystickIndex(joystickIndex int) {
	ip.gameControllerJoystickIndices[joystickIndex] = true
}

func (ip *Processor) IsGameControllerJoystick(joystickIndex int) bool {
	return ip.gameControllerJoystickIndices[joystickIndex]
}

func (ip *Processor) ProcessSDLEvent(event sdl.Event) *Event {
	logger := GetInternalLogger()
	switch e := event.(type) {
	case *sdl.KeyboardEvent:
		keyCode := e.Keysym.Sym
		if button, exists := ip.mapping.KeyboardMap[keyCode]; exists {
			logger.Debug("Keyboard input mapped",
				"keyCode", keyCode,
				"button", button)
			return &Event{
				Button:  button,
				Pressed: e.Type == sdl.KEYDOWN,
				Source:  SourceKeyboard,
				RawCode: int(keyCode),
			}
		}
		logger.Debug("Keyboard input not mapped",
			"keyCode", keyCode,
			"mappingSize", len(ip.mapping.KeyboardMap))
	case *sdl.ControllerButtonEvent:
		if button, exists := ip.mapping.ControllerButtonMap[sdl.GameControllerButton(e.Button)]; exists {
			logger.Debug("Controller button mapped", "button", e.Button, "mappedTo", button)
			return &Event{
				Button:  button,
				Pressed: e.Type == sdl.CONTROLLERBUTTONDOWN,
				Source:  SourceController,
				RawCode: int(e.Button),
			}
		}
		logger.Debug("Controller button not mapped", "button", e.Button)
	case *sdl.JoyHatEvent:
		if e.Value != sdl.HAT_CENTERED {
			if button, exists := ip.mapping.JoystickHatMap[e.Value]; exists {
				logger.Debug("Joy hat mapped", "value", e.Value, "mappedTo", button)
				return &Event{
					Button:  button,
					Pressed: true,
					Source:  SourceHatSwitch,
					RawCode: int(e.Value),
				}
			}
		}
		logger.Debug("Joy hat not mapped", "value", e.Value)
	case *sdl.ControllerAxisEvent:
		if axisConfig, exists := ip.mapping.JoystickAxisMap[e.Axis]; exists {
			if e.Value > axisConfig.Threshold {
				logger.Debug("Controller axis positive threshold exceeded",
					"axis", e.Axis,
					"value", e.Value,
					"threshold", axisConfig.Threshold,
					"mappedTo", axisConfig.PositiveButton)
				return &Event{
					Button:  axisConfig.PositiveButton,
					Pressed: true,
					Source:  SourceController,
					RawCode: int(e.Axis),
				}
			}
			if e.Value < -axisConfig.Threshold {
				logger.Debug("Controller axis negative threshold exceeded",
					"axis", e.Axis,
					"value", e.Value,
					"threshold", axisConfig.Threshold,
					"mappedTo", axisConfig.NegativeButton)
				return &Event{
					Button:  axisConfig.NegativeButton,
					Pressed: true,
					Source:  SourceController,
					RawCode: int(e.Axis),
				}
			}
		}
		logger.Debug("Controller axis not mapped or threshold not exceeded",
			"axis", e.Axis,
			"value", e.Value)
	case *sdl.JoyButtonEvent:
		if button, exists := ip.mapping.JoystickButtonMap[e.Button]; exists {
			logger.Debug("Joy button mapped", "button", e.Button, "mappedTo", button)
			return &Event{
				Button:  button,
				Pressed: e.Type == sdl.JOYBUTTONDOWN,
				Source:  SourceJoystick,
				RawCode: int(e.Button),
			}
		}
		logger.Debug("Joy button not mapped", "button", e.Button)
	case *sdl.JoyAxisEvent:
		if axisConfig, exists := ip.mapping.JoystickAxisMap[e.Axis]; exists {
			if e.Value > axisConfig.Threshold {
				logger.Debug("Joy axis positive threshold exceeded",
					"axis", e.Axis,
					"value", e.Value,
					"threshold", axisConfig.Threshold,
					"mappedTo", axisConfig.PositiveButton)
				return &Event{
					Button:  axisConfig.PositiveButton,
					Pressed: true,
					Source:  SourceJoystick,
					RawCode: int(e.Axis),
				}
			}
			if e.Value < -axisConfig.Threshold {
				logger.Debug("Joy axis negative threshold exceeded",
					"axis", e.Axis,
					"value", e.Value,
					"threshold", axisConfig.Threshold,
					"mappedTo", axisConfig.NegativeButton)
				return &Event{
					Button:  axisConfig.NegativeButton,
					Pressed: true,
					Source:  SourceJoystick,
					RawCode: int(e.Axis),
				}
			}
		}
		logger.Debug("Joy axis not mapped or threshold not exceeded",
			"axis", e.Axis,
			"value", e.Value)
	}
	return nil
}

func CloseAllControllers() {
	for _, controller := range gameControllers {
		if controller != nil {
			controller.Close()
		}
	}
	for _, joystick := range rawJoysticks {
		if joystick != nil {
			joystick.Close()
		}
	}
}
