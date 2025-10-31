// controller_config.go
package gabagool

import (
	_ "embed"
	"encoding/json"
	"os"
	"strconv"
	"strings"
)

//go:embed resources/gamecontrollerdb.txt
var gameControllerDB string

type ButtonConfig struct {
	Up     uint8 `json:"up"`
	Down   uint8 `json:"down"`
	Left   uint8 `json:"left"`
	Right  uint8 `json:"right"`
	A      uint8 `json:"a"`
	B      uint8 `json:"b"`
	X      uint8 `json:"x"`
	Y      uint8 `json:"y"`
	Start  uint8 `json:"start"`
	Select uint8 `json:"select"`
	Menu   uint8 `json:"menu"`
	F1     uint8 `json:"f1"`
	F2     uint8 `json:"f2"`
	L1     uint8 `json:"l1"`
	R1     uint8 `json:"r1"`
	L2     uint8 `json:"l2"`
	R2     uint8 `json:"r2"`
}

type ControllerProfile struct {
	GUID     string       `json:"guid,omitempty"`
	Name     string       `json:"name,omitempty"`
	Platform string       `json:"platform,omitempty"`
	Buttons  ButtonConfig `json:"buttons"`
}

type ControllerConfigFile struct {
	Profiles []ControllerProfile `json:"profiles"`
}

type ControllerMapping struct {
	GUID     string
	Name     string
	Platform string
	Mappings map[string]string
}

type DynamicButtonMapping struct {
	ControllerGUID string
	ControllerName string
	ButtonMap      map[Button]uint8
	MappingSource  string // "config", "database", or "default"
}

var (
	controllerMappings map[string]ControllerMapping
	activeMappings     map[string]DynamicButtonMapping
	configFile         *ControllerConfigFile
)

func init() {
	controllerMappings = make(map[string]ControllerMapping)
	activeMappings = make(map[string]DynamicButtonMapping)
	loadControllerMappings()
}

func LoadControllerConfiguration(configPath string) error {
	if configPath == "" {
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		GetLoggerInstance().Error("Failed to read controller config file", "path", configPath, "error", err)
		return err
	}

	configFile = &ControllerConfigFile{}
	if err := json.Unmarshal(data, configFile); err != nil {
		GetLoggerInstance().Error("Failed to parse controller config file", "path", configPath, "error", err)
		return err
	}

	GetLoggerInstance().Debug("Loaded controller configuration", "profiles", len(configFile.Profiles))
	return nil
}

func InitializeControllerMapping(controllerName, guid string) {
	var mapping DynamicButtonMapping
	var source string

	// Priority 1: Check config file
	if configFile != nil {
		mapping, source = findConfigMapping(controllerName, guid)
		if source != "" {
			mapping.MappingSource = source
			activeMappings[guid] = mapping
			activeMappings[controllerName] = mapping
			GetLoggerInstance().Debug("Applied config file mapping",
				"controller", controllerName,
				"guid", guid,
				"source", source)
			return
		}
	}

	// Priority 2: Check embedded database
	if dbMapping := findDatabaseMapping(controllerName); dbMapping.GUID != "" {
		mapping = createMappingFromDatabase(dbMapping, guid, controllerName)
		mapping.MappingSource = "database"
		activeMappings[guid] = mapping
		activeMappings[controllerName] = mapping
		GetLoggerInstance().Debug("Applied database mapping",
			"controller", controllerName,
			"guid", guid)
		return
	}

	// Priority 3: Use defaults
	mapping = createDefaultMapping(guid, controllerName)
	mapping.MappingSource = "default"
	activeMappings[guid] = mapping
	activeMappings[controllerName] = mapping
	GetLoggerInstance().Debug("Applied default mapping",
		"controller", controllerName,
		"guid", guid)
}

func findConfigMapping(controllerName, guid string) (DynamicButtonMapping, string) {
	if configFile == nil || len(configFile.Profiles) == 0 {
		return DynamicButtonMapping{}, ""
	}

	// First try exact GUID match
	for _, profile := range configFile.Profiles {
		if profile.GUID != "" && profile.GUID == guid {
			return profileToMapping(profile), "config"
		}
	}

	// Then try name match (case-insensitive)
	controllerNameLower := strings.ToLower(controllerName)
	for _, profile := range configFile.Profiles {
		if profile.Name != "" && strings.EqualFold(profile.Name, controllerName) {
			return profileToMapping(profile), "config"
		}

		// Partial match
		if profile.Name != "" && (strings.Contains(controllerNameLower, strings.ToLower(profile.Name)) ||
			strings.Contains(strings.ToLower(profile.Name), controllerNameLower)) {
			return profileToMapping(profile), "config"
		}
	}

	return DynamicButtonMapping{}, ""
}

func profileToMapping(profile ControllerProfile) DynamicButtonMapping {
	return DynamicButtonMapping{
		ControllerGUID: profile.GUID,
		ControllerName: profile.Name,
		ButtonMap: map[Button]uint8{
			ButtonUp:     profile.Buttons.Up,
			ButtonDown:   profile.Buttons.Down,
			ButtonLeft:   profile.Buttons.Left,
			ButtonRight:  profile.Buttons.Right,
			ButtonA:      profile.Buttons.A,
			ButtonB:      profile.Buttons.B,
			ButtonX:      profile.Buttons.X,
			ButtonY:      profile.Buttons.Y,
			ButtonStart:  profile.Buttons.Start,
			ButtonSelect: profile.Buttons.Select,
			ButtonMenu:   profile.Buttons.Menu,
			ButtonF1:     profile.Buttons.F1,
			ButtonF2:     profile.Buttons.F2,
			ButtonL1:     profile.Buttons.L1,
			ButtonR1:     profile.Buttons.R1,
			ButtonL2:     profile.Buttons.L2,
			ButtonR2:     profile.Buttons.R2,
		},
	}
}

func findDatabaseMapping(controllerName string) ControllerMapping {
	// First try exact name match
	for _, mapping := range controllerMappings {
		if strings.EqualFold(mapping.Name, controllerName) {
			return mapping
		}
	}

	// Then try partial name matching (case insensitive)
	controllerNameLower := strings.ToLower(controllerName)
	for _, mapping := range controllerMappings {
		mappingNameLower := strings.ToLower(mapping.Name)
		if strings.Contains(controllerNameLower, mappingNameLower) ||
			strings.Contains(mappingNameLower, controllerNameLower) {
			return mapping
		}
	}

	return ControllerMapping{}
}

func createMappingFromDatabase(mapping ControllerMapping, guid, name string) DynamicButtonMapping {
	dynamicMapping := DynamicButtonMapping{
		ControllerGUID: guid,
		ControllerName: name,
		ButtonMap:      make(map[Button]uint8),
	}

	buttonMappings := map[string]Button{
		"a":             ButtonA,
		"b":             ButtonB,
		"x":             ButtonX,
		"y":             ButtonY,
		"back":          ButtonSelect,
		"guide":         ButtonMenu,
		"start":         ButtonStart,
		"leftshoulder":  ButtonL1,
		"rightshoulder": ButtonR1,
		"lefttrigger":   ButtonL2,
		"righttrigger":  ButtonR2,
		"dpup":          ButtonUp,
		"dpdown":        ButtonDown,
		"dpleft":        ButtonLeft,
		"dpright":       ButtonRight,
	}

	for gamepadButton, ourButton := range buttonMappings {
		if sdlMapping, exists := mapping.Mappings[gamepadButton]; exists {
			if sdlIndex := parseSDLButtonIndex(sdlMapping); sdlIndex != 255 {
				dynamicMapping.ButtonMap[ourButton] = sdlIndex
			}
		}
	}

	return dynamicMapping
}

func createDefaultMapping(guid, name string) DynamicButtonMapping {
	return DynamicButtonMapping{
		ControllerGUID: guid,
		ControllerName: name,
		ButtonMap: map[Button]uint8{
			ButtonA:      1,
			ButtonB:      0,
			ButtonX:      3,
			ButtonY:      2,
			ButtonSelect: 4,
			ButtonMenu:   5,
			ButtonStart:  6,
			ButtonL1:     9,
			ButtonR1:     10,
			ButtonL2:     15,
			ButtonR2:     16,
			ButtonUp:     11,
			ButtonDown:   12,
			ButtonLeft:   13,
			ButtonRight:  14,
			ButtonF1:     7,
			ButtonF2:     8,
		},
	}
}

func parseSDLButtonIndex(sdlMapping string) uint8 {
	if strings.HasPrefix(sdlMapping, "b") && len(sdlMapping) > 1 {
		if index, err := strconv.Atoi(sdlMapping[1:]); err == nil && index >= 0 && index <= 255 {
			return uint8(index)
		}
	}
	return 255
}

func loadControllerMappings() {
	controllerMappings = make(map[string]ControllerMapping)

	lines := strings.Split(gameControllerDB, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		mapping := parseControllerLine(line)
		if mapping.GUID != "" {
			controllerMappings[mapping.GUID] = mapping
		}
	}

	GetLoggerInstance().Debug("Loaded controller mappings", "count", len(controllerMappings))
}

func parseControllerLine(line string) ControllerMapping {
	parts := strings.Split(line, ",")
	if len(parts) < 3 {
		return ControllerMapping{}
	}

	guid := parts[0]
	name := parts[1]

	mapping := ControllerMapping{
		GUID:     guid,
		Name:     name,
		Platform: "Linux",
		Mappings: make(map[string]string),
	}

	for i := 2; i < len(parts); i++ {
		part := strings.TrimSpace(parts[i])
		if strings.Contains(part, ":") {
			keyValue := strings.SplitN(part, ":", 2)
			if len(keyValue) == 2 {
				key := strings.TrimSpace(keyValue[0])
				value := strings.TrimSpace(keyValue[1])

				if key == "platform" {
					mapping.Platform = value
				} else {
					mapping.Mappings[key] = value
				}
			}
		}
	}

	return mapping
}

func ApplyBestControllerMapping() {
	if len(activeMappings) == 0 {
		GetLoggerInstance().Info("No controller mappings available, using defaults")
		return
	}

	// Find the best mapping (prioritize config, then database, then default)
	var bestMapping DynamicButtonMapping
	var found bool

	for _, mapping := range activeMappings {
		if mapping.MappingSource == "config" {
			bestMapping = mapping
			found = true
			break
		}
	}

	if !found {
		for _, mapping := range activeMappings {
			if mapping.MappingSource == "database" {
				bestMapping = mapping
				found = true
				break
			}
		}
	}

	if !found {
		for _, mapping := range activeMappings {
			bestMapping = mapping
			found = true
			break
		}
	}

	if !found {
		return
	}

	stringMapping := make(map[string]Button)
	for ourButton, sdlButton := range bestMapping.ButtonMap {
		buttonName := getButtonConstantName(ourButton)
		if buttonName != "" {
			stringMapping[buttonName] = Button(sdlButton)
		}
	}

	UpdateButtonMapping(stringMapping)

	GetLoggerInstance().Debug("Applied controller mapping to global constants",
		"source", bestMapping.MappingSource,
		"controller", bestMapping.ControllerName,
		"mappings_applied", len(stringMapping))
}

func GetMappingStatistics() map[string]int {
	stats := map[string]int{
		"config":   0,
		"database": 0,
		"default":  0,
		"total":    0,
	}

	for _, mapping := range activeMappings {
		stats[mapping.MappingSource]++
		stats["total"]++
	}

	return stats
}

func GetAllControllerMappings() map[string]ControllerMapping {
	return controllerMappings
}

func getButtonConstantName(button Button) string {
	switch button {
	case ButtonA:
		return "ButtonA"
	case ButtonB:
		return "ButtonB"
	case ButtonX:
		return "ButtonX"
	case ButtonY:
		return "ButtonY"
	case ButtonSelect:
		return "ButtonSelect"
	case ButtonMenu:
		return "ButtonMenu"
	case ButtonStart:
		return "ButtonStart"
	case ButtonL1:
		return "ButtonL1"
	case ButtonR1:
		return "ButtonR1"
	case ButtonL2:
		return "ButtonL2"
	case ButtonR2:
		return "ButtonR2"
	case ButtonUp:
		return "ButtonUp"
	case ButtonDown:
		return "ButtonDown"
	case ButtonLeft:
		return "ButtonLeft"
	case ButtonRight:
		return "ButtonRight"
	case ButtonF1:
		return "ButtonF1"
	case ButtonF2:
		return "ButtonF2"
	default:
		return ""
	}
}
