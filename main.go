package main

import (
	"fmt"
	"nextui-sdl2/models"
	"nextui-sdl2/ui"
	"os"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	WindowWidth   = 1024
	WindowHeight  = 768
	FontSize      = 50
	SmallFontSize = 30
)

func main() {
	if err := sdl.Init(sdl.INIT_VIDEO | sdl.INIT_GAMECONTROLLER); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize SDL: %s\n", err)
		os.Exit(1)
	}
	defer sdl.Quit()

	if err := ttf.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize TTF: %s\n", err)
		os.Exit(1)
	}
	defer ttf.Quit()

	window, err := sdl.CreateWindow("NextUI UI (Damn that's kinda meta)", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		WindowWidth, WindowHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		os.Exit(1)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		os.Exit(1)
	}
	defer renderer.Destroy()

	// Load font
	font, err := ttf.OpenFont("BPreplayBold.ttf", FontSize)
	if err != nil {
		fmt.Printf("Warning: Failed to load font, using system font: %s\n", err)
		font, err = ttf.OpenFont("DejaVuSans.ttf", FontSize)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load fallback font: %s\n", err)
			os.Exit(1)
		}
	}
	defer font.Close()

	smallFont, err := ttf.OpenFont("BPreplay.ttf", SmallFontSize)
	if err != nil {
		smallFont, err = ttf.OpenFont("DejaVuSans.ttf", SmallFontSize)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load fallback small font: %s\n", err)
			os.Exit(1)
		}
	}
	defer smallFont.Close()

	// Initialize game controllers
	var gameController *sdl.GameController
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

	// Menu items
	menuItems := []models.MenuItem{
		{"Megathread", true},
		{"SMB", false},
		{"RomM", false},
		{"nginx", false},
		{"Apache", false},
		{"Not a Sponsor", false},
	}

	// Create list controller
	listController := ui.NewListController(menuItems, 40)

	// Configure list appearance
	listController.Settings = ui.MenuSettings{
		Spacing:  90, // 90 pixels between menu items
		XMargin:  20, // 20 pixel left margin
		YMargin:  10, // 10 pixel top margin within each item
		TextXPad: 30, // 30 pixel horizontal padding around text
		TextYPad: 8,  // 8 pixel vertical padding around text
	}

	// Set callback for item selection
	listController.OnSelect = func(index int, item *models.MenuItem) {
		fmt.Printf("Selected: %s\n", item.Text)
	}

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
				} else {
					listController.HandleEvent(event)
				}
			case *sdl.ControllerButtonEvent:
				if t.Type == sdl.CONTROLLERBUTTONDOWN && t.Button == sdl.CONTROLLER_BUTTON_X {
					os.Exit(0)
				} else {
					listController.HandleEvent(event)
				}
			}
		}

		// Clear screen
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		// Draw the menu
		listController.Draw(renderer, font)

		renderer.Present()
		sdl.Delay(6) // Cap at ~60 FPS
	}
}
