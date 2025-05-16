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
	DefaultInputDelay         = 20 * time.Millisecond
	DefaultTitleSpacing int32 = 5
)
