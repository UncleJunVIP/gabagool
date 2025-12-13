package internal

import (
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
	XLarge: 60,
	Large:  50,
	Medium: 44,
	Small:  34,
	Tiny:   24,
	Micro:  18,
}

var Fonts fontsManager

type fontsManager struct {
	ExtraLargeFont *ttf.Font
	LargeFont      *ttf.Font
	MediumFont     *ttf.Font
	SmallFont      *ttf.Font
	TinyFont       *ttf.Font
	MicroFont      *ttf.Font

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
	fontPath := GetTheme().FontPath
	fallback := os.Getenv("FALLBACK_FONT")
	symbolPath := "/mnt/SDCARD/.system/res/font1.ttf"

	// Calculate all sizes
	calcSize := func(base int) int {
		return CalculateFontSizeForResolution(base, screenWidth)
	}

	Fonts = fontsManager{
		ExtraLargeFont:   loadFont(fontPath, fallback, calcSize(sizes.XLarge)),
		LargeFont:        loadFont(fontPath, fallback, calcSize(sizes.Large)),
		MediumFont:       loadFont(fontPath, fallback, calcSize(sizes.Medium)),
		SmallFont:        loadFont(fontPath, fallback, calcSize(sizes.Small)),
		TinyFont:         loadFont(fontPath, fallback, calcSize(sizes.Tiny)),
		MicroFont:        loadFont(fontPath, fallback, calcSize(sizes.Micro)),
		LargeSymbolFont:  loadFont(symbolPath, fallback, calcSize(sizes.Large)),
		mediumSymbolFont: loadFont(symbolPath, fallback, calcSize(sizes.Medium)),
		smallSymbolFont:  loadFont(symbolPath, fallback, calcSize(sizes.Small)),
		tinySymbolFont:   loadFont(symbolPath, fallback, calcSize(sizes.Tiny)),
		microSymbolFont:  loadFont(symbolPath, fallback, calcSize(sizes.Micro)),
	}
}

func loadFont(path string, fallback string, size int) *ttf.Font {
	var font *ttf.Font
	var err error

	if path == "" {
		if fallback == "" {
			GetInternalLogger().Error("Both font path and fallback are empty", "size", size)
			os.Exit(1)
		}
		path = fallback
	}

	font, err = ttf.OpenFont(path, size)
	if err == nil {
		return font
	}

	if fallback == "" || fallback == path {
		GetInternalLogger().Error("Failed to load font", "path", path, "size", size, "error", err)
		os.Exit(1)
	}

	font, err = ttf.OpenFont(fallback, size)
	if err != nil {
		GetInternalLogger().Error("Failed to load font and fallback", "path", path, "fallback", fallback, "size", size, "error", err)
		os.Exit(1)
	}

	return font
}

func closeFonts() {
	Fonts.LargeFont.Close()
	Fonts.MediumFont.Close()
	Fonts.SmallFont.Close()
	Fonts.TinyFont.Close()
	Fonts.MicroFont.Close()

	Fonts.LargeSymbolFont.Close()
	Fonts.mediumSymbolFont.Close()
	Fonts.smallSymbolFont.Close()
	Fonts.tinySymbolFont.Close()
	Fonts.microSymbolFont.Close()
}
