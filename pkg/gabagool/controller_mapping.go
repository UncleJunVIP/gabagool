package gabagool

import (
	_ "embed"
	"strings"
)

//go:embed resources/gamecontrollerdb.txt
var gameControllerDB string

type ControllerMapping struct {
	GUID     string
	Name     string
	Platform string
	Mappings map[string]string
}

var controllerMappings map[string]ControllerMapping

func init() {
	loadControllerMappings()
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

	GetLoggerInstance().Info("Loaded controller mappings", "count", len(controllerMappings))
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
		Platform: "Linux", // Default platform
		Mappings: make(map[string]string),
	}

	for i := 2; i < len(parts); i++ {
		part := strings.TrimSpace(parts[i])
		if strings.Contains(part, ":") {
			keyValue := strings.SplitN(part, ":", 2)
			if len(keyValue) == 2 {
				key := strings.TrimSpace(keyValue[0])
				value := strings.TrimSpace(keyValue[1])

				// Handle platform specification
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

func GetControllerMapping(guid string) (ControllerMapping, bool) {
	mapping, exists := controllerMappings[guid]
	return mapping, exists
}

func GetAllControllerMappings() map[string]ControllerMapping {
	return controllerMappings
}
