package internal

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"os"
)

var window *Window
var gameControllers []*sdl.GameController

func Init(applicationName string) {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_GAMECONTROLLER); err != nil {
		os.Exit(1)
	}

	if err := ttf.Init(); err != nil {
		os.Exit(1)
	}

	numJoysticks := sdl.NumJoysticks()

	for i := 0; i < numJoysticks; i++ {
		if sdl.IsGameController(i) {
			controller := sdl.GameControllerOpen(i)
			if controller != nil {
				gameControllers = append(gameControllers, controller)
			}
		}
	}

	window = InitWindow(applicationName)
}

func GetWindow() *Window {
	return window
}

func SDLCleanup() {
	window.CloseWindow()
	for _, controller := range gameControllers {
		controller.Close()
	}
	ttf.Quit()
	sdl.Quit()
}
