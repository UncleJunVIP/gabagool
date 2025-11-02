package gabagool

import (
	"log/slog"

	"github.com/UncleJunVIP/gabagool/pkg/gabagool/core"
	"github.com/UncleJunVIP/gabagool/pkg/gabagool/platform/cannoli"
	"github.com/UncleJunVIP/gabagool/pkg/gabagool/platform/nextui"
)

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

	config := GetConfig()

	if options.IsCannoli {
		core.SetTheme(cannoli.InitCannoliTheme(config.Theme.DefaultFontPath))
	} else {
		core.SetTheme(nextui.InitNextUITheme())
	}

	Init(options.WindowTitle, options.ShowBackground)
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
