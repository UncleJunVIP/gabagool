package gabagool

import (
	"fmt"
	"github.com/veandco/go-sdl2/ttf"
	"os"
)

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
	xlFont := loadFont(currentTheme.FontPath, "fonts/Rounded_Mplus_1c_Bold.ttf", FontSizeXLarge*scale)
	largeFont := loadFont(currentTheme.FontPath, "fonts/Rounded_Mplus_1c_Bold.ttf", FontSizeLarge*scale)
	mediumFont := loadFont(currentTheme.FontPath, "fonts/Rounded_Mplus_1c_Bold.ttf", FontSizeMedium*scale)
	smallFont := loadFont(currentTheme.FontPath, "fonts/Rounded_Mplus_1c_Bold.ttf", FontSizeSmall*scale)
	tinyFont := loadFont(currentTheme.FontPath, "fonts/Rounded_Mplus_1c_Bold.ttf", FontSizeTiny*scale)
	microFont := loadFont(currentTheme.FontPath, "fonts/Rounded_Mplus_1c_Bold.ttf", FontSizeMicro*scale)

	largeSymbolFont := loadFont("fonts/CFPG.ttf", "", FontSizeLarge*scale)
	mediumSymbolFont := loadFont("fonts/CFPG.ttf", "", FontSizeMedium*scale)
	smallSymbolFont := loadFont("fonts/CFPG.ttf", "", FontSizeSmall*scale)
	tinySymbolFont := loadFont("fonts/CFPG.ttf", "", FontSizeTiny*scale)
	microSymbolFont := loadFont("fonts/CFPG.ttf", "", FontSizeMicro*scale)

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
