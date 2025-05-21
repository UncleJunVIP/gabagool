package gabagool

import (
	"os"
	"time"
)

type Button uint8

const (
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
)

const NextUIBackgroundPath = "/mnt/SDCARD/bg.png"

const EnvSettingsFile = "SETTINGS_FILE"
const EnvBackgroundPath = "BACKGROUND_PATH"

var IsDev = os.Getenv("ENVIRONMENT") == Development

const Development = "DEV"

const (
	DefaultWindowWidth  = int32(1024)
	DefaultWindowHeight = int32(768)
)

const (
	FontSizeXLarge = 20
	FontSizeLarge  = 17
	FontSizeMedium = 15
	FontSizeSmall  = 13
	FontSizeTiny   = 9
	FontSizeMicro  = 7
)

type TextAlign int

const (
	AlignLeft TextAlign = iota
	AlignCenter
	AlignRight
)

const (
	DefaultInputDelay         = 20 * time.Millisecond
	DefaultTitleSpacing int32 = 5
)
