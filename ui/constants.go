package ui

import "time"

type TextAlignment int

const (
	AlignLeft TextAlignment = iota
	AlignCenter
	AlignRight
)

const (
	DefaultMenuSpacing  int32 = 10
	DefaultMenuXMargin  int32 = 15 // Left margin for the menu items
	DefaultMenuYMargin  int32 = 5  // Top/bottom margin within each menu item
	DefaultTextPadding  int32 = 15 // Padding around text in the pill
	DefaultInputDelay         = 20 * time.Millisecond
	DefaultTitleSpacing int32 = 30 // Space between title and first menu item
	DefaultTitleXMargin int32 = 20
)

func DefaultMenuSettings(title string) MenuSettings {
	return MenuSettings{
		Spacing:      DefaultMenuSpacing,
		XMargin:      DefaultMenuXMargin,
		YMargin:      DefaultMenuYMargin,
		TextXPad:     DefaultTextPadding,
		TextYPad:     DefaultTextPadding,
		InputDelay:   DefaultInputDelay,
		Title:        title,
		TitleAlign:   AlignLeft,
		TitleSpacing: DefaultTitleSpacing,
		TitleXMargin: DefaultTitleXMargin,
	}
}
