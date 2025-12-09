package internal

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

var globalInputProcessor *Processor
var gameControllers []*sdl.GameController
var rawJoysticks []*sdl.Joystick

func InitInputProcessor() {
	globalInputProcessor = NewInputProcessor()

	numJoysticks := sdl.NumJoysticks()
	GetInternalLogger().Debug("Detecting controllers", "joystick_count", numJoysticks)

	for i := 0; i < numJoysticks; i++ {
		if sdl.IsGameController(i) {
			controller := sdl.GameControllerOpen(i)
			if controller != nil {
				name := controller.Name()

				// Register this joystick INDEX (not instance ID) as being handled by a game controller
				globalInputProcessor.RegisterGameControllerJoystickIndex(i)

				GetInternalLogger().Debug("Opened game controller",
					"index", i,
					"name", name,
				)

				gameControllers = append(gameControllers, controller)
			} else {
				GetInternalLogger().Error("Failed to open game controller", "index", i)
			}
		} else {
			joystick := sdl.JoystickOpen(i)
			if joystick != nil {
				name := joystick.Name()
				GetInternalLogger().Debug("Opened raw joystick (not a standard game controller)",
					"index", i,
					"name", name,
				)

				rawJoysticks = append(rawJoysticks, joystick)
			} else {
				GetInternalLogger().Debug("Failed to open raw joystick", "index", i)
			}
		}
	}

	GetInternalLogger().Debug("Controller detection complete",
		"game_controllers", len(gameControllers),
		"raw_joysticks", len(rawJoysticks),
		"total_joysticks", numJoysticks,
	)
}

func GetInputProcessor() *Processor {
	return globalInputProcessor
}

type Processor struct {
	mapping                       *InputMapping
	gameControllerJoystickIndices map[int]bool
	axisStates                    map[uint8]int8  // tracks which direction each axis is pressed: -1 (negative), 0 (none), 1 (positive)
	hatStates                     map[uint8]uint8 // tracks the current hat position
}

func NewInputProcessor() *Processor {
	return &Processor{
		mapping:                       GetInputMapping(),
		gameControllerJoystickIndices: make(map[int]bool),
		axisStates:                    make(map[uint8]int8),
		hatStates:                     make(map[uint8]uint8),
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
		keyName := sdl.GetKeyName(keyCode)
		if button, exists := ip.mapping.KeyboardMap[keyCode]; exists {
			if e.Type == sdl.KEYDOWN {
				logger.Debug("Keyboard input mapped",
					"physical", keyName,
					"keyCode", fmt.Sprintf("%s (%d)", keyName, keyCode),
					"virtualButton", button.GetName())
			}
			return &Event{
				Button:  button,
				Pressed: e.Type == sdl.KEYDOWN,
				Source:  SourceKeyboard,
				RawCode: int(keyCode),
			}
		}
		logger.Debug("Keyboard input not mapped",
			"key_code", fmt.Sprintf("%s (%d)", keyName, keyCode),
			"mappingSize", len(ip.mapping.KeyboardMap))
	case *sdl.ControllerButtonEvent:
		buttonName := sdl.GameControllerGetStringForButton(sdl.GameControllerButton(e.Button))
		if button, exists := ip.mapping.ControllerButtonMap[sdl.GameControllerButton(e.Button)]; exists {
			if e.Type == sdl.CONTROLLERBUTTONDOWN {
				logger.Debug("Controller button mapped",
					"physical", buttonName,
					"buttonCode", fmt.Sprintf("%s (%d)", buttonName, e.Button),
					"virtualButton", button.GetName())
			}
			return &Event{
				Button:  button,
				Pressed: e.Type == sdl.CONTROLLERBUTTONDOWN,
				Source:  SourceController,
				RawCode: int(e.Button),
			}
		}
		logger.Debug("Controller button not mapped",
			"button_code", fmt.Sprintf("%s (%d)", buttonName, e.Button))
	case *sdl.JoyHatEvent:
		previousValue := ip.hatStates[e.Hat]
		ip.hatStates[e.Hat] = e.Value

		// If hat moved to center, generate release event for previous direction
		if e.Value == sdl.HAT_CENTERED && previousValue != sdl.HAT_CENTERED {
			if button, exists := ip.mapping.JoystickHatMap[previousValue]; exists {
				hatDirection := getHatDirectionName(previousValue)
				logger.Debug("Joy hat released",
					"hat_value", fmt.Sprintf("%s (%d)", hatDirection, previousValue),
					"virtual_button", button.GetName())
				return &Event{
					Button:  button,
					Pressed: false,
					Source:  SourceHatSwitch,
					RawCode: int(previousValue),
				}
			}
		}

		// If hat moved to a new direction, generate press event
		if e.Value != sdl.HAT_CENTERED {
			hatDirection := getHatDirectionName(e.Value)
			if button, exists := ip.mapping.JoystickHatMap[e.Value]; exists {
				logger.Debug("Joy hat mapped",
					"hat_value", fmt.Sprintf("%s (%d)", hatDirection, e.Value),
					"virtual_button", button.GetName())
				return &Event{
					Button:  button,
					Pressed: true,
					Source:  SourceHatSwitch,
					RawCode: int(e.Value),
				}
			}
			logger.Debug("Joy hat not mapped",
				"hat_value", fmt.Sprintf("%s (%d)", hatDirection, e.Value))
		}
	case *sdl.ControllerAxisEvent:
		axisName := sdl.GameControllerGetStringForAxis(sdl.GameControllerAxis(e.Axis))
		if axisConfig, exists := ip.mapping.JoystickAxisMap[e.Axis]; exists {
			previousState := ip.axisStates[e.Axis]
			var newState int8 = 0

			// Determine new state
			if e.Value > axisConfig.Threshold {
				newState = 1
			} else if e.Value < -axisConfig.Threshold {
				newState = -1
			}

			// If state changed, generate appropriate event
			if newState != previousState {
				ip.axisStates[e.Axis] = newState

				// Generate release event for previous state
				if previousState == 1 {
					logger.Debug("Controller axis positive released",
						"axis_code", fmt.Sprintf("%s+ (%d)", axisName, e.Axis),
						"value", e.Value,
						"virtual_button", axisConfig.PositiveButton.GetName())
					return &Event{
						Button:  axisConfig.PositiveButton,
						Pressed: false,
						Source:  SourceController,
						RawCode: int(e.Axis),
					}
				} else if previousState == -1 {
					logger.Debug("Controller axis negative released",
						"axis_code", fmt.Sprintf("%s- (%d)", axisName, e.Axis),
						"value", e.Value,
						"virtual_button", axisConfig.NegativeButton.GetName())
					return &Event{
						Button:  axisConfig.NegativeButton,
						Pressed: false,
						Source:  SourceController,
						RawCode: int(e.Axis),
					}
				}

				// Generate press event for new state
				if newState == 1 {
					logger.Debug("Controller axis positive threshold exceeded",
						"axis_code", fmt.Sprintf("%s+ (%d)", axisName, e.Axis),
						"value", e.Value,
						"threshold", axisConfig.Threshold,
						"virtual_button", axisConfig.PositiveButton.GetName())
					return &Event{
						Button:  axisConfig.PositiveButton,
						Pressed: true,
						Source:  SourceController,
						RawCode: int(e.Axis),
					}
				} else if newState == -1 {
					logger.Debug("Controller axis negative threshold exceeded",
						"axis_code", fmt.Sprintf("%s- (%d)", axisName, e.Axis),
						"value", e.Value,
						"threshold", axisConfig.Threshold,
						"virtual_button", axisConfig.NegativeButton.GetName())
					return &Event{
						Button:  axisConfig.NegativeButton,
						Pressed: true,
						Source:  SourceController,
						RawCode: int(e.Axis),
					}
				}
			}
		}
		logger.Debug("Controller axis not mapped or threshold not exceeded",
			"axis_code", fmt.Sprintf("%s (%d)", axisName, e.Axis),
			"value", e.Value)
	case *sdl.JoyButtonEvent:
		joyButtonName := getJoyButtonName(e.Button)
		if button, exists := ip.mapping.JoystickButtonMap[e.Button]; exists {
			logger.Debug("Joy button mapped",
				"button_code", fmt.Sprintf("%s (%d)", joyButtonName, e.Button),
				"virtual_button", button.GetName())
			return &Event{
				Button:  button,
				Pressed: e.Type == sdl.JOYBUTTONDOWN,
				Source:  SourceJoystick,
				RawCode: int(e.Button),
			}
		}
		logger.Debug("Joy button not mapped",
			"button_code", fmt.Sprintf("%s (%d)", joyButtonName, e.Button))
	case *sdl.JoyAxisEvent:
		joyAxisName := getJoyAxisName(e.Axis)
		if axisConfig, exists := ip.mapping.JoystickAxisMap[e.Axis]; exists {
			previousState := ip.axisStates[e.Axis]
			var newState int8 = 0

			// Determine new state
			if e.Value > axisConfig.Threshold {
				newState = 1
			} else if e.Value < -axisConfig.Threshold {
				newState = -1
			}

			// If state changed, generate appropriate event
			if newState != previousState {
				ip.axisStates[e.Axis] = newState

				// Generate release event for previous state
				if previousState == 1 {
					logger.Debug("Joy axis positive released",
						"axis_code", fmt.Sprintf("%s+ (%d)", joyAxisName, e.Axis),
						"value", e.Value,
						"virtual_button", axisConfig.PositiveButton.GetName())
					return &Event{
						Button:  axisConfig.PositiveButton,
						Pressed: false,
						Source:  SourceJoystick,
						RawCode: int(e.Axis),
					}
				} else if previousState == -1 {
					logger.Debug("Joy axis negative released",
						"axis_code", fmt.Sprintf("%s- (%d)", joyAxisName, e.Axis),
						"value", e.Value,
						"virtual_button", axisConfig.NegativeButton.GetName())
					return &Event{
						Button:  axisConfig.NegativeButton,
						Pressed: false,
						Source:  SourceJoystick,
						RawCode: int(e.Axis),
					}
				}

				// Generate press event for new state
				if newState == 1 {
					logger.Debug("Joy axis positive threshold exceeded",
						"axis_code", fmt.Sprintf("%s+ (%d)", joyAxisName, e.Axis),
						"value", e.Value,
						"threshold", axisConfig.Threshold,
						"virtual_button", axisConfig.PositiveButton.GetName())
					return &Event{
						Button:  axisConfig.PositiveButton,
						Pressed: true,
						Source:  SourceJoystick,
						RawCode: int(e.Axis),
					}
				} else if newState == -1 {
					logger.Debug("Joy axis negative threshold exceeded",
						"axis_code", fmt.Sprintf("%s- (%d)", joyAxisName, e.Axis),
						"value", e.Value,
						"threshold", axisConfig.Threshold,
						"virtual_button", axisConfig.NegativeButton.GetName())
					return &Event{
						Button:  axisConfig.NegativeButton,
						Pressed: true,
						Source:  SourceJoystick,
						RawCode: int(e.Axis),
					}
				}
			}
		}
		logger.Debug("Joy axis not mapped or threshold not exceeded",
			"axis_code", fmt.Sprintf("%s (%d)", joyAxisName, e.Axis),
			"value", e.Value)
	}
	return nil
}

func getHatDirectionName(value uint8) string {
	switch value {
	case sdl.HAT_UP:
		return "Hat Up"
	case sdl.HAT_DOWN:
		return "Hat Down"
	case sdl.HAT_LEFT:
		return "Hat Left"
	case sdl.HAT_RIGHT:
		return "Hat Right"
	case sdl.HAT_LEFTUP:
		return "Hat Left Up"
	case sdl.HAT_LEFTDOWN:
		return "Hat Left Down"
	case sdl.HAT_RIGHTUP:
		return "Hat Right Up"
	case sdl.HAT_RIGHTDOWN:
		return "Hat Right Down"
	default:
		return "Hat Unknown"
	}
}

func getJoyButtonName(button uint8) string {
	return fmt.Sprintf("JoyButton%d", button)
}

func getJoyAxisName(axis uint8) string {
	return fmt.Sprintf("JoyAxis%d", axis)
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
