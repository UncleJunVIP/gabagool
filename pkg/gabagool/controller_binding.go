package gabagool

import (
	"strconv"
	"strings"
)

type DynamicButtonMapping struct {
	ControllerGUID string
	ControllerName string
	ButtonMap      map[Button]uint8 // Maps our Button enum to SDL button indices
	MappingSource  string           // "guid", "name", or "default"
}

var activeMappings map[string]DynamicButtonMapping
var nameMappings map[string]DynamicButtonMapping

func init() {
	activeMappings = make(map[string]DynamicButtonMapping)
	nameMappings = make(map[string]DynamicButtonMapping)
}

func InitializeControllerMapping(controllerName, guid string) {
	var dynamicMapping DynamicButtonMapping
	var mappingFound bool
	var mappingSource string

	// Priority 1: Try to find mapping by GUID
	if guid != "" {
		if mapping, exists := GetControllerMapping(guid); exists {
			dynamicMapping = createMappingFromControllerDB(mapping, guid, controllerName)
			mappingSource = "guid"
			mappingFound = true
			GetLoggerInstance().Info("Found controller mapping by GUID",
				"guid", guid,
				"name", controllerName)
		}
	}

	// Priority 2: Try to find mapping by controller name
	if !mappingFound {
		if mapping := findMappingByName(controllerName); mapping.GUID != "" {
			dynamicMapping = createMappingFromControllerDB(mapping, guid, controllerName)
			mappingSource = "name"
			mappingFound = true
			GetLoggerInstance().Info("Found controller mapping by name",
				"guid", guid,
				"name", controllerName)
		}
	}

	// Priority 3: Use default mapping
	if !mappingFound {
		dynamicMapping = createDefaultMapping(guid, controllerName)
		mappingSource = "default"
		GetLoggerInstance().Info("Using default controller mapping",
			"guid", guid,
			"name", controllerName)
	}

	dynamicMapping.MappingSource = mappingSource

	if guid != "" {
		activeMappings[guid] = dynamicMapping
	}
	activeMappings[controllerName] = dynamicMapping
	nameMappings[controllerName] = dynamicMapping

	GetLoggerInstance().Info("Initialized controller mapping",
		"guid", guid,
		"name", controllerName,
		"source", mappingSource,
		"mappings", len(dynamicMapping.ButtonMap))
}

func findMappingByName(controllerName string) ControllerMapping {
	allMappings := GetAllControllerMappings()

	// First try exact name match
	for _, mapping := range allMappings {
		if strings.EqualFold(mapping.Name, controllerName) {
			return mapping
		}
	}

	// Then try partial name matching (case insensitive)
	controllerNameLower := strings.ToLower(controllerName)
	for _, mapping := range allMappings {
		mappingNameLower := strings.ToLower(mapping.Name)
		if strings.Contains(controllerNameLower, mappingNameLower) ||
			strings.Contains(mappingNameLower, controllerNameLower) {
			return mapping
		}
	}

	return ControllerMapping{}
}

func createMappingFromControllerDB(mapping ControllerMapping, guid, name string) DynamicButtonMapping {
	dynamicMapping := DynamicButtonMapping{
		ControllerGUID: guid,
		ControllerName: name,
		ButtonMap:      make(map[Button]uint8),
	}

	// Map standard gamepad buttons from gamecontrollerdb format
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
	return 255 // Invalid index
}

func GetMappingStatistics() map[string]int {
	stats := map[string]int{
		"guid":    0,
		"name":    0,
		"default": 0,
		"total":   0,
	}

	for _, mapping := range activeMappings {
		stats[mapping.MappingSource]++
		stats["total"]++
	}

	return stats
}
