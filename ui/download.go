package ui

import (
	"fmt"
	"github.com/UncleJunVIP/gabagool/internal"
	"github.com/UncleJunVIP/gabagool/models"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

type DownloadReturn struct {
	CompletedDownloads []models.Download
	FailedDownloads    []models.Download
	Errors             []error
	LastPressedKey     sdl.Keycode
	LastPressedBtn     uint8
	Cancelled          bool
}

type downloadManager struct {
	window            *internal.Window
	downloads         []models.Download
	currentIndex      int
	downloadProgress  float64
	totalSize         int64
	downloadedSize    int64
	isDownloading     bool
	downloadComplete  bool
	downloadError     error
	cancelDownload    chan struct{}
	downloadDone      chan bool
	cancellationError error

	progressBarWidth  int32
	progressBarHeight int32
	progressBarX      int32
	progressBarY      int32

	headers map[string]string
}

func Download(downloads []models.Download, headers map[string]string) (DownloadReturn, error) {
	downloadManager := newDownloadManager(downloads, headers)

	result := DownloadReturn{
		CompletedDownloads: []models.Download{},
		FailedDownloads:    []models.Download{},
		Errors:             []error{},
		LastPressedKey:     0,
		LastPressedBtn:     0,
		Cancelled:          false,
	}

	if len(downloads) == 0 {
		return result, nil
	}

	window := internal.GetWindow()
	renderer := window.Renderer

	downloadManager.render(renderer)
	renderer.Present()

	sdl.Delay(100)

	downloadManager.startDownload()

	running := true
	allDownloadsComplete := false
	var err error

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
				err = sdl.GetError()
				downloadManager.cancellationError = fmt.Errorf("download cancelled by user")
				close(downloadManager.cancelDownload)
				result.Cancelled = true

			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {
					result.LastPressedKey = e.Keysym.Sym

					if allDownloadsComplete {
						running = false
						continue
					}

					if e.Keysym.Sym == sdl.K_ESCAPE {
						if downloadManager.isDownloading {
							close(downloadManager.cancelDownload)
							downloadManager.cancelDownload = make(chan struct{})
							downloadManager.cancellationError = fmt.Errorf("download cancelled by user")
						}
						result.Cancelled = true
						downloadManager.downloadComplete = true
					}
				}

			case *sdl.ControllerButtonEvent:
				if e.Type == sdl.CONTROLLERBUTTONDOWN {
					result.LastPressedBtn = e.Button

					if allDownloadsComplete {
						running = false
						continue
					}

					if e.Button == BrickButton_B {
						if downloadManager.isDownloading {
							close(downloadManager.cancelDownload)
							downloadManager.cancelDownload = make(chan struct{})
							downloadManager.cancellationError = fmt.Errorf("download cancelled by user")
						}
						result.Cancelled = true
						downloadManager.downloadComplete = true
					} else if e.Button == BrickButton_Y {
						if downloadManager.isDownloading {
							close(downloadManager.cancelDownload)
							downloadManager.cancelDownload = make(chan struct{})
						}
						result.Cancelled = true
						downloadManager.downloadComplete = true

						for i := downloadManager.currentIndex; i < len(downloadManager.downloads); i++ {
							result.FailedDownloads = append(result.FailedDownloads, downloadManager.downloads[i])
							result.Errors = append(result.Errors, fmt.Errorf("download cancelled by user (Y button)"))
						}
					}
				}
			}
		}

		select {
		case success := <-downloadManager.downloadDone:
			if success {
				result.CompletedDownloads = append(result.CompletedDownloads,
					downloadManager.downloads[downloadManager.currentIndex])

				downloadManager.currentIndex++
				if downloadManager.currentIndex < len(downloadManager.downloads) {
					downloadManager.startDownload()
				} else {

					allDownloadsComplete = true
				}
			} else {
				result.FailedDownloads = append(result.FailedDownloads,
					downloadManager.downloads[downloadManager.currentIndex])
				result.Errors = append(result.Errors, downloadManager.downloadError)

				downloadManager.currentIndex++
				if downloadManager.currentIndex < len(downloadManager.downloads) {
					downloadManager.startDownload()
				} else {

					allDownloadsComplete = true
				}
			}
		default:
		}

		downloadManager.render(renderer)

		renderer.Present()

		sdl.Delay(16)
	}

	return result, err
}

func newDownloadManager(downloads []models.Download, headers map[string]string) *downloadManager {
	window := internal.GetWindow()

	progressBarWidth := window.Width * 3 / 4
	progressBarHeight := int32(30)
	progressBarX := (window.Width - progressBarWidth) / 2
	progressBarY := window.Height / 2

	return &downloadManager{
		window:            window,
		downloads:         downloads,
		headers:           headers,
		currentIndex:      0,
		isDownloading:     false,
		downloadComplete:  false,
		cancelDownload:    make(chan struct{}),
		downloadDone:      make(chan bool),
		progressBarWidth:  progressBarWidth,
		progressBarHeight: progressBarHeight,
		progressBarX:      progressBarX,
		progressBarY:      progressBarY,
	}
}

func (dm *downloadManager) startDownload() {
	if dm.currentIndex >= len(dm.downloads) {
		return
	}

	download := dm.downloads[dm.currentIndex]

	if dm.isDownloading {
		close(dm.cancelDownload)
		dm.cancelDownload = make(chan struct{})
	}

	dm.downloadProgress = 0
	dm.totalSize = 0
	dm.downloadedSize = 0
	dm.isDownloading = true
	dm.downloadComplete = false
	dm.downloadError = nil

	go dm.downloadFile(download.URL, download.Location)
}

func (dm *downloadManager) downloadFile(url, filePath string) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		dm.downloadError = err
		dm.isDownloading = false
		dm.downloadDone <- false
		return
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		dm.downloadError = err
		dm.isDownloading = false
		dm.downloadDone <- false
		return
	}

	if dm.headers != nil {
		for k, v := range dm.headers {
			req.Header.Add(k, v)
		}
	}

	client := &http.Client{
		Timeout: 15 * time.Minute,
	}
	resp, err := client.Do(req)
	if err != nil {
		dm.downloadError = err
		dm.isDownloading = false
		dm.downloadDone <- false
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		dm.downloadError = fmt.Errorf("bad status: %s", resp.Status)
		dm.isDownloading = false
		dm.downloadDone <- false
		return
	}

	dm.totalSize = resp.ContentLength

	out, err := os.Create(filePath)
	if err != nil {
		dm.downloadError = err
		dm.isDownloading = false
		dm.downloadDone <- false
		return
	}
	defer out.Close()

	reader := &progressReader{
		reader: resp.Body,
		onProgress: func(bytesRead int64) {
			dm.downloadedSize = bytesRead
			if dm.totalSize > 0 {
				dm.downloadProgress = float64(bytesRead) / float64(dm.totalSize)
			}
		},
	}

	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(out, reader)
		done <- err
	}()

	select {
	case err := <-done:
		if err != nil {
			dm.downloadError = err
			dm.isDownloading = false
			dm.downloadDone <- false
		} else {
			dm.downloadComplete = true
			dm.isDownloading = false
			dm.downloadDone <- true
		}
	case <-dm.cancelDownload:
		dm.downloadError = fmt.Errorf("Download Cancelled")
		dm.isDownloading = false
		dm.downloadDone <- false
	}
}

func (dm *downloadManager) render(renderer *sdl.Renderer) {
	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.Clear()

	font := internal.GetSmallFont()

	if len(dm.downloads) > 1 {
		titleFont := internal.GetXLargeFont()
		titleText := "Download Manager"
		titleSurface, err := titleFont.RenderUTF8Solid(titleText, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		if err == nil {
			titleTexture, err := renderer.CreateTextureFromSurface(titleSurface)
			if err == nil {
				titleRect := &sdl.Rect{
					X: (dm.window.Width - titleSurface.W) / 2,
					Y: dm.window.Height/4 - titleSurface.H/2,
					W: titleSurface.W,
					H: titleSurface.H,
				}
				renderer.Copy(titleTexture, nil, titleRect)
				titleTexture.Destroy()
			}
			titleSurface.Free()
		}

		currentDownloadIndex := dm.currentIndex + 1
		if currentDownloadIndex > len(dm.downloads) {
			currentDownloadIndex = len(dm.downloads)
		}

		progressText := fmt.Sprintf("Download %d of %d", currentDownloadIndex, len(dm.downloads))
		progressSurface, err := font.RenderUTF8Solid(progressText, sdl.Color{R: 200, G: 200, B: 200, A: 255})
		if err == nil {
			progressTexture, err := renderer.CreateTextureFromSurface(progressSurface)
			if err == nil {
				progressRect := &sdl.Rect{
					X: (dm.window.Width - progressSurface.W) / 2,
					Y: dm.window.Height/4 + 30,
					W: progressSurface.W,
					H: progressSurface.H,
				}
				renderer.Copy(progressTexture, nil, progressRect)
				progressTexture.Destroy()
			}
			progressSurface.Free()
		}
	}

	displayNameY := dm.progressBarY - 75

	if dm.currentIndex < len(dm.downloads) {
		font := internal.GetSmallFont()
		var displayText string
		if dm.downloads[dm.currentIndex].DisplayName != "" {
			displayText = dm.downloads[dm.currentIndex].DisplayName
		} else {
			displayText = filepath.Base(dm.downloads[dm.currentIndex].Location)
		}

		maxWidth := dm.window.Width * 3 / 4
		renderMultilineText(renderer, displayText, font, maxWidth, dm.window.Width/2, displayNameY, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	}

	if dm.downloadComplete && len(dm.downloads) == 1 {
		font := internal.GetSmallFont()
		var displayText string
		if dm.downloads[0].DisplayName != "" {
			displayText = dm.downloads[0].DisplayName
		} else {
			displayText = filepath.Base(dm.downloads[0].Location)
		}

		maxWidth := dm.window.Width * 3 / 4
		renderMultilineText(renderer, displayText, font, maxWidth, dm.window.Width/2, displayNameY, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	}

	renderer.SetDrawColor(50, 50, 50, 255)
	progressBarBg := sdl.Rect{
		X: dm.progressBarX,
		Y: dm.progressBarY,
		W: dm.progressBarWidth,
		H: dm.progressBarHeight,
	}
	renderer.FillRect(&progressBarBg)

	progressWidth := int32(float64(dm.progressBarWidth) * dm.downloadProgress)
	if progressWidth > 0 {
		renderer.SetDrawColor(100, 150, 255, 255)
		progressBarFill := sdl.Rect{
			X: dm.progressBarX,
			Y: dm.progressBarY,
			W: progressWidth,
			H: dm.progressBarHeight,
		}
		renderer.FillRect(&progressBarFill)
	}

	percentTextY := dm.progressBarY + dm.progressBarHeight + 30

	percentText := fmt.Sprintf("%.1f%%", dm.downloadProgress*100)
	if dm.totalSize > 0 {
		downloadedMB := float64(dm.downloadedSize) / 1048576.0
		totalMB := float64(dm.totalSize) / 1048576.0
		percentText = fmt.Sprintf("%.1f%% (%.2f MB / %.2f MB)", dm.downloadProgress*100, downloadedMB, totalMB)
	}

	percentSurface, err := font.RenderUTF8Solid(percentText, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err == nil {
		percentTexture, err := renderer.CreateTextureFromSurface(percentSurface)
		if err == nil {
			percentRect := &sdl.Rect{
				X: (dm.window.Width - percentSurface.W) / 2,
				Y: percentTextY,
				W: percentSurface.W,
				H: percentSurface.H,
			}
			renderer.Copy(percentTexture, nil, percentRect)
			percentTexture.Destroy()
		}
		percentSurface.Free()
	}

	if dm.downloadError != nil {
		errorText := fmt.Sprintf("%s", dm.downloadError.Error())
		errorSurface, err := font.RenderUTF8Solid(errorText, sdl.Color{R: 255, G: 100, B: 100, A: 255})
		if err == nil {
			errorTexture, err := renderer.CreateTextureFromSurface(errorSurface)
			if err == nil {
				errorRect := &sdl.Rect{
					X: (dm.window.Width - errorSurface.W) / 2,
					Y: percentTextY + 40,
					W: errorSurface.W,
					H: errorSurface.H,
				}
				renderer.Copy(errorTexture, nil, errorRect)
				errorTexture.Destroy()
			}
			errorSurface.Free()
		}
	}

	var helpText string

	if dm.downloadComplete {
		helpText = "Press Any Button To Continue"
	} else if len(dm.downloads) > 1 {
		helpText = "B: Cancel Download | Y: Cancel All"
	} else {
		helpText = "B: Cancel Download"
	}

	helpSurface, err := font.RenderUTF8Solid(helpText, sdl.Color{R: 180, G: 180, B: 180, A: 255})
	if err == nil {
		helpTexture, err := renderer.CreateTextureFromSurface(helpSurface)
		if err == nil {
			helpRect := &sdl.Rect{
				X: (dm.window.Width - helpSurface.W) / 2,
				Y: dm.window.Height - helpSurface.H - 20,
				W: helpSurface.W,
				H: helpSurface.H,
			}
			renderer.Copy(helpTexture, nil, helpRect)
			helpTexture.Destroy()
		}
		helpSurface.Free()
	}
}

type progressReader struct {
	reader     io.Reader
	onProgress func(bytesRead int64)
	bytesRead  int64
}

func (r *progressReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.bytesRead += int64(n)
	if r.onProgress != nil {
		r.onProgress(r.bytesRead)
	}
	return
}
