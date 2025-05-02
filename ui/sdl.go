package ui

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"os"
)

var GameControllers []*sdl.GameController

func init() {
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
				GameControllers = append(GameControllers, controller)
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
