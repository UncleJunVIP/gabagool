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

func HexToColor(hex uint32) sdl.Color {
	r := uint8((hex >> 16) & 0xFF)
	g := uint8((hex >> 8) & 0xFF)
	b := uint8(hex & 0xFF)

	return sdl.Color{R: r, G: g, B: b, A: 255}
}
