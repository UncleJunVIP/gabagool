package models

type DownloadReturn struct {
	CompletedDownloads []Download
	FailedDownloads    []Download
	Errors             []error
	LastPressedBtn     uint8
	Cancelled          bool
}

type Download struct {
	URL         string
	Location    string
	DisplayName string
}
