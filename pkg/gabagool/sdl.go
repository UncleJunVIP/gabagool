package gabagool

import (
	"os"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

var window *Window
var gameControllers []*sdl.GameController
var rawJoysticks []*sdl.Joystick

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

	detectAndOpenAllGameControllers()

	window = initWindow(title, showBackground)

	if !IsDev {
		window.initPowerButtonHandling()
	}
}

func detectAndOpenAllGameControllers() {
	numJoysticks := sdl.NumJoysticks()
	GetLoggerInstance().Debug("Detecting controllers", "numJoysticks", numJoysticks)

	processor := GetInputProcessor()

	for i := 0; i < numJoysticks; i++ {
		if sdl.IsGameController(i) {
			controller := sdl.GameControllerOpen(i)
			if controller != nil {
				name := controller.Name()

				// Register this joystick INDEX (not instance ID) as being handled by a game controller
				processor.RegisterGameControllerJoystickIndex(i)

				GetLoggerInstance().Debug("Opened game controller",
					"index", i,
					"name", name,
				)

				gameControllers = append(gameControllers, controller)
			} else {
				GetLoggerInstance().Error("Failed to open game controller", "index", i)
			}
		} else {
			joystick := sdl.JoystickOpen(i)
			if joystick != nil {
				name := joystick.Name()
				GetLoggerInstance().Debug("Opened raw joystick (not a standard game controller)",
					"index", i,
					"name", name,
				)

				rawJoysticks = append(rawJoysticks, joystick)
			} else {
				GetLoggerInstance().Debug("Failed to open raw joystick", "index", i)
			}
		}
	}

	GetLoggerInstance().Debug("Controller detection complete",
		"gameControllers", len(gameControllers),
		"rawJoysticks", len(rawJoysticks),
		"totalJoysticks", numJoysticks,
	)
}

func SDLCleanup() {
	window.closeWindow()
	for _, controller := range gameControllers {
		if controller != nil {
			controller.Close()
		}
	}
	for _, joystick := range rawJoysticks {
		if joystick != nil {
			joystick.Close()
		}
	}
	ttf.Quit()
	img.Quit()
	sdl.Quit()
}
