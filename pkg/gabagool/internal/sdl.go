package internal

import (
	"os"

	"github.com/UncleJunVIP/gabagool/v2/pkg/gabagool/constants"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var window *Window

func Init(title string, showBackground bool) {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_AUDIO |
		img.INIT_PNG | img.INIT_JPG | img.INIT_TIF | img.INIT_WEBP |
		sdl.INIT_GAMECONTROLLER | sdl.INIT_JOYSTICK); err != nil {
		os.Exit(1)
	}

	if err := ttf.Init(); err != nil {
		os.Exit(1)
	}

	InitInputProcessor()

	window = initWindow(title, showBackground)

	initFonts(GetConfig())

	if !constants.IsDevMode() {
		window.initPowerButtonHandling()
	}
}

func SDLCleanup() {
	window.closeWindow()
	CloseAllControllers()
	closeFonts()
	ttf.Quit()
	img.Quit()
	sdl.Quit()
	CloseLogger()
}
