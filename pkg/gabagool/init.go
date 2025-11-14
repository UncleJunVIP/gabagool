package gabagool

import (
	"log/slog"
	"os"

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
}

// InitSDL initializes SDL and the UI
// Must be called before any other UI functions!
func InitSDL(options Options) {
	setLogFilename(options.LogFilename)

	if os.Getenv("ENVIRONMENT") == "DEV" || os.Getenv("INPUT_CAPTURE") != "" {
		SetLogLevel(slog.LevelDebug)
	} else {
		SetLogLevel(slog.LevelError)
	}

	config := GetConfig()

	if options.IsCannoli {
		core.SetTheme(cannoli.InitCannoliTheme(config.Theme.DefaultFontPath))
	} else {
		core.SetTheme(nextui.InitNextUITheme())
	}

	Init(options.WindowTitle, options.ShowBackground)

	if os.Getenv("INPUT_CAPTURE") != "" {
		mapping := InputLogger()
		if mapping != nil {
			err := mapping.SaveToJSON("custom_input_mapping.json")
			if err != nil {
				GetLoggerInstance().Error("Failed to save custom input mapping", "error", err)
			}
		}
		os.Exit(0)
	}
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
