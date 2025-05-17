package ui

import (
	"github.com/UncleJunVIP/gabagool/internal"
	"github.com/UncleJunVIP/gabagool/models"
)

func InitOptions(title string) models.GabagoolOptions {
	return models.GabagoolOptions{
		WindowTitle:    title,
		ShowBackground: false,
	}
}

// InitSDL initializes SDL and the UI
// Must be called before any other UI functions!
func InitSDL(options models.GabagoolOptions) {
	internal.Init(options)
}

// CloseSDL Tidies up SDL and the UI
// Must be called after all UI functions!
func CloseSDL() {
	internal.SDLCleanup()
}
