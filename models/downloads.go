package models

import "github.com/veandco/go-sdl2/sdl"

// Add this to the appropriate file in your models package

// DownloadReturn contains the result of a download operation
type DownloadReturn struct {
	CompletedDownloads []Download  // List of downloads that completed successfully
	FailedDownloads    []Download  // List of downloads that failed
	Errors             []error     // List of errors that occurred during download
	LastPressedKey     sdl.Keycode // Last pressed keyboard key
	LastPressedBtn     uint8       // Last pressed controller button
	Cancelled          bool        // Whether the download was cancelled by the user
}

// Define Download type if it doesn't already exist in models package
type Download struct {
	URL         string
	Location    string
	DisplayName string
}
