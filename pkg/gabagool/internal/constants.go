package internal

import (
	"os"
	"time"
)

const Development = "DEV"

const BackgroundPathEnvVar = "BACKGROUND_PATH"

func IsDevMode() bool {
	return os.Getenv("ENVIRONMENT") == Development
}

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
