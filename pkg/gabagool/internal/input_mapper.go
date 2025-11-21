package internal

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

const MappingPathEnvVar = "INPUT_MAPPING_PATH"

type VirtualButton int

const (
	VirtualButtonUnassigned VirtualButton = iota
	VirtualButtonUp
	VirtualButtonDown
	VirtualButtonLeft
	VirtualButtonRight
	VirtualButtonA
	VirtualButtonB
	VirtualButtonX
	VirtualButtonY
	VirtualButtonL1
	VirtualButtonL2
	VirtualButtonR1
	VirtualButtonR2
	VirtualButtonStart
	VirtualButtonSelect
	VirtualButtonMenu
	VirtualButtonF1
	VirtualButtonF2
	VirtualButtonVolumeUp
	VirtualButtonVolumeDown
	VirtualButtonPower
)

type Source int

const (
	SourceKeyboard Source = iota
	SourceController
	SourceJoystick
	SourceHatSwitch
)

type Event struct {
	Button  VirtualButton
	Pressed bool
	Source  Source
	RawCode int
}

type JoystickAxisMapping struct {
	PositiveButton VirtualButton
	NegativeButton VirtualButton
	Threshold      int16
}

type InternalInputMapping struct {
	KeyboardMap map[sdl.Keycode]VirtualButton

	ControllerButtonMap map[sdl.GameControllerButton]VirtualButton

	ControllerHatMap map[uint8]VirtualButton

	JoystickAxisMap map[uint8]JoystickAxisMapping

	JoystickButtonMap map[uint8]VirtualButton

	JoystickHatMap map[uint8]VirtualButton
}

type Mapping struct {
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
		KeyboardMap: map[sdl.Keycode]VirtualButton{
			sdl.K_UP:     VirtualButtonUp,
			sdl.K_DOWN:   VirtualButtonDown,
			sdl.K_LEFT:   VirtualButtonLeft,
			sdl.K_RIGHT:  VirtualButtonRight,
			sdl.K_a:      VirtualButtonA,
			sdl.K_b:      VirtualButtonB,
			sdl.K_x:      VirtualButtonX,
			sdl.K_y:      VirtualButtonY,
			sdl.K_RETURN: VirtualButtonStart,
			sdl.K_SPACE:  VirtualButtonSelect,
			sdl.K_h:      VirtualButtonMenu,
		},
		ControllerButtonMap: map[sdl.GameControllerButton]VirtualButton{
			sdl.CONTROLLER_BUTTON_DPAD_UP:    VirtualButtonUp,
			sdl.CONTROLLER_BUTTON_DPAD_DOWN:  VirtualButtonDown,
			sdl.CONTROLLER_BUTTON_DPAD_LEFT:  VirtualButtonLeft,
			sdl.CONTROLLER_BUTTON_DPAD_RIGHT: VirtualButtonRight,
			sdl.CONTROLLER_BUTTON_A:          VirtualButtonB,
			sdl.CONTROLLER_BUTTON_B:          VirtualButtonA,
			sdl.CONTROLLER_BUTTON_X:          VirtualButtonY,
			sdl.CONTROLLER_BUTTON_Y:          VirtualButtonX,
			sdl.CONTROLLER_BUTTON_START:      VirtualButtonStart,
			sdl.CONTROLLER_BUTTON_BACK:       VirtualButtonSelect,
			sdl.CONTROLLER_BUTTON_GUIDE:      VirtualButtonMenu,
		},
	}
}

// GetInputMapping returns the input mapping from the environment variable if set,
// otherwise returns the default mapping
func GetInputMapping() *InternalInputMapping {
	logger := GetLogger()
	mappingPath := os.Getenv(MappingPathEnvVar)
	if mappingPath != "" {
		mapping, err := LoadInputMappingFromJSON(mappingPath)
		if err == nil {
			logger.Info("Loaded custom input mapping from environment variable", "path", mappingPath)
			return mapping
		}
		logger.Warn("Failed to load custom input mapping, using default", "path", mappingPath, "error", err)
	}
	return DefaultInputMapping()
}

func LoadInputMappingFromJSON(filePath string) (*InternalInputMapping, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %w", err)
	}

	var serializableMapping Mapping
	err = json.Unmarshal(data, &serializableMapping)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	mapping := &InternalInputMapping{
		KeyboardMap:         make(map[sdl.Keycode]VirtualButton),
		ControllerButtonMap: make(map[sdl.GameControllerButton]VirtualButton),
		ControllerHatMap:    make(map[uint8]VirtualButton),
		JoystickAxisMap:     make(map[uint8]JoystickAxisMapping),
		JoystickButtonMap:   make(map[uint8]VirtualButton),
		JoystickHatMap:      make(map[uint8]VirtualButton),
	}

	if serializableMapping.KeyboardMap != nil {
		for keyCode, button := range serializableMapping.KeyboardMap {
			mapping.KeyboardMap[sdl.Keycode(keyCode)] = VirtualButton(button)
		}
	}

	if serializableMapping.ControllerButtonMap != nil {
		for button, vb := range serializableMapping.ControllerButtonMap {
			mapping.ControllerButtonMap[sdl.GameControllerButton(button)] = VirtualButton(vb)
		}
	}

	if serializableMapping.ControllerHatMap != nil {
		for hat, button := range serializableMapping.ControllerHatMap {
			mapping.ControllerHatMap[uint8(hat)] = VirtualButton(button)
		}
	}

	if serializableMapping.JoystickAxisMap != nil {
		for axis, axisMapping := range serializableMapping.JoystickAxisMap {
			mapping.JoystickAxisMap[uint8(axis)] = JoystickAxisMapping{
				PositiveButton: VirtualButton(axisMapping.PositiveButton),
				NegativeButton: VirtualButton(axisMapping.NegativeButton),
				Threshold:      axisMapping.Threshold,
			}
		}
	}

	if serializableMapping.JoystickButtonMap != nil {
		for button, vb := range serializableMapping.JoystickButtonMap {
			mapping.JoystickButtonMap[uint8(button)] = VirtualButton(vb)
		}
	}

	if serializableMapping.JoystickHatMap != nil {
		for hat, button := range serializableMapping.JoystickHatMap {
			mapping.JoystickHatMap[uint8(hat)] = VirtualButton(button)
		}
	}

	return mapping, nil
}

func (im *InternalInputMapping) SaveToJSON(filePath string) error {
	serializableMapping := &Mapping{
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

	for button, VirtualButton := range im.ControllerButtonMap {
		serializableMapping.ControllerButtonMap[int(button)] = int(VirtualButton)
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

	for button, VirtualButton := range im.JoystickButtonMap {
		serializableMapping.JoystickButtonMap[int(button)] = int(VirtualButton)
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
