package internal

import (
	"fmt"
	"github.com/veandco/go-sdl2/ttf"
	"os"
)

const (
	FontSizeXLarge = 20
	FontSizeLarge  = 16
	FontSizeMedium = 14
	FontSizeSmall  = 12
	FontSizeTiny   = 10
	FontSizeMicro  = 8
)

var xlFont, largeFont, mediumFont, smallFont, tinyFont, microFont,
	largeSymbolFont, mediumSymbolFont, smallSymbolFont, tinySymbolFont, microSymbolFont *ttf.Font

func InitFonts(scale int) {
	xlFont = loadFont("fonts/Rounded_Mplus_1c_Bold.ttf", FontSizeXLarge*scale)
	largeFont = loadFont("fonts/Rounded_Mplus_1c_Bold.ttf", FontSizeLarge*scale)
	mediumFont = loadFont("fonts/Rounded_Mplus_1c_Bold.ttf", FontSizeMedium*scale)
	smallFont = loadFont("fonts/Rounded_Mplus_1c_Bold.ttf", FontSizeSmall*scale)
	tinyFont = loadFont("fonts/Rounded_Mplus_1c_Bold.ttf", FontSizeTiny*scale)
	microFont = loadFont("fonts/Rounded_Mplus_1c_Bold.ttf", FontSizeMicro*scale)

	largeSymbolFont = loadFont("fonts/CFPG.ttf", FontSizeLarge*scale)
	mediumSymbolFont = loadFont("fonts/CFPG.ttf", FontSizeMedium*scale)
	smallSymbolFont = loadFont("fonts/CFPG.ttf", FontSizeSmall*scale)
	tinySymbolFont = loadFont("fonts/CFPG.ttf", FontSizeTiny*scale)
	microSymbolFont = loadFont("fonts/CFPG.ttf", FontSizeMicro*scale)
}

func loadFont(path string, size int) *ttf.Font {
	font, err := ttf.OpenFont(path, size)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load font: %s\n", err)
		os.Exit(1)
	}

	return font
}

func GetXLargeFont() *ttf.Font {
	return xlFont
}

func GetLargeFont() *ttf.Font {
	return largeFont
}

func GetMediumFont() *ttf.Font {
	return mediumFont
}

func GetSmallFont() *ttf.Font {
	return smallFont
}

func GetTinyFont() *ttf.Font {
	return tinyFont
}

func GetMicroFont() *ttf.Font {
	return microFont
}

func GetLargeSymbolFont() *ttf.Font {
	return largeSymbolFont
}

func GetMediumSymbolFont() *ttf.Font {
	return mediumSymbolFont
}

func GetSmallSymbolFont() *ttf.Font {
	return smallSymbolFont
}

func GetTinySymbolFont() *ttf.Font {
	return tinySymbolFont
}

func GetMicroSymbolFont() *ttf.Font {
	return microSymbolFont
}

func CloseFonts() {
	largeFont.Close()
	mediumFont.Close()
	smallFont.Close()
	tinyFont.Close()
	microFont.Close()

	largeSymbolFont.Close()
	mediumSymbolFont.Close()
	smallSymbolFont.Close()
	tinySymbolFont.Close()
	microSymbolFont.Close()
}
