package internal

import (
	"fmt"
	"os"

	"github.com/veandco/go-sdl2/ttf"
)

type FontSizes struct {
	XLarge int `json:"xlarge" yaml:"xlarge"`
	Large  int `json:"large" yaml:"large"`
	Medium int `json:"medium" yaml:"medium"`
	Small  int `json:"small" yaml:"small"`
	Tiny   int `json:"tiny" yaml:"tiny"`
	Micro  int `json:"micro" yaml:"micro"`
}

var DefaultFontSizes = FontSizes{
	XLarge: 66,
	Large:  54,
	Medium: 48,
	Small:  36,
	Tiny:   24,
	Micro:  18,
}

var Fonts fontsManager

type fontsManager struct {
	ExtraLargeFont *ttf.Font
	LargeFont      *ttf.Font
	MediumFont     *ttf.Font
	SmallFont      *ttf.Font
	tinyFont       *ttf.Font
	microFont      *ttf.Font

	LargeSymbolFont  *ttf.Font
	mediumSymbolFont *ttf.Font
	smallSymbolFont  *ttf.Font
	tinySymbolFont   *ttf.Font
	microSymbolFont  *ttf.Font
}

func CalculateFontSizeForResolution(baseSize int, screenWidth int32) int {
	const referenceWidth int32 = 1024
	scaleFactor := float32(screenWidth) / float32(referenceWidth)

	// Apply damping for larger screens to reduce scaling growth
	if screenWidth > referenceWidth {
		scaleFactor = 1.0 + (scaleFactor-1.0)*0.75 // 75% of the growth above 1x
	}

	return int(float32(baseSize) * scaleFactor)
}

// GetScaleFactor returns the scale factor based on current screen width
func GetScaleFactor() float32 {
	const referenceWidth int32 = 1024
	screenWidth := GetWindow().GetWidth()

	scaleFactor := float32(screenWidth) / float32(referenceWidth)

	// Apply damping for larger screens
	if screenWidth > referenceWidth {
		scaleFactor = 1.0 + (scaleFactor-1.0)*0.75
	}

	return scaleFactor
}

func initFonts(sizes FontSizes) {
	screenWidth := GetWindow().GetWidth()

	xlSize := CalculateFontSizeForResolution(sizes.XLarge, screenWidth)
	largeSize := CalculateFontSizeForResolution(sizes.Large, screenWidth)
	mediumSize := CalculateFontSizeForResolution(sizes.Medium, screenWidth)
	smallSize := CalculateFontSizeForResolution(sizes.Small, screenWidth)
	tinySize := CalculateFontSizeForResolution(sizes.Tiny, screenWidth)
	microSize := CalculateFontSizeForResolution(sizes.Micro, screenWidth)

	xlFont := loadFont(GetTheme().FontPath, os.Getenv("FALLBACK_FONT"), xlSize)
	LargeFont := loadFont(GetTheme().FontPath, os.Getenv("FALLBACK_FONT"), largeSize)
	MediumFont := loadFont(GetTheme().FontPath, os.Getenv("FALLBACK_FONT"), mediumSize)
	SmallFont := loadFont(GetTheme().FontPath, os.Getenv("FALLBACK_FONT"), smallSize)
	tinyFont := loadFont(GetTheme().FontPath, os.Getenv("FALLBACK_FONT"), tinySize)
	microFont := loadFont(GetTheme().FontPath, os.Getenv("FALLBACK_FONT"), microSize)

	LargeSymbolFont := loadFont("/mnt/SDCARD/.system/res/font1.ttf", os.Getenv("FALLBACK_FONT"), largeSize)
	mediumSymbolFont := loadFont("/mnt/SDCARD/.system/res/font1.ttf", os.Getenv("FALLBACK_FONT"), mediumSize)
	smallSymbolFont := loadFont("/mnt/SDCARD/.system/res/font1.ttf", os.Getenv("FALLBACK_FONT"), smallSize)
	tinySymbolFont := loadFont("/mnt/SDCARD/.system/res/font1.ttf", os.Getenv("FALLBACK_FONT"), tinySize)
	microSymbolFont := loadFont("/mnt/SDCARD/.system/res/font1.ttf", os.Getenv("FALLBACK_FONT"), microSize)

	Fonts = fontsManager{
		ExtraLargeFont: xlFont,
		LargeFont:      LargeFont,
		MediumFont:     MediumFont,
		SmallFont:      SmallFont,
		tinyFont:       tinyFont,
		microFont:      microFont,

		LargeSymbolFont:  LargeSymbolFont,
		mediumSymbolFont: mediumSymbolFont,
		smallSymbolFont:  smallSymbolFont,
		tinySymbolFont:   tinySymbolFont,
		microSymbolFont:  microSymbolFont,
	}
}

func loadFont(path string, fallback string, size int) *ttf.Font {
	font, err := ttf.OpenFont(path, size)
	if err != nil && fallback == "" {
		GetInternalLogger().Error("Failed to load font!", err)
		os.Exit(1)
	} else if err != nil {
		GetInternalLogger().Error("Failed to load font! Attempting to use fallback...", err)
		font, err = ttf.OpenFont(fallback, size)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fallback font: %s\n", err)
			os.Exit(1)
		}
	}

	return font
}

func closeFonts() {
	Fonts.LargeFont.Close()
	Fonts.MediumFont.Close()
	Fonts.SmallFont.Close()
	Fonts.tinyFont.Close()
	Fonts.microFont.Close()

	Fonts.LargeSymbolFont.Close()
	Fonts.mediumSymbolFont.Close()
	Fonts.smallSymbolFont.Close()
	Fonts.tinySymbolFont.Close()
	Fonts.microSymbolFont.Close()
}
