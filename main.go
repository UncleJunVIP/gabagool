package main

import (
	"fmt"
	"nextui-sdl2/models"
	"nextui-sdl2/ui"
	"os"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	WindowWidth   = 1024
	WindowHeight  = 768
	FontSize      = 50
	SmallFontSize = 30
)

func main() {

	window := ui.InitWindow("Mortar", WindowWidth, WindowHeight, FontSize, SmallFontSize)

	// Menu items
	menuItems := []models.MenuItem{
		{"Megathread", true},
		{"SMB", false},
		{"RomM", false},
		{"nginx", false},
		{"Apache", false},
		{"Potato", false},
		{"Salad", false},
		{"Oh look", false},
		{"There is more", false},
		{"This scrolls?", false},
		{"This scrolls!", false},
		{"And its in Go!?", false},
		{"Wait does that mean?", false},
		{"Can this be a library?", false},
		{"Hopefully.", false},
		{"Should make things so much smoother.", false},
		{"Better UX / UI", false},
		{"Better Dev Experience", false},
		{"START FROM THE TOP YO!", false},
	}

	listController := ui.NewListController(menuItems, 40)

	// Configure list appearance
	listController.Settings = ui.MenuSettings{
		Spacing:      90, // 90 pixels between menu items
		XMargin:      10, // 20 pixel left margin
		YMargin:      10, // 10 pixel top margin within each item
		TextXPad:     10, // 30 pixel horizontal padding around text
		TextYPad:     10, // 8 pixel vertical padding around text
		Title:        "Mortar",
		TitleXMargin: 20,
		TitleSpacing: 20,
	}

	listController.MaxVisibleItems = 7

	// Set callback for item selection
	listController.OnSelect = func(index int, item *models.MenuItem) {
		fmt.Printf("Selected: %s\n", item.Text)
	}

	apngPlayer, err := ui.NewAPNGPlayer(window.Renderer, "mortar.apng")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create APNG player: %s\n", err)
		os.Exit(1)
	}
	defer apngPlayer.Destroy()

	// Position the animation in the center of the window
	apngPlayer.SetPosition(0, 0)

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

		//// Clear screen
		//renderer.SetDrawColor(0, 0, 0, 255)
		//renderer.Clear()
		//
		//// Draw the menu
		//listController.Draw(renderer, font)

		apngPlayer.Update()

		// Render
		window.Renderer.SetDrawColor(0, 0, 0, 255)
		window.Renderer.Clear()
		apngPlayer.Render()
		window.Renderer.Present()

		time.Sleep(time.Millisecond * 16)

	}

}
