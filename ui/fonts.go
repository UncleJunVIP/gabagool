package ui

import (
	"fmt"
	"github.com/veandco/go-sdl2/ttf"
	"os"
)

var font, smallFont *ttf.Font

func InitFonts(fontSize int, smallFontSize int) {
	var err error

	font, err = ttf.OpenFont("BPreplayBold.ttf", fontSize)
	if err != nil {
		fmt.Printf("Warning: Failed to load font, using system font: %s\n", err)
		font, err = ttf.OpenFont("DejaVuSans.ttf", fontSize)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load fallback font: %s\n", err)
			os.Exit(1)
		}
	}

	smallFont, err = ttf.OpenFont("BPreplay.ttf", smallFontSize)
	if err != nil {
		smallFont, err = ttf.OpenFont("DejaVuSans.ttf", smallFontSize)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load fallback small font: %s\n", err)
			os.Exit(1)
		}
	}
}

func GetFont() *ttf.Font {
	return font
}

func GetSmallFont() *ttf.Font {
	return smallFont
}

func CloseFonts() {
	font.Close()
	smallFont.Close()
}
