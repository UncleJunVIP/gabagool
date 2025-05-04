package models

import "github.com/veandco/go-sdl2/sdl"

type DownloadReturn struct {
	CompletedDownloads []Download
	FailedDownloads    []Download
	Errors             []error
	LastPressedKey     sdl.Keycode
	LastPressedBtn     uint8
	Cancelled          bool
}

type Download struct {
	URL         string
	Location    string
	DisplayName string
}
