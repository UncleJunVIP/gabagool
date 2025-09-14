package gabagool

import (
	"fmt"
	"os"
	"strings"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
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

	detectAndOpenAllGameControllers()

	window = initWindow(title, showBackground)

	if !IsDev {
		window.initPowerButtonHandling()
	}
}

func SDLCleanup() {
	window.closeWindow()
	for _, controller := range gameControllers {
		if controller != nil {
			controller.Close()
		}
	}
	ttf.Quit()
	img.Quit()
	sdl.Quit()
}

func detectAndOpenAllGameControllers() {
	numJoysticks := sdl.NumJoysticks()
	GetLoggerInstance().Info("Detecting controllers", "numJoysticks", numJoysticks)

	for i := 0; i < numJoysticks; i++ {
		if sdl.IsGameController(i) {
			controller := sdl.GameControllerOpen(i)
			if controller != nil {
				name := controller.Name()

				// Try to get GUID for mapping (this might not work with current SDL version)
				// If GUID methods aren't available, we'll use controller name for basic mapping
				guid := fmt.Sprintf("controller_%d_%s", i, strings.ReplaceAll(name, " ", "_"))

				GetLoggerInstance().Info("Opened game controller",
					"index", i,
					"name", name,
					"guid", guid,
				)

				// Initialize dynamic button mapping for this controller
				InitializeControllerMapping(name, guid)

				gameControllers = append(gameControllers, controller)
			} else {
				GetLoggerInstance().Error("Failed to open game controller", "index", i)
			}
		} else {
			// This is a joystick but not recognized as a game controller
			joystick := sdl.JoystickOpen(i)
			if joystick != nil {
				name := joystick.Name()
				GetLoggerInstance().Info("Found joystick (not a game controller)",
					"index", i,
					"name", name,
				)
				joystick.Close()
			}
		}
	}

	GetLoggerInstance().Info("Controller detection complete",
		"gameControllers", len(gameControllers),
		"totalJoysticks", numJoysticks,
	)

	// Apply the best available controller mapping to global constants
	ApplyBestControllerMapping()

	// Log the final button mapping that's being used
	currentMapping := GetCurrentButtonMapping()
	GetLoggerInstance().Info("Final button mapping applied", "mapping", currentMapping)
}

func GetConnectedControllers() []map[string]interface{} {
	var controllers []map[string]interface{}

	for i, controller := range gameControllers {
		if controller == nil {
			continue
		}

		controllerInfo := map[string]interface{}{
			"index": i,
			"name":  controller.Name(),
		}

		controllers = append(controllers, controllerInfo)
	}

	return controllers
}

func LogControllerDetails() {
	controllers := GetConnectedControllers()

	if len(controllers) == 0 {
		GetLoggerInstance().Info("No game controllers connected")
		return
	}

	for _, controller := range controllers {
		GetLoggerInstance().Info("Controller Details",
			"index", controller["index"],
			"name", controller["name"],
		)
	}
}
