package gabagool

import (
	"fmt"
	"github.com/veandco/go-sdl2/ttf"
	"os"
)

const fallbackFont = "/Users/btk/Library/Fonts/Rounded MPlus 1c Bold.ttf"

var fonts fontsManager

type fontsManager struct {
	extraLargeFont *ttf.Font
	largeFont      *ttf.Font
	mediumFont     *ttf.Font
	smallFont      *ttf.Font
	tinyFont       *ttf.Font
	microFont      *ttf.Font

	largeSymbolFont  *ttf.Font
	mediumSymbolFont *ttf.Font
	smallSymbolFont  *ttf.Font
	tinySymbolFont   *ttf.Font
	microSymbolFont  *ttf.Font
}

func initFonts(scale int) {
	xlFont := loadFont(currentTheme.FontPath, fallbackFont, FontSizeXLarge*scale)
	largeFont := loadFont(currentTheme.FontPath, fallbackFont, FontSizeLarge*scale)
	mediumFont := loadFont(currentTheme.FontPath, fallbackFont, FontSizeMedium*scale)
	smallFont := loadFont(currentTheme.FontPath, fallbackFont, FontSizeSmall*scale)
	tinyFont := loadFont(currentTheme.FontPath, fallbackFont, FontSizeTiny*scale)
	microFont := loadFont(currentTheme.FontPath, fallbackFont, FontSizeMicro*scale)

	largeSymbolFont := loadFont("/mnt/SDCARD/.system/res/font1.ttf", fallbackFont, FontSizeLarge*scale)
	mediumSymbolFont := loadFont("/mnt/SDCARD/.system/res/font1.ttf", fallbackFont, FontSizeMedium*scale)
	smallSymbolFont := loadFont("/mnt/SDCARD/.system/res/font1.ttf", fallbackFont, FontSizeSmall*scale)
	tinySymbolFont := loadFont("/mnt/SDCARD/.system/res/font1.ttf", fallbackFont, FontSizeTiny*scale)
	microSymbolFont := loadFont("/mnt/SDCARD/.system/res/font1.ttf", fallbackFont, FontSizeMicro*scale)

	fonts = fontsManager{
		extraLargeFont: xlFont,
		largeFont:      largeFont,
		mediumFont:     mediumFont,
		smallFont:      smallFont,
		tinyFont:       tinyFont,
		microFont:      microFont,

		largeSymbolFont:  largeSymbolFont,
		mediumSymbolFont: mediumSymbolFont,
		smallSymbolFont:  smallSymbolFont,
		tinySymbolFont:   tinySymbolFont,
		microSymbolFont:  microSymbolFont,
	}
}

func loadFont(path string, fallback string, size int) *ttf.Font {
	font, err := ttf.OpenFont(path, size)
	if err != nil && fallback == "" {
		fmt.Fprintf(os.Stderr, "Failed to load font: %s\n", err)
		os.Exit(1)
	} else if err != nil {
		font, err = ttf.OpenFont(fallback, size)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fallback font: %s\n", err)
			os.Exit(1)
		}
	}

	return font
}

func closeFonts() {
	fonts.largeFont.Close()
	fonts.mediumFont.Close()
	fonts.smallFont.Close()
	fonts.tinyFont.Close()
	fonts.microFont.Close()

	fonts.largeSymbolFont.Close()
	fonts.mediumSymbolFont.Close()
	fonts.smallSymbolFont.Close()
	fonts.tinySymbolFont.Close()
	fonts.microSymbolFont.Close()
}
