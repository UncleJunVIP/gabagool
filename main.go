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

	downloadScene := scenes.NewDownloadScene(window)

	var downloads []scenes.Download
	downloads = append(downloads, scenes.Download{
		URL:         "https://myrient.erista.me/files/No-Intro/Nintendo%20-%20Game%20Boy%20Color/Pokemon%20-%20Crystal%20Version%20%28USA%2C%20Europe%29%20%28Rev%201%29.zip",
		Location:    "/mnt/SDCARD/Roms/2) Game Boy Color (GBC)/Pokemon - Crystal Version (USA).zip",
		DisplayName: "Pokémon - Crystal Version",
	})

	downloads = append(downloads, scenes.Download{
		URL:         "https://myrient.erista.me/files/No-Intro/Nintendo%20-%20Game%20Boy%20Color/Pokemon%20-%20Gold%20Version%20%28USA%2C%20Europe%29%20%28SGB%20Enhanced%29%20%28GB%20Compatible%29.zip",
		Location:    "/mnt/SDCARD/Roms/2) Game Boy Color (GBC)/Pokemon - Gold Version (USA).zip",
		DisplayName: "Pokémon - Gold Version",
	})

	downloadScene.SetDownloads(downloads)

	sceneManager.AddScene("download", downloadScene)

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
