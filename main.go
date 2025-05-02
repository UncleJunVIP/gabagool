package main

import (
	"nextui-sdl2/scenes"
	"nextui-sdl2/ui"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

func main() {
	if err := ui.InitLogger("application.log"); err != nil {
		os.Exit(1)
	}

	defer ui.SDLCleanup()

	window := ui.InitWindow("Mortar")
	defer window.CloseWindow()

	sceneManager := ui.NewSceneManager(window)

	menuScene := scenes.NewMenuScene(window.Renderer)
	sceneManager.AddScene("mainMenu", menuScene)

	keyboardScene := scenes.NewKeyboardScene(window)
	sceneManager.AddScene("keyboard", keyboardScene)

	sceneManager.SwitchTo("mainMenu")

	running := true
	var event sdl.Event

	for running {
		for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				if t.Type == sdl.KEYDOWN && t.Keysym.Sym == sdl.K_ESCAPE {
					running = false
				} else if t.Type == sdl.KEYDOWN && t.Keysym.Sym == sdl.K_SPACE {
					switch sceneManager.GetCurrentSceneID() {
					case "mainMenu":
						sceneManager.SwitchTo("keyboard")
					case "keyboard":
						sceneManager.SwitchTo("mainMenu")
					}
				} else {
					sceneManager.HandleEvent(event)
				}

			case *sdl.ControllerButtonEvent:
				if t.Type == sdl.CONTROLLERBUTTONDOWN {
					ui.Logger.Info("Controller button pressed",
						"controller", t.Which,
						"button", t.Button)
					sceneManager.HandleEvent(event)
				}

			case *sdl.ControllerAxisEvent:
				if t.Value > 10000 || t.Value < -10000 {
					ui.Logger.Debug("Controller axis moved",
						"controller", t.Which,
						"axis", t.Axis,
						"value", t.Value)
					sceneManager.HandleEvent(event)
				}
			default:
				sceneManager.HandleEvent(event)
			}
		}

		sceneManager.Update()
		sceneManager.Render()
		window.Renderer.Present()

		time.Sleep(time.Millisecond * 16)
	}
}
