package main

import (
	"fmt"
	"github.com/UncleJunVIP/gabagool/internal"
	"github.com/UncleJunVIP/gabagool/models"
	"github.com/UncleJunVIP/gabagool/ui"
	"github.com/veandco/go-sdl2/sdl"
	"time"
)

func main() {
	internal.Init("Mortar")
	defer internal.SDLCleanup()

	conf, err := ui.NewBlockingConfirmation("Do you like potatoes?")

	fmt.Println(conf)

	aniRes, err := ui.NewBlockingAnimation("mortar.apng",
		ui.WithLooping(true),
		ui.WithMaxDisplayTime(time.Second*10),
		ui.WithBackgroundColor(sdl.Color{R: 20, G: 20, B: 50, A: 255}),
	)

	fmt.Println(aniRes)

	menuItems := []models.MenuItem{
		{Text: "Megathread"},
		{Text: "SMB"},
		{Text: "RomM"},
		{Text: "nginx"},
		{Text: "Apache"},
		{Text: "Potato"},
		{Text: "Salad"},
		{Text: "Oh look"},
		{Text: "There is more"},
		{Text: "This scrolls?"},
		{Text: "This scrolls!"},
		{Text: "And its in Go!?"},
		{Text: "Wait does that mean?"},
		{Text: "Can this be a library?"},
		{Text: "Hopefully."},
		{Text: "Should make things so much smoother."},
		{Text: "Better UX / UI"},
		{Text: "Better Dev Experience"},
		{Text: "START FROM THE TOP YO!"},
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

	for _, download := range result.CompletedDownloads {
		fmt.Printf("Successfully downloaded: %s\n", download.DisplayName)
	}

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

}
