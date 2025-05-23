package gabagool

// GabagoolOptions is used to configure global settings
type GabagoolOptions struct {
	WindowTitle    string
	ShowBackground bool
}

// InitSDL initializes SDL and the UI
// Must be called before any other UI functions!
func InitSDL(options GabagoolOptions) {
	initTheme()
	Init(options.WindowTitle, options.ShowBackground)
}

// CloseSDL Tidies up SDL and the UI
// Must be called after all UI functions!
func CloseSDL() {
	closeFonts()
	SDLCleanup()
}
