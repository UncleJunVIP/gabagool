package ui

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"log/slog"
	"os"
)

var Logger *slog.Logger
var GameControllers []*sdl.GameController

func init() {
	// Configure the logger
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stderr, opts)
	Logger = slog.New(handler)

	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_GAMECONTROLLER); err != nil {
		Logger.Error("Failed to initialize SDL", "error", err)
		os.Exit(1)
	}

	if err := ttf.Init(); err != nil {
		Logger.Error("Failed to initialize TTF", "error", err)
		os.Exit(1)
	}

	numJoysticks := sdl.NumJoysticks()

	for i := 0; i < numJoysticks; i++ {
		if sdl.IsGameController(i) {
			controller := sdl.GameControllerOpen(i)
			if controller != nil {
				GameControllers = append(GameControllers, controller)
				name := sdl.GameControllerNameForIndex(i)
				Logger.Info("Found gamepad", "name", name)
			}
		}
	}
}

func SDLCleanup() {
	for _, controller := range GameControllers {
		controller.Close()
	}
	ttf.Quit()
	sdl.Quit()
}
