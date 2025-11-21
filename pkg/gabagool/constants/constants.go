package constants

import (
	"os"
	"time"
)

const Development = "DEV"

const BackgroundPathEnvVar = "BACKGROUND_PATH"

func IsDevMode() bool {
	return os.Getenv("ENVIRONMENT") == Development
}

type VirtualButton int

const (
	VirtualButtonUnassigned VirtualButton = iota
	VirtualButtonUp
	VirtualButtonDown
	VirtualButtonLeft
	VirtualButtonRight
	VirtualButtonA
	VirtualButtonB
	VirtualButtonX
	VirtualButtonY
	VirtualButtonL1
	VirtualButtonL2
	VirtualButtonR1
	VirtualButtonR2
	VirtualButtonStart
	VirtualButtonSelect
	VirtualButtonMenu
	VirtualButtonF1
	VirtualButtonF2
	VirtualButtonVolumeUp
	VirtualButtonVolumeDown
	VirtualButtonPower
)

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
