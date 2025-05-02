package main

import (
	"log/slog"
	"nextui-sdl2/scenes"
	"nextui-sdl2/ui"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	WindowWidth   = 1024
	WindowHeight  = 768
	FontSize      = 40
	SmallFontSize = 20
)

func main() {
	defer ui.SDLCleanup()

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stderr, opts)
	logger := slog.New(handler)

	window := ui.InitWindow("Mortar", WindowWidth, WindowHeight, FontSize, SmallFontSize)
	defer window.CloseWindow()

	sceneManager := ui.NewSceneManager(window)

	menuScene := scenes.NewMenuScene(window.Renderer)
	sceneManager.AddScene("mainMenu", menuScene)

	apngScene := scenes.NewAPNGScene(window.Renderer)
	sceneManager.AddScene("apng", apngScene)

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
					switch sceneManager.CurrentSceneName() {
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
					logger.Info("Controller button pressed",
						"controller", t.Which,
						"button", t.Button)

					if t.Button == sdl.CONTROLLER_BUTTON_X {
						os.Exit(0)
					} else if t.Button == sdl.CONTROLLER_BUTTON_START {
						switch sceneManager.CurrentSceneName() {
						case "mainMenu":
							sceneManager.SwitchTo("keyboard")
						case "keyboard":
							sceneManager.SwitchTo("mainMenu")
						}
					}

					sceneManager.HandleEvent(event)
				}

			case *sdl.ControllerAxisEvent:
				if t.Value > 10000 || t.Value < -10000 {
					logger.Debug("Controller axis moved",
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
