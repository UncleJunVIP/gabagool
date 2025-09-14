package gabagool

type Button uint8

var (
	ButtonUnassigned Button = 0

	// D-Pad
	ButtonUp    Button = 11
	ButtonDown  Button = 12
	ButtonLeft  Button = 13
	ButtonRight Button = 14

	// Face buttons
	ButtonA Button = 1
	ButtonB Button = 0
	ButtonX Button = 3
	ButtonY Button = 2

	// Control buttons
	ButtonStart  Button = 6
	ButtonSelect Button = 4
	ButtonMenu   Button = 5

	// Function buttons
	ButtonF1 Button = 7
	ButtonF2 Button = 8

	// Shoulder buttons
	ButtonL1 Button = 9
	ButtonR1 Button = 10
	ButtonL2 Button = 15
	ButtonR2 Button = 16
)

var defaultButtonValues = map[string]Button{
	"ButtonUp":     11,
	"ButtonDown":   12,
	"ButtonLeft":   13,
	"ButtonRight":  14,
	"ButtonA":      1,
	"ButtonB":      0,
	"ButtonX":      3,
	"ButtonY":      2,
	"ButtonStart":  6,
	"ButtonSelect": 4,
	"ButtonMenu":   5,
	"ButtonF1":     7,
	"ButtonF2":     8,
	"ButtonL1":     9,
	"ButtonR1":     10,
	"ButtonL2":     15,
	"ButtonR2":     16,
}

func UpdateButtonMapping(mapping map[string]Button) {
	if val, exists := mapping["ButtonUp"]; exists {
		ButtonUp = val
	}
	if val, exists := mapping["ButtonDown"]; exists {
		ButtonDown = val
	}
	if val, exists := mapping["ButtonLeft"]; exists {
		ButtonLeft = val
	}
	if val, exists := mapping["ButtonRight"]; exists {
		ButtonRight = val
	}
	if val, exists := mapping["ButtonA"]; exists {
		ButtonA = val
	}
	if val, exists := mapping["ButtonB"]; exists {
		ButtonB = val
	}
	if val, exists := mapping["ButtonX"]; exists {
		ButtonX = val
	}
	if val, exists := mapping["ButtonY"]; exists {
		ButtonY = val
	}
	if val, exists := mapping["ButtonStart"]; exists {
		ButtonStart = val
	}
	if val, exists := mapping["ButtonSelect"]; exists {
		ButtonSelect = val
	}
	if val, exists := mapping["ButtonMenu"]; exists {
		ButtonMenu = val
	}
	if val, exists := mapping["ButtonF1"]; exists {
		ButtonF1 = val
	}
	if val, exists := mapping["ButtonF2"]; exists {
		ButtonF2 = val
	}
	if val, exists := mapping["ButtonL1"]; exists {
		ButtonL1 = val
	}
	if val, exists := mapping["ButtonR1"]; exists {
		ButtonR1 = val
	}
	if val, exists := mapping["ButtonL2"]; exists {
		ButtonL2 = val
	}
	if val, exists := mapping["ButtonR2"]; exists {
		ButtonR2 = val
	}
}

func ResetButtonMappingToDefaults() {
	ButtonUp = defaultButtonValues["ButtonUp"]
	ButtonDown = defaultButtonValues["ButtonDown"]
	ButtonLeft = defaultButtonValues["ButtonLeft"]
	ButtonRight = defaultButtonValues["ButtonRight"]
	ButtonA = defaultButtonValues["ButtonA"]
	ButtonB = defaultButtonValues["ButtonB"]
	ButtonX = defaultButtonValues["ButtonX"]
	ButtonY = defaultButtonValues["ButtonY"]
	ButtonStart = defaultButtonValues["ButtonStart"]
	ButtonSelect = defaultButtonValues["ButtonSelect"]
	ButtonMenu = defaultButtonValues["ButtonMenu"]
	ButtonF1 = defaultButtonValues["ButtonF1"]
	ButtonF2 = defaultButtonValues["ButtonF2"]
	ButtonL1 = defaultButtonValues["ButtonL1"]
	ButtonR1 = defaultButtonValues["ButtonR1"]
	ButtonL2 = defaultButtonValues["ButtonL2"]
	ButtonR2 = defaultButtonValues["ButtonR2"]
}

func GetCurrentButtonMapping() map[string]Button {
	return map[string]Button{
		"ButtonUp":     ButtonUp,
		"ButtonDown":   ButtonDown,
		"ButtonLeft":   ButtonLeft,
		"ButtonRight":  ButtonRight,
		"ButtonA":      ButtonA,
		"ButtonB":      ButtonB,
		"ButtonX":      ButtonX,
		"ButtonY":      ButtonY,
		"ButtonStart":  ButtonStart,
		"ButtonSelect": ButtonSelect,
		"ButtonMenu":   ButtonMenu,
		"ButtonF1":     ButtonF1,
		"ButtonF2":     ButtonF2,
		"ButtonL1":     ButtonL1,
		"ButtonR1":     ButtonR1,
		"ButtonL2":     ButtonL2,
		"ButtonR2":     ButtonR2,
	}
}

func ApplyBestControllerMapping() {
	stats := GetMappingStatistics()
	if stats["total"] == 0 {
		GetLoggerInstance().Info("No controller mappings available, using defaults")
		return
	}

	// Find the best mapping to use (prioritize GUID, then name, then default)
	var bestMapping DynamicButtonMapping
	var found bool

	// Priority 1: Look for a GUID-based mapping
	for _, mapping := range activeMappings {
		if mapping.MappingSource == "guid" {
			bestMapping = mapping
			found = true
			break
		}
	}

	// Priority 2: Look for a name-based mapping
	if !found {
		for _, mapping := range activeMappings {
			if mapping.MappingSource == "name" {
				bestMapping = mapping
				found = true
				break
			}
		}
	}

	// Priority 3: Use any available mapping
	if !found {
		for _, mapping := range activeMappings {
			bestMapping = mapping
			found = true
			break
		}
	}

	if !found {
		GetLoggerInstance().Info("No controller mappings to apply")
		return
	}

	// Convert the button mapping to string-based mapping for UpdateButtonMapping
	stringMapping := make(map[string]Button)

	for ourButton, sdlButton := range bestMapping.ButtonMap {
		buttonName := getButtonConstantName(ourButton)
		if buttonName != "" {
			stringMapping[buttonName] = Button(sdlButton)
		}
	}

	// Apply the mapping to global constants
	UpdateButtonMapping(stringMapping)

	GetLoggerInstance().Info("Applied controller mapping to global constants",
		"source", bestMapping.MappingSource,
		"controller", bestMapping.ControllerName,
		"mappings_applied", len(stringMapping))
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
