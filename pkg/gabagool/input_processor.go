package gabagool

import (
	"github.com/veandco/go-sdl2/sdl"
)

var globalInputProcessor *InputProcessor
var gameControllers []*sdl.GameController
var rawJoysticks []*sdl.Joystick

func InitInputProcessor() {
	globalInputProcessor = NewInputProcessor()
	DetectAndOpenAllGameControllers()
}

func GetInputProcessor() *InputProcessor {
	if globalInputProcessor == nil {
		InitInputProcessor()
	}
	return globalInputProcessor
}

type InputProcessor struct {
	mapping                       *InternalInputMapping
	gameControllerJoystickIndices map[int]bool
}

func NewInputProcessor() *InputProcessor {
	return &InputProcessor{
		mapping:                       GetInputMapping(),
		gameControllerJoystickIndices: make(map[int]bool),
	}
}

func (ip *InputProcessor) RegisterGameControllerJoystickIndex(joystickIndex int) {
	ip.gameControllerJoystickIndices[joystickIndex] = true
}

func (ip *InputProcessor) IsGameControllerJoystick(joystickIndex int) bool {
	return ip.gameControllerJoystickIndices[joystickIndex]
}

func (ip *InputProcessor) ProcessSDLEvent(event sdl.Event) *InputEvent {
	GetLoggerInstance().Debug("Processing SDL event", "eventType", event.GetType())
	switch e := event.(type) {
	case *sdl.KeyboardEvent:
		keyCode := e.Keysym.Sym
		GetLoggerInstance().Debug("Keyboard input detected", "keyCode", keyCode)
		if button, exists := ip.mapping.KeyboardMap[keyCode]; exists {
			GetLoggerInstance().Debug("Keyboard input mapped",
				"keyCode", keyCode,
				"button", button)
			return &InputEvent{
				Button:  button,
				Pressed: e.Type == sdl.KEYDOWN,
				Source:  InputSourceKeyboard,
				RawCode: int(keyCode),
			}
		}
		GetLoggerInstance().Debug("Keyboard input not mapped",
			"keyCode", keyCode,
			"mappingSize", len(ip.mapping.KeyboardMap))
	case *sdl.ControllerButtonEvent:
		GetLoggerInstance().Debug("Controller button input detected", "button", e.Button)
		if button, exists := ip.mapping.ControllerButtonMap[sdl.GameControllerButton(e.Button)]; exists {
			GetLoggerInstance().Debug("Controller button mapped", "button", e.Button, "mappedTo", button)
			return &InputEvent{
				Button:  button,
				Pressed: e.Type == sdl.CONTROLLERBUTTONDOWN,
				Source:  InputSourceController,
				RawCode: int(e.Button),
			}
		}
		GetLoggerInstance().Debug("Controller button not mapped", "button", e.Button)
	case *sdl.JoyHatEvent:
		GetLoggerInstance().Debug("Joy hat input detected", "value", e.Value)
		if e.Value != sdl.HAT_CENTERED {
			if button, exists := ip.mapping.JoystickHatMap[e.Value]; exists {
				GetLoggerInstance().Debug("Joy hat mapped", "value", e.Value, "mappedTo", button)
				return &InputEvent{
					Button:  button,
					Pressed: true,
					Source:  InputSourceHatSwitch,
					RawCode: int(e.Value),
				}
			}
		}
		GetLoggerInstance().Debug("Joy hat not mapped", "value", e.Value)
	case *sdl.ControllerAxisEvent:
		GetLoggerInstance().Debug("Controller axis input detected", "axis", e.Axis, "value", e.Value)
		if axisConfig, exists := ip.mapping.JoystickAxisMap[e.Axis]; exists {
			if e.Value > axisConfig.Threshold {
				GetLoggerInstance().Debug("Controller axis positive threshold exceeded",
					"axis", e.Axis,
					"value", e.Value,
					"threshold", axisConfig.Threshold,
					"mappedTo", axisConfig.PositiveButton)
				return &InputEvent{
					Button:  axisConfig.PositiveButton,
					Pressed: true,
					Source:  InputSourceController,
					RawCode: int(e.Axis),
				}
			}
			if e.Value < -axisConfig.Threshold {
				GetLoggerInstance().Debug("Controller axis negative threshold exceeded",
					"axis", e.Axis,
					"value", e.Value,
					"threshold", axisConfig.Threshold,
					"mappedTo", axisConfig.NegativeButton)
				return &InputEvent{
					Button:  axisConfig.NegativeButton,
					Pressed: true,
					Source:  InputSourceController,
					RawCode: int(e.Axis),
				}
			}
		}
		GetLoggerInstance().Debug("Controller axis not mapped or threshold not exceeded",
			"axis", e.Axis,
			"value", e.Value)
	case *sdl.JoyButtonEvent:
		GetLoggerInstance().Debug("Joy button input detected", "button", e.Button)
		if button, exists := ip.mapping.JoystickButtonMap[e.Button]; exists {
			GetLoggerInstance().Debug("Joy button mapped", "button", e.Button, "mappedTo", button)
			return &InputEvent{
				Button:  button,
				Pressed: e.Type == sdl.JOYBUTTONDOWN,
				Source:  InputSourceJoystick,
				RawCode: int(e.Button),
			}
		}
		GetLoggerInstance().Debug("Joy button not mapped", "button", e.Button)
	case *sdl.JoyAxisEvent:
		GetLoggerInstance().Debug("Joy axis input detected", "axis", e.Axis, "value", e.Value)
		if axisConfig, exists := ip.mapping.JoystickAxisMap[e.Axis]; exists {
			if e.Value > axisConfig.Threshold {
				GetLoggerInstance().Debug("Joy axis positive threshold exceeded",
					"axis", e.Axis,
					"value", e.Value,
					"threshold", axisConfig.Threshold,
					"mappedTo", axisConfig.PositiveButton)
				return &InputEvent{
					Button:  axisConfig.PositiveButton,
					Pressed: true,
					Source:  InputSourceJoystick,
					RawCode: int(e.Axis),
				}
			}
			if e.Value < -axisConfig.Threshold {
				GetLoggerInstance().Debug("Joy axis negative threshold exceeded",
					"axis", e.Axis,
					"value", e.Value,
					"threshold", axisConfig.Threshold,
					"mappedTo", axisConfig.NegativeButton)
				return &InputEvent{
					Button:  axisConfig.NegativeButton,
					Pressed: true,
					Source:  InputSourceJoystick,
					RawCode: int(e.Axis),
				}
			}
		}
		GetLoggerInstance().Debug("Joy axis not mapped or threshold not exceeded",
			"axis", e.Axis,
			"value", e.Value)
	}
	return nil
}

func DetectAndOpenAllGameControllers() {
	numJoysticks := sdl.NumJoysticks()
	GetLoggerInstance().Debug("Detecting controllers", "numJoysticks", numJoysticks)

	processor := GetInputProcessor()

	for i := 0; i < numJoysticks; i++ {
		if sdl.IsGameController(i) {
			controller := sdl.GameControllerOpen(i)
			if controller != nil {
				name := controller.Name()

				// Register this joystick INDEX (not instance ID) as being handled by a game controller
				processor.RegisterGameControllerJoystickIndex(i)

				GetLoggerInstance().Debug("Opened game controller",
					"index", i,
					"name", name,
				)

				gameControllers = append(gameControllers, controller)
			} else {
				GetLoggerInstance().Error("Failed to open game controller", "index", i)
			}
		} else {
			joystick := sdl.JoystickOpen(i)
			if joystick != nil {
				name := joystick.Name()
				GetLoggerInstance().Debug("Opened raw joystick (not a standard game controller)",
					"index", i,
					"name", name,
				)

				rawJoysticks = append(rawJoysticks, joystick)
			} else {
				GetLoggerInstance().Debug("Failed to open raw joystick", "index", i)
			}
		}
	}

	GetLoggerInstance().Debug("Controller detection complete",
		"gameControllers", len(gameControllers),
		"rawJoysticks", len(rawJoysticks),
		"totalJoysticks", numJoysticks,
	)
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
