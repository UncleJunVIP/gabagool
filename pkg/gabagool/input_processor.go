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
	switch e := event.(type) {
	case *sdl.KeyboardEvent:
		keyCode := e.Keysym.Sym
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
		if button, exists := ip.mapping.ControllerButtonMap[sdl.GameControllerButton(e.Button)]; exists {
			return &InputEvent{
				Button:  button,
				Pressed: e.Type == sdl.CONTROLLERBUTTONDOWN,
				Source:  InputSourceController,
				RawCode: int(e.Button),
			}
		}
	case *sdl.JoyHatEvent:
		if e.Value != sdl.HAT_CENTERED {
			if button, exists := ip.mapping.JoystickHatMap[e.Value]; exists {
				return &InputEvent{
					Button:  button,
					Pressed: true,
					Source:  InputSourceHatSwitch,
					RawCode: int(e.Value),
				}
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
