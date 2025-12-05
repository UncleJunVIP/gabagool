package gabagool

import (
	"log/slog"
	"os"

	"github.com/UncleJunVIP/gabagool/v2/pkg/gabagool/internal"
	"github.com/UncleJunVIP/gabagool/v2/pkg/gabagool/platform/cannoli"
	"github.com/UncleJunVIP/gabagool/v2/pkg/gabagool/platform/nextui"
)

type Options struct {
	WindowTitle          string
	ShowBackground       bool
	PrimaryThemeColorHex uint32
	IsCannoli            bool
	IsNextUI             bool
	ControllerConfigFile string
	LogFilename          string
}

// Init initializes SDL and the UI
// Must be called before any other UI functions!
func Init(options Options) {
	internal.SetFilename(options.LogFilename)

	if os.Getenv("NITRATES") != "" || os.Getenv("INPUT_CAPTURE") != "" {
		internal.SetInternalLogLevel(slog.LevelDebug)
	} else {
		internal.SetInternalLogLevel(slog.LevelError)
	}

	config := internal.GetConfig()

	if options.IsNextUI {
		theme := nextui.InitNextUITheme()
		internal.SetTheme(theme)
	} else if options.IsCannoli {
		internal.SetTheme(cannoli.InitCannoliTheme(config.Theme.DefaultFontPath))
	} else {
		internal.SetTheme(cannoli.InitCannoliTheme(config.Theme.DefaultFontPath)) // TODO fix this
	}

	if options.PrimaryThemeColorHex != 0 && !options.IsNextUI {
		theme := internal.GetTheme()
		theme.PrimaryAccentColor = internal.HexToColor(options.PrimaryThemeColorHex)
		internal.SetTheme(theme)
	}

	internal.Init(options.WindowTitle, options.ShowBackground)

	if os.Getenv("INPUT_CAPTURE") != "" {
		mapping := InputLogger()
		if mapping != nil {
			err := mapping.SaveToJSON("custom_input_mapping.json")
			if err != nil {
				internal.GetInternalLogger().Error("Failed to save custom input mapping", "error", err)
			}
		}
		os.Exit(0)
	}
}

// Close Tidies up SDL and the UI
// Must be called after all UI functions!
func Close() {
	internal.SDLCleanup()
}

func GetLogger() *slog.Logger {
	return internal.GetLogger()
}

func SetLogLevel(level slog.Level) {
	internal.SetLogLevel(level)
}

func SetRawLogLevel(level string) {
	internal.SetRawLogLevel(level)
}

func GetWindow() *internal.Window {
	return internal.GetWindow()
}

func HideWindow() {
	internal.GetWindow().Window.Hide()
}

func ShowWindow() {
	internal.GetWindow().Window.Show()
}
