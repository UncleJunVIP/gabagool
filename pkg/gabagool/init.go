package gabagool

import "log/slog"

type Options struct {
	WindowTitle          string
	ShowBackground       bool
	IsCannoli            bool
	ControllerConfigFile string
	LogFilename          string
	LogLevel             slog.Level
}

// InitSDL initializes SDL and the UI
// Must be called before any other UI functions!
func InitSDL(options Options) {
	setLogFilename(options.LogFilename)
	SetLogLevel(options.LogLevel)

	if options.IsCannoli {
		initTheme()
	} else {
		initNextUITheme()
	}

	Init(options.WindowTitle, options.ShowBackground, options.ControllerConfigFile)
}

// CloseSDL Tidies up SDL and the UI
// Must be called after all UI functions!
func CloseSDL() {
	closeFonts()
	SDLCleanup()
}

func HideWindow() {
	window.Window.Hide()
}

func ShowWindow() {
	window.Window.Show()
}
