package internal

import (
	_ "github.com/UncleJunVIP/certifiable"
	"time"
)

type TextAlignment int

const (
	AlignLeft TextAlignment = iota
	AlignCenter
	AlignRight
)

const (
	DefaultMenuSpacing  int32 = 10
	DefaultTextPadding  int32 = 15 // Padding around text in the pill
	DefaultInputDelay         = 20 * time.Millisecond
	DefaultTitleSpacing int32 = 30 // Space between title and first menu item
)
