package gabagool

import (
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"os"
)

var window *Window
var gameControllers []*sdl.GameController

func Init(title string, showBackground bool) {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO |
		img.INIT_PNG | img.INIT_JPG | img.INIT_TIF | img.INIT_WEBP |
		sdl.INIT_GAMECONTROLLER | sdl.INIT_JOYSTICK); err != nil {
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

	window = initWindow(title, showBackground)
}

func SDLCleanup() {
	window.closeWindow()
	for _, controller := range gameControllers {
		controller.Close()
	}
	ttf.Quit()
	img.Quit()
	sdl.Quit()
}
