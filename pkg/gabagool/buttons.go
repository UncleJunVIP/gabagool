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
