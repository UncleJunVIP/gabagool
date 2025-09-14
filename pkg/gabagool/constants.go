package gabagool

import (
	"os"
	"time"
)

const Development = "DEV"

var IsDev = os.Getenv("ENVIRONMENT") == Development

type TextAlign int

const (
	TextAlignLeft TextAlign = iota
	TextAlignCenter
	TextAlignRight
)

type Button uint8

const (
	ButtonUnassigned Button = 0

	ButtonUp    Button = 11
	ButtonDown  Button = 12
	ButtonLeft  Button = 13
	ButtonRight Button = 14

	ButtonA Button = 1
	ButtonB Button = 0
	ButtonX Button = 3
	ButtonY Button = 2

	ButtonStart  Button = 6
	ButtonSelect Button = 4
	ButtonMenu   Button = 5

	ButtonF1 Button = 7
	ButtonF2 Button = 8

	ButtonL1 Button = 9
	ButtonR1 Button = 10
	ButtonL2        = 15

	ButtonR2 = 16
)

const EnvSettingsFile = "SETTINGS_FILE"

const (
	DefaultWindowWidth  = int32(1024)
	DefaultWindowHeight = int32(768)
)

const (
	DefaultInputDelay         = 20 * time.Millisecond
	DefaultTitleSpacing int32 = 5
)
