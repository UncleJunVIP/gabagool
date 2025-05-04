package main

import (
	"fmt"
	"nextui-sdl2/internal"
	"nextui-sdl2/models"
	"nextui-sdl2/ui"
)

func main() {
	internal.Init("Mortar")
	defer internal.SDLCleanup()

	menuItems := []models.MenuItem{
		{"Megathread", false, false, nil},
		{"SMB", false, false, nil},
		{"RomM", false, false, nil},
		{"nginx", false, false, nil},
		{"Apache", false, false, nil},
		{"Potato", false, false, nil},
		{"Salad", false, false, nil},
		{"Oh look", false, false, nil},
		{"There is more", false, false, nil},
		{"This scrolls?", false, false, nil},
		{"This scrolls!", false, false, nil},
		{"And its in Go!?", false, false, nil},
		{"Wait does that mean?", false, false, nil},
		{"Can this be a library?", false, false, nil},
		{"Hopefully.", false, false, nil},
		{"Should make things so much smoother.", false, false, nil},
		{"Better UX / UI", false, false, nil},
		{"Better Dev Experience", false, false, nil},
		{"START FROM THE TOP YO!", false, false, nil},
	}

	sel, err := ui.NewBlockingList("Mortar", menuItems, 20)
	if err != nil {
		internal.Logger.Error("Failed to create blocking list",
			"error", err)
	}

	fmt.Println(sel.SelectedItem.Text)

	downloads := []models.Download{
		{
			URL:         "https://myrient.erista.me/files/No-Intro/Nintendo%20-%20Game%20Boy%20Color/Pokemon%20-%20Crystal%20Version%20%28USA%2C%20Europe%29%20%28Rev%201%29.zip",
			Location:    "/mnt/SDCARD/Roms/2) Game Boy Color (GBC)/Pokemon - Crystal Version (USA).zip",
			DisplayName: "Pokémon - Crystal Version",
		},
		{
			URL:         "https://myrient.erista.me/files/No-Intro/Nintendo%20-%20Game%20Boy%20Color/Pokemon%20-%20Gold%20Version%20%28USA%2C%20Europe%29%20%28SGB%20Enhanced%29%20%28GB%20Compatible%29.zip",
			Location:    "/mnt/SDCARD/Roms/2) Game Boy Color (GBC)/Pokemon - Gold Version (USA).zip",
			DisplayName: "Pokémon - Gold Version",
		},
	}

	result, err := ui.NewBlockingDownload(downloads)
	if err != nil {
		internal.Logger.Error("Download error", "error", err)
	}

	fmt.Printf("Completed: %d, Failed: %d, Cancelled: %t\n",
		len(result.CompletedDownloads),
		len(result.FailedDownloads),
		result.Cancelled)

	// Process completed downloads
	for _, download := range result.CompletedDownloads {
		fmt.Printf("Successfully downloaded: %s\n", download.DisplayName)
	}

	// Process failed downloads and their errors
	for i, download := range result.FailedDownloads {
		fmt.Printf("Failed to download: %s, Error: %s\n",
			download.DisplayName,
			result.Errors[i].Error())
	}

	res, err := ui.NewBlockingKeyboard("Hello world")
	if err != nil {
		internal.Logger.Error("Failed to create blocking keyboard",
			"error", err)
	}

	fmt.Println(res)

	//sceneManager := internal.NewSceneManager()
	//
	//menuScene := scenes.NewMenuScene(internal.GetWindow().Renderer)
	//sceneManager.AddScene("mainMenu", menuScene)
	//
	//downloadScene := scenes.NewDownloadScene(internal.GetWindow())
	//
	//var downloads []scenes.Download
	downloads = append(downloads)

	downloads = append(downloads, models.Download{
		URL:         "https://myrient.erista.me/files/No-Intro/Nintendo%20-%20Game%20Boy%20Color/Pokemon%20-%20Gold%20Version%20%28USA%2C%20Europe%29%20%28SGB%20Enhanced%29%20%28GB%20Compatible%29.zip",
		Location:    "/mnt/SDCARD/Roms/2) Game Boy Color (GBC)/Pokemon - Gold Version (USA).zip",
		DisplayName: "Pokémon - Gold Version",
	})
	//
	//downloadScene.SetDownloads(downloads)
	//
	//sceneManager.AddScene("download", downloadScene)
	//
	//sceneManager.SwitchTo("mainMenu")
	//
	//running := true
	//var event sdl.Event
	//
	//for running {
	//	for event = sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
	//		switch t := event.(type) {
	//		case *sdl.QuitEvent:
	//			running = false
	//		case *sdl.KeyboardEvent:
	//			if t.Type == sdl.KEYDOWN && t.Keysym.Sym == sdl.K_ESCAPE {
	//				running = false
	//			} else if t.Type == sdl.KEYDOWN && t.Keysym.Sym == sdl.K_SPACE {
	//				sceneManager.HandleEvent(event)
	//			}
	//
	//		case *sdl.ControllerButtonEvent:
	//			if t.Type == sdl.CONTROLLERBUTTONDOWN {
	//				internal.Logger.Info("Controller button pressed",
	//					"controller", t.Which,
	//					"button", t.Button)
	//				sceneManager.HandleEvent(event)
	//			}
	//
	//		case *sdl.ControllerAxisEvent:
	//			if t.Value > 10000 || t.Value < -10000 {
	//				internal.Logger.Debug("Controller axis moved",
	//					"controller", t.Which,
	//					"axis", t.Axis,
	//					"value", t.Value)
	//				sceneManager.HandleEvent(event)
	//			}
	//		default:
	//			sceneManager.HandleEvent(event)
	//		}
	//	}
	//
	//	sceneManager.Update()
	//	sceneManager.Render()
	//	internal.GetWindow().Renderer.Present()
	//
	//	time.Sleep(time.Millisecond * 16)
	//}
}
