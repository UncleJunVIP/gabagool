package main

import gaba "github.com/UncleJunVIP/gabagool/pkg/gabagool"

func main() {
	gaba.InitSDL(gaba.Options{
		WindowTitle:    "Input Tester",
		ShowBackground: true,
		IsCannoli:      true,
	})

	defer gaba.CloseSDL()

	gaba.InputLogger()

}
