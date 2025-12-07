package internal

import (
	"github.com/veandco/go-sdl2/sdl"
)

type Theme struct {
	MainColor             sdl.Color
	PrimaryAccentColor    sdl.Color
	SecondaryAccentColor  sdl.Color
	HintInfoColor         sdl.Color
	ListTextColor         sdl.Color
	ListTextSelectedColor sdl.Color
	BGColor               sdl.Color
	FontPath              string
	BackgroundImagePath   string
}

var currentTheme Theme

func SetTheme(theme Theme) {
	currentTheme = theme
}

func GetTheme() Theme {
	return currentTheme
}
