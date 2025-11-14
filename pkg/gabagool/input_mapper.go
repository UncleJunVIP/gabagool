package gabagool

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

const InputMappingPathEnvVar = "INPUT_MAPPING_PATH"

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
	InternalButtonF1
	InternalButtonF2
	InternalButtonVolumeUp
	InternalButtonVolumeDown
	InternalButtonPower
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

type JoystickAxisMapping struct {
	PositiveButton InternalButton
	NegativeButton InternalButton
	Threshold      int16
}

type InternalInputMapping struct {
	KeyboardMap map[sdl.Keycode]InternalButton

	ControllerButtonMap map[sdl.GameControllerButton]InternalButton

	ControllerHatMap map[uint8]InternalButton

	JoystickAxisMap map[uint8]JoystickAxisMapping

	JoystickButtonMap map[uint8]InternalButton

	JoystickHatMap map[uint8]InternalButton
}

type InputMapping struct {
	KeyboardMap map[int]int `json:"keyboard_map"`

	ControllerButtonMap map[int]int `json:"controller_button_map"`

	ControllerHatMap map[int]int `json:"controller_hat_map"`

	JoystickAxisMap map[int]struct {
		PositiveButton int   `json:"positive_button"`
		NegativeButton int   `json:"negative_button"`
		Threshold      int16 `json:"threshold"`
	} `json:"joystick_axis_map"`

	JoystickButtonMap map[int]int `json:"joystick_button_map"`

	JoystickHatMap map[int]int `json:"joystick_hat_map"`
}

func DefaultInputMapping() *InternalInputMapping {
	return &InternalInputMapping{
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
	}
}

// GetInputMapping returns the input mapping from the environment variable if set,
// otherwise returns the default mapping
func GetInputMapping() *InternalInputMapping {
	mappingPath := os.Getenv(InputMappingPathEnvVar)
	if mappingPath != "" {
		mapping, err := LoadInputMappingFromJSON(mappingPath)
		if err == nil {
			logger := GetLoggerInstance()
			logger.Info("Loaded custom input mapping from environment variable", "path", mappingPath)
			return mapping
		}
		logger := GetLoggerInstance()
		logger.Warn("Failed to load custom input mapping, using default", "path", mappingPath, "error", err)
	}
	return DefaultInputMapping()
}

func LoadInputMappingFromJSON(filePath string) (*InternalInputMapping, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %w", err)
	}

	var serializableMapping InputMapping
	err = json.Unmarshal(data, &serializableMapping)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	mapping := &InternalInputMapping{
		KeyboardMap:         make(map[sdl.Keycode]InternalButton),
		ControllerButtonMap: make(map[sdl.GameControllerButton]InternalButton),
		ControllerHatMap:    make(map[uint8]InternalButton),
		JoystickAxisMap:     make(map[uint8]JoystickAxisMapping),
		JoystickButtonMap:   make(map[uint8]InternalButton),
		JoystickHatMap:      make(map[uint8]InternalButton),
	}

	if serializableMapping.KeyboardMap != nil {
		for keyCode, button := range serializableMapping.KeyboardMap {
			mapping.KeyboardMap[sdl.Keycode(keyCode)] = InternalButton(button)
		}
	}

	if serializableMapping.ControllerButtonMap != nil {
		for button, internalButton := range serializableMapping.ControllerButtonMap {
			mapping.ControllerButtonMap[sdl.GameControllerButton(button)] = InternalButton(internalButton)
		}
	}

	if serializableMapping.ControllerHatMap != nil {
		for hat, button := range serializableMapping.ControllerHatMap {
			mapping.ControllerHatMap[uint8(hat)] = InternalButton(button)
		}
	}

	if serializableMapping.JoystickAxisMap != nil {
		for axis, axisMapping := range serializableMapping.JoystickAxisMap {
			mapping.JoystickAxisMap[uint8(axis)] = JoystickAxisMapping{
				PositiveButton: InternalButton(axisMapping.PositiveButton),
				NegativeButton: InternalButton(axisMapping.NegativeButton),
				Threshold:      axisMapping.Threshold,
			}
		}
	}

	if serializableMapping.JoystickButtonMap != nil {
		for button, internalButton := range serializableMapping.JoystickButtonMap {
			mapping.JoystickButtonMap[uint8(button)] = InternalButton(internalButton)
		}
	}

	if serializableMapping.JoystickHatMap != nil {
		for hat, button := range serializableMapping.JoystickHatMap {
			mapping.JoystickHatMap[uint8(hat)] = InternalButton(button)
		}
	}

	return mapping, nil
}

func (im *InternalInputMapping) SaveToJSON(filePath string) error {
	serializableMapping := &InputMapping{
		KeyboardMap:         make(map[int]int),
		ControllerButtonMap: make(map[int]int),
		ControllerHatMap:    make(map[int]int),
		JoystickAxisMap: make(map[int]struct {
			PositiveButton int   `json:"positive_button"`
			NegativeButton int   `json:"negative_button"`
			Threshold      int16 `json:"threshold"`
		}),
		JoystickButtonMap: make(map[int]int),
		JoystickHatMap:    make(map[int]int),
	}

	for keyCode, button := range im.KeyboardMap {
		serializableMapping.KeyboardMap[int(keyCode)] = int(button)
	}

	for button, internalButton := range im.ControllerButtonMap {
		serializableMapping.ControllerButtonMap[int(button)] = int(internalButton)
	}

	for hat, button := range im.ControllerHatMap {
		serializableMapping.ControllerHatMap[int(hat)] = int(button)
	}

	for axis, axisMapping := range im.JoystickAxisMap {
		serializableMapping.JoystickAxisMap[int(axis)] = struct {
			PositiveButton int   `json:"positive_button"`
			NegativeButton int   `json:"negative_button"`
			Threshold      int16 `json:"threshold"`
		}{
			PositiveButton: int(axisMapping.PositiveButton),
			NegativeButton: int(axisMapping.NegativeButton),
			Threshold:      axisMapping.Threshold,
		}
	}

	for button, internalButton := range im.JoystickButtonMap {
		serializableMapping.JoystickButtonMap[int(button)] = int(internalButton)
	}

	for hat, button := range im.JoystickHatMap {
		serializableMapping.JoystickHatMap[int(hat)] = int(button)
	}

	data, err := json.MarshalIndent(serializableMapping, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal mapping to JSON: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}
