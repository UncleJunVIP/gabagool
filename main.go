package main

import (
	"fmt"
	"github.com/UncleJunVIP/gabagool/internal"
	"github.com/UncleJunVIP/gabagool/models"
	"github.com/UncleJunVIP/gabagool/ui"
)

func main() {
	internal.Init("Option List Demo")
	defer internal.SDLCleanup()
	//
	//conf, err := ui.NewBlockingConfirmation("Do you like potatoes?")
	//
	//fmt.Println(conf)

	//aniRes, err := ui.NewBlockingAnimation("mortar.apng",
	//	ui.WithLooping(true),
	//	ui.WithMaxDisplayTime(time.Second*10),
	//	ui.WithBackgroundColor(sdl.Color{R: 20, G: 20, B: 50, A: 255}),
	//)
	//
	//fmt.Println(aniRes)

	//menuItems := []models.MenuItem{
	//	{Text: "Game 1"},
	//	{Text: "Game 2"},
	//	{Text: "Game 3"},
	//	{Text: "Game 4"},
	//	{Text: "Game 5"},
	//	{Text: "Game 6"},
	//	{Text: "Game 7"},
	//	{Text: "Game 8"},
	//	{Text: "Game 9"},
	//	{Text: "Game 10"},
	//	{Text: "Game 11"},
	//	{Text: "Game 12"},
	//}
	//
	//fhi := []ui.FooterHelpItem{
	//	{ButtonName: "B", HelpText: "Quit"},
	//	{ButtonName: "A", HelpText: "Select"},
	//}
	//
	//_, _ = ui.NewBlockingList("List Reorder Demo", menuItems, "", fhi, false, false, true)

	items := []ui.ItemWithOptions{
		{
			Item: models.MenuItem{
				Text: "Download Art",
			},
			Options: []ui.Option{
				{DisplayName: "True", Value: true},
				{DisplayName: "False", Value: false},
			},
		},
		{
			Item: models.MenuItem{
				Text: "Art Type",
			},
			Options: []ui.Option{
				{DisplayName: "Box Art", Value: "BOX_ART"},
				{DisplayName: "Title Screen", Value: "TITLE_SCREEN"},
				{DisplayName: "Logos", Value: "LOGOS"},
				{DisplayName: "Screenshots", Value: "SCREENSHOTS"},
			},
			SelectedOption: 1, // Default to 1080p
		},
		{
			Item: models.MenuItem{
				Text: "Cache Game Lists",
			},
			Options: []ui.Option{
				{DisplayName: "False", Value: false},
				{DisplayName: "True", Value: true},
			},
			SelectedOption: 0, // Default to High
		},
	}

	footerHelpItems := []ui.FooterHelpItem{
		{ButtonName: "A", HelpText: "Select"},
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "←→", HelpText: "Change option"},
	}

	result, err := ui.NewBlockingOptionsList(
		"Settings",
		items,
		footerHelpItems,
	)

	if err != nil {
		fmt.Println("Error showing options list:", err)
		return
	}

	fmt.Println(result)

	//downloads := []models.Download{
	//	{
	//		URL:         "https://myrient.erista.me/files/No-Intro/Nintendo%20-%20Game%20Boy%20Color/Pokemon%20-%20Crystal%20Version%20%28USA%2C%20Europe%29%20%28Rev%201%29.zip",
	//		Location:    "/mnt/SDCARD/Roms/2) Game Boy Color (GBC)/Pokemon - Crystal Version (USA).zip",
	//		DisplayName: "Pokémon - Crystal Version",
	//	},
	//	{
	//		URL:         "https://myrient.erista.me/files/No-Intro/Nintendo%20-%20Game%20Boy%20Color/Pokemon%20-%20Gold%20Version%20%28USA%2C%20Europe%29%20%28SGB%20Enhanced%29%20%28GB%20Compatible%29.zip",
	//		Location:    "/mnt/SDCARD/Roms/2) Game Boy Color (GBC)/Pokemon - Gold Version (USA).zip",
	//		DisplayName: "Pokémon - Gold Version",
	//	},
	//}

	//result, err := ui.NewBlockingDownload(downloads)
	//if err != nil {
	//	internal.Logger.Error("Download error", "error", err)
	//}
	//
	//fmt.Printf("Completed: %d, Failed: %d, Cancelled: %t\n",
	//	len(result.CompletedDownloads),
	//	len(result.FailedDownloads),
	//	result.Cancelled)
	//
	//for _, download := range result.CompletedDownloads {
	//	fmt.Printf("Successfully downloaded: %s\n", download.DisplayName)
	//}
	//
	//for i, download := range result.FailedDownloads {
	//	fmt.Printf("Failed to download: %s, Error: %s\n",
	//		download.DisplayName,
	//		result.Errors[i].Error())
	//}

	//res, err := ui.NewBlockingKeyboard("Hello world")
	//if err != nil {
	//}
	//
	//fmt.Println(res)

}
