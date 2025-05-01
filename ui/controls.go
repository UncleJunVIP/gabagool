package ui

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
)

var gameController *sdl.GameController

func init() {
	for i := 0; i < sdl.NumJoysticks(); i++ {
		if sdl.IsGameController(i) {
			gameController = sdl.GameControllerOpen(i)
			if gameController != nil {
				fmt.Printf("Found game controller: %s\n", sdl.GameControllerNameForIndex(i))
				break
			}
		}
	}
	if gameController != nil {
		defer gameController.Close()
	}
}

func GetGameController() *sdl.GameController {
	return gameController
}

func CloseGameController() {
	gameController.Close()
}
