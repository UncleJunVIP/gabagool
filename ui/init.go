package ui

import "github.com/UncleJunVIP/gabagool/internal"

func InitSDL(applicationName string) {
	internal.Init(applicationName)
}

func CloseSDL() {
	internal.SDLCleanup()
}
