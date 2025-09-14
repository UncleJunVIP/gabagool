package gabagool

import (
	"os"

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
		controller.Close()
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
				joystick := controller.GetJoystick()

				var guid string
				var instanceID int32
				if joystick != nil {
					guid = joystick.GetGUID().String()
					instanceID = joystick.InstanceID()
				}

				GetLoggerInstance().Info("Opened game controller",
					"index", i,
					"name", name,
					"guid", guid,
					"instanceID", instanceID,
				)

				gameControllers = append(gameControllers, controller)
			} else {
				GetLoggerInstance().Error("Failed to open game controller", "index", i)
			}
		} else {
			joystick := sdl.JoystickOpen(i)
			if joystick != nil {
				name := joystick.Name()
				guid := joystick.GetGUID().String()
				GetLoggerInstance().Info("Found joystick (not a game controller)",
					"index", i,
					"name", name,
					"guid", guid,
				)
				joystick.Close()
			}
		}
	}

	GetLoggerInstance().Info("Controller detection complete",
		"gameControllers", len(gameControllers),
		"totalJoysticks", numJoysticks,
	)
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

		joystick := controller.GetJoystick()
		if joystick != nil {
			controllerInfo["guid"] = joystick.GetGUID().String()
			controllerInfo["instanceID"] = joystick.InstanceID()
			controllerInfo["numAxes"] = joystick.NumAxes()
			controllerInfo["numButtons"] = joystick.NumButtons()
			controllerInfo["numHats"] = joystick.NumHats()
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
			"guid", controller["guid"],
			"instanceID", controller["instanceID"],
			"numAxes", controller["numAxes"],
			"numButtons", controller["numButtons"],
			"numHats", controller["numHats"],
		)
	}
}
