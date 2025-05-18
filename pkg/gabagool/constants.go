package gabagool

import (
	"os"
	"time"
)

const (
	BrickButton_UP    = 11
	BrickButton_DOWN  = 12
	BrickButton_LEFT  = 13
	BrickButton_RIGHT = 14

	BrickButton_A = 1
	BrickButton_B = 0
	BrickButton_X = 3
	BrickButton_Y = 2

	BrickButton_START  = 6
	BrickButton_SELECT = 4
	BrickButton_MENU   = 5

	BrickButton_F1 = 7
	BrickButton_F2 = 8

	BrickButton_L1 = 9
	BrickButton_R1 = 10
)

const NextUISettingPath = "/mnt/SDCARD/.userdata/shared/minuisettings.txt"
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
	FontSizeLarge  = 16
	FontSizeMedium = 14
	FontSizeSmall  = 12
	FontSizeTiny   = 10
	FontSizeMicro  = 8
)

type TextAlignment int

const (
	AlignLeft TextAlignment = iota
	AlignCenter
	AlignRight
)

const (
	DefaultInputDelay         = 20 * time.Millisecond
	DefaultTitleSpacing int32 = 5
)
