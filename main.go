package main

import (
	"nextui-sdl2/scenes"
	"nextui-sdl2/ui"
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
	window := ui.InitWindow("Mortar", WindowWidth, WindowHeight, FontSize, SmallFontSize)
	defer window.CloseWindow()

	sceneManager := ui.NewSceneManager(window)

	menuScene := scenes.NewMenuScene(window.Renderer)
	sceneManager.AddScene("mainMenu", menuScene)

	apngScene := scenes.NewAPNGScene(window.Renderer)
	sceneManager.AddScene("apng", apngScene)

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
						sceneManager.SwitchTo("apng")
					case "apng":
						sceneManager.SwitchTo("mainMenu")
					}
				} else {
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
