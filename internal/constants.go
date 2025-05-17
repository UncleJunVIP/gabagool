package internal

import (
	_ "github.com/UncleJunVIP/certifiable"
	"time"
)

const nextUISettingPath = "/mnt/SDCARD/.userdata/shared/minuisettings.txt"
const nextUIBackgroundPath = "/mnt/SDCARD/bg.png"

const envSettingsFile = "SETTINGS_FILE"
const envBackgroundPath = "BACKGROUND_PATH"

const Development = "DEV"

const (
	DefaultWindowWidth  = int32(1024)
	DefaultWindowHeight = int32(768)
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
