package main

import (
	gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"
)

func main() {
	gaba.InitSDL(gaba.Options{
		WindowTitle:    "Input Tester",
		ShowBackground: true,
		IsCannoli:      true,
		LogFilename:    "input_tester.log",
	})

	defer gaba.CloseSDL()

	gaba.InputLogger()

}
