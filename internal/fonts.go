package internal

import (
	"fmt"
	"github.com/veandco/go-sdl2/ttf"
	"os"
)

var titleFont, mainFont, smallFont *ttf.Font

func InitFonts(titleSize int, fontSize int, smallFontSize int) {
	var err error

	titleFont, err = ttf.OpenFont("fonts/CFPG.ttf", titleSize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load title font: %s\n", err)
		os.Exit(1)
	}

	mainFont, err = ttf.OpenFont("fonts/CFPG.ttf", fontSize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load mmain font: %s\n", err)
		os.Exit(1)
	}

	smallFont, err = ttf.OpenFont("fonts/CFPG.ttf", smallFontSize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load small font: %s\n", err)
		os.Exit(1)
	}

}

func GetTitleFont() *ttf.Font {
	return titleFont
}

func GetFont() *ttf.Font {
	return mainFont
}

func GetSmallFont() *ttf.Font {
	return smallFont
}

func CloseFonts() {
	titleFont.Close()
	mainFont.Close()
	smallFont.Close()
}
