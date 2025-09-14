package gabagool

import (
	"os"
	"time"
)

const Development = "DEV"

var IsDev = os.Getenv("ENVIRONMENT") == Development

const EnvSettingsFile = "SETTINGS_FILE"

type TextAlign int

const (
	TextAlignLeft TextAlign = iota
	TextAlignCenter
	TextAlignRight
)

const (
	DefaultInputDelay         = 20 * time.Millisecond
	DefaultTitleSpacing int32 = 5
)
