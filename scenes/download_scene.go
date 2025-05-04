package scenes

import (
	"fmt"
	"io"
	"net/http"
	"nextui-sdl2/internal"
	"nextui-sdl2/ui"
	"os"
	"path/filepath"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

type Download struct {
	URL         string
	Location    string
	DisplayName string
}

type DownloadScene struct {
	window *internal.Window

	fileName         string
	displayName      string
	downloadProgress float64
	totalSize        int64
	downloadedSize   int64

	isDownloading    bool
	downloadComplete bool
	downloadError    error
	active           bool

	progressBarWidth  int32
	progressBarHeight int32
	progressBarX      int32
	progressBarY      int32

	cancelDownload chan struct{}
	downloadDone   chan bool

	// Store URL and destination for download
	downloadURL      string
	downloadLocation string

	// Store queue of downloads
	downloadQueue []Download
	currentIndex  int
}

func NewDownloadScene(window *internal.Window) *DownloadScene {
	// Calculate progress bar dimensions and position
	progressBarWidth := window.Width * 3 / 4 // 75% of screen width
	progressBarHeight := int32(30)
	progressBarX := (window.Width - progressBarWidth) / 2
	progressBarY := window.Height / 2

	return &DownloadScene{
		window:            window,
		progressBarWidth:  progressBarWidth,
		progressBarHeight: progressBarHeight,
		progressBarX:      progressBarX,
		progressBarY:      progressBarY,
		cancelDownload:    make(chan struct{}),
		downloadDone:      make(chan bool),
		active:            false,
		downloadQueue:     []Download{},
		currentIndex:      0,
	}
}

// AddDownload adds a download to the queue
func (s *DownloadScene) AddDownload(url, location, displayName string) {
	s.downloadQueue = append(s.downloadQueue, Download{
		URL:         url,
		Location:    location,
		DisplayName: displayName,
	})

	// If scene is active and not currently downloading, start the first download
	if s.active && !s.isDownloading && len(s.downloadQueue) == 1 {
		s.startNextDownload()
	}
}

// SetDownloads replaces the entire download queue
func (s *DownloadScene) SetDownloads(downloads []Download) {
	s.downloadQueue = downloads
	s.currentIndex = 0

	// If scene is active and not currently downloading, start the first download
	if s.active && !s.isDownloading && len(s.downloadQueue) > 0 {
		s.startNextDownload()
	}
}

// startNextDownload starts the next download in the queue
func (s *DownloadScene) startNextDownload() {
	if s.currentIndex >= len(s.downloadQueue) {
		return
	}

	download := s.downloadQueue[s.currentIndex]
	s.StartDownload(download.URL, download.Location, download.DisplayName)
}

func (s *DownloadScene) StartDownload(url, fileName, displayName string) {
	if s.isDownloading {
		// Cancel current download if one is in progress
		close(s.cancelDownload)
		s.cancelDownload = make(chan struct{})
	}

	s.fileName = fileName
	s.displayName = displayName
	s.downloadURL = url
	s.downloadLocation = fileName
	s.downloadProgress = 0
	s.totalSize = 0
	s.downloadedSize = 0
	s.isDownloading = true
	s.downloadComplete = false // Ensure this is reset
	s.downloadError = nil

	// Start download in a goroutine
	go s.downloadFile(url, fileName)
}

func (s *DownloadScene) downloadFile(url, fileName string) {
	dir := filepath.Dir(fileName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		s.downloadError = err
		s.isDownloading = false
		s.downloadDone <- false
		return
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.downloadError = err
		s.isDownloading = false
		s.downloadDone <- false
		return
	}

	// Execute request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		s.downloadError = err
		s.isDownloading = false
		s.downloadDone <- false
		return
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		s.downloadError = fmt.Errorf("bad status: %s", resp.Status)
		s.isDownloading = false
		s.downloadDone <- false
		return
	}

	// Get file size for progress calculation
	s.totalSize = resp.ContentLength

	// Create output file
	out, err := os.Create(fileName)
	if err != nil {
		s.downloadError = err
		s.isDownloading = false
		s.downloadDone <- false
		return
	}
	defer out.Close()

	// Create a custom reader to track progress
	reader := &progressReader{
		reader: resp.Body,
		onProgress: func(bytesRead int64) {
			s.downloadedSize = bytesRead
			if s.totalSize > 0 {
				s.downloadProgress = float64(bytesRead) / float64(s.totalSize)
			}
		},
	}

	// Copy data with progress updates and cancellation support
	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(out, reader)
		done <- err
	}()

	// Inside the downloadFile method, replace the select/case block with this:
	select {
	case err := <-done:
		if err != nil {
			s.downloadError = err
			s.isDownloading = false
			s.downloadDone <- false
		} else {
			s.downloadComplete = true
			s.isDownloading = false
			s.downloadDone <- true

			// Log the download completion
			internal.Logger.Info("Download complete",
				"index", s.currentIndex,
				"file", s.fileName)

			// Start the next download directly from here
			if s.currentIndex < len(s.downloadQueue)-1 {
				s.currentIndex++
				// Start the next download directly
				nextDownload := s.downloadQueue[s.currentIndex]
				internal.Logger.Info("Starting next download directly",
					"index", s.currentIndex,
					"file", nextDownload.Location)

				// Reset state for next download
				s.downloadComplete = false
				s.StartDownload(nextDownload.URL, nextDownload.Location, nextDownload.DisplayName)
			}
		}
	case <-s.cancelDownload:
		// Download cancelled
		s.downloadError = fmt.Errorf("download cancelled")
		s.isDownloading = false
		s.downloadDone <- false
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

func (s *DownloadScene) Init() error {
	s.fileName = ""
	s.displayName = ""
	s.downloadURL = ""
	s.downloadLocation = ""
	s.downloadProgress = 0
	s.totalSize = 0
	s.downloadedSize = 0
	s.isDownloading = false
	s.downloadComplete = false
	s.downloadError = nil
	s.downloadQueue = []Download{}
	s.currentIndex = 0

	// Reset channels
	s.cancelDownload = make(chan struct{})
	s.downloadDone = make(chan bool)

	return nil
}

func (s *DownloadScene) Activate() error {
	s.active = true

	// If there are downloads in the queue, start the first one
	if !s.isDownloading && !s.downloadComplete && len(s.downloadQueue) > 0 && s.currentIndex < len(s.downloadQueue) {
		s.startNextDownload()
	} else if !s.isDownloading && !s.downloadComplete && s.downloadURL != "" && s.downloadLocation != "" {
		// Legacy support for old method
		s.StartDownload(s.downloadURL, s.downloadLocation, filepath.Base(s.downloadLocation))
	}

	return nil
}

func (s *DownloadScene) Deactivate() error {
	s.active = false
	return nil
}

func (s *DownloadScene) Update() error {
	// Only update if the scene is active
	if !s.active {
		return nil
	}

	// If we're not downloading but have a completed download and there are more items to download
	if !s.isDownloading && s.downloadComplete && s.currentIndex < len(s.downloadQueue)-1 {
		// Log the state before starting next download
		internal.Logger.Info("Starting next download",
			"current", s.currentIndex,
			"total", len(s.downloadQueue),
			"completed", s.downloadComplete)

		// Reset downloadComplete flag
		s.downloadComplete = false

		// Move to next download
		s.currentIndex++

		// Start the next download
		s.startNextDownload()
	}

	return nil
}

func (s *DownloadScene) Render() error {
	// Only render if the scene is active
	if !s.active {
		return nil
	}

	renderer := s.window.Renderer

	// Clear screen
	renderer.SetDrawColor(20, 20, 20, 255)
	renderer.Clear()

	// Draw title
	titleFont := internal.GetTitleFont()
	titleText := "Download Manager"
	titleSurface, err := titleFont.RenderUTF8Solid(titleText, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err == nil {
		titleTexture, err := renderer.CreateTextureFromSurface(titleSurface)
		if err == nil {
			titleRect := &sdl.Rect{
				X: (s.window.Width - titleSurface.W) / 2,
				Y: s.window.Height/4 - titleSurface.H/2,
				W: titleSurface.W,
				H: titleSurface.H,
			}
			renderer.Copy(titleTexture, nil, titleRect)
			titleTexture.Destroy()
		}
		titleSurface.Free()
	}

	// Draw progress info (current/total downloads)
	if len(s.downloadQueue) > 0 {
		font := internal.GetSmallFont()
		progressText := fmt.Sprintf("Download %d of %d", s.currentIndex+1, len(s.downloadQueue))
		progressSurface, err := font.RenderUTF8Solid(progressText, sdl.Color{R: 200, G: 200, B: 200, A: 255})
		if err == nil {
			progressTexture, err := renderer.CreateTextureFromSurface(progressSurface)
			if err == nil {
				progressRect := &sdl.Rect{
					X: (s.window.Width - progressSurface.W) / 2,
					Y: s.window.Height/4 + 30,
					W: progressSurface.W,
					H: progressSurface.H,
				}
				renderer.Copy(progressTexture, nil, progressRect)
				progressTexture.Destroy()
			}
			progressSurface.Free()
		}
	}

	// Draw file name
	if s.fileName != "" {
		font := internal.GetFont()
		var displayText string
		if s.displayName != "" {
			displayText = s.displayName
		} else {
			displayText = filepath.Base(s.fileName)
		}

		fileNameSurface, err := font.RenderUTF8Solid(displayText, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		if err == nil {
			fileNameTexture, err := renderer.CreateTextureFromSurface(fileNameSurface)
			if err == nil {
				fileNameRect := &sdl.Rect{
					X: (s.window.Width - fileNameSurface.W) / 2,
					Y: s.progressBarY - 60,
					W: fileNameSurface.W,
					H: fileNameSurface.H,
				}
				renderer.Copy(fileNameTexture, nil, fileNameRect)
				fileNameTexture.Destroy()
			}
			fileNameSurface.Free()
		}
	}

	// Draw progress bar background
	renderer.SetDrawColor(50, 50, 50, 255)
	progressBarBg := &sdl.Rect{
		X: s.progressBarX,
		Y: s.progressBarY,
		W: s.progressBarWidth,
		H: s.progressBarHeight,
	}
	renderer.FillRect(progressBarBg)

	// Draw progress bar fill
	if s.isDownloading || s.downloadComplete {
		fillWidth := int32(float64(s.progressBarWidth) * s.downloadProgress)

		// Choose color based on download state
		if s.downloadComplete {
			renderer.SetDrawColor(0, 255, 0, 255) // Green for complete
		} else {
			renderer.SetDrawColor(0, 150, 255, 255) // Blue for in progress
		}

		progressBarFill := &sdl.Rect{
			X: s.progressBarX,
			Y: s.progressBarY,
			W: fillWidth,
			H: s.progressBarHeight,
		}
		renderer.FillRect(progressBarFill)
	}

	// Draw progress percentage
	if s.isDownloading || s.downloadComplete {
		font := internal.GetSmallFont()

		var progressText string
		if s.totalSize > 0 {
			progressText = fmt.Sprintf("%.1f%%", s.downloadProgress*100)
			// Add downloaded/total size
			downloadedMB := float64(s.downloadedSize) / 1024 / 1024
			totalMB := float64(s.totalSize) / 1024 / 1024
			progressText += fmt.Sprintf(" (%.1f/%.1f MB)", downloadedMB, totalMB)
		} else {
			progressText = fmt.Sprintf("%.1f KB", float64(s.downloadedSize)/1024)
		}

		progressSurface, err := font.RenderUTF8Solid(progressText, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		if err == nil {
			progressTexture, err := renderer.CreateTextureFromSurface(progressSurface)
			if err == nil {
				progressRect := &sdl.Rect{
					X: (s.window.Width - progressSurface.W) / 2,
					Y: s.progressBarY + s.progressBarHeight + 10,
					W: progressSurface.W,
					H: progressSurface.H,
				}
				renderer.Copy(progressTexture, nil, progressRect)
				progressTexture.Destroy()
			}
			progressSurface.Free()
		}
	}

	// Draw status message
	statusText := ""
	statusColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}

	if s.downloadComplete {
		statusText = "Download Complete!"
		statusColor = sdl.Color{R: 0, G: 255, B: 0, A: 255}
	} else if s.downloadError != nil {
		if s.downloadError.Error() == "download cancelled" {
			statusText = "Download Canceled"
			statusColor = sdl.Color{R: 255, G: 140, B: 0, A: 255} // Orange for canceled
		} else {
			statusText = "Download Failed"
			statusColor = sdl.Color{R: 255, G: 0, B: 0, A: 255}
		}
	}

	if statusText != "" {
		font := internal.GetFont()
		statusSurface, err := font.RenderUTF8Solid(statusText, statusColor)
		if err == nil {
			statusTexture, err := renderer.CreateTextureFromSurface(statusSurface)
			if err == nil {
				statusRect := &sdl.Rect{
					X: (s.window.Width - statusSurface.W) / 2,
					Y: s.progressBarY + s.progressBarHeight + 50,
					W: statusSurface.W,
					H: statusSurface.H,
				}
				renderer.Copy(statusTexture, nil, statusRect)
				statusTexture.Destroy()
			}
			statusSurface.Free()
		}
	}

	// Draw instructions based on download state
	instructionsText := ""
	if s.isDownloading {
		instructionsText = "Press B to cancel download"
	} else if s.downloadComplete {
		instructionsText = "Press B to continue"
	} else {
		instructionsText = "Press B to continue"
	}

	font := internal.GetSmallFont()
	instructionsSurface, err := font.RenderUTF8Solid(instructionsText, sdl.Color{R: 200, G: 200, B: 200, A: 255})
	if err == nil {
		instructionsTexture, err := renderer.CreateTextureFromSurface(instructionsSurface)
		if err == nil {
			instructionsRect := &sdl.Rect{
				X: (s.window.Width - instructionsSurface.W) / 2,
				Y: s.window.Height - instructionsSurface.H - 30,
				W: instructionsSurface.W,
				H: instructionsSurface.H,
			}
			renderer.Copy(instructionsTexture, nil, instructionsRect)
			instructionsTexture.Destroy()
		}
		instructionsSurface.Free()
	}

	return nil
}

func (s *DownloadScene) HandleEvent(event sdl.Event) bool {
	if !s.active {
		return false
	}

	switch e := event.(type) {
	case *sdl.KeyboardEvent:
		if e.Type == sdl.KEYDOWN {
			switch e.Keysym.Sym {
			case sdl.K_ESCAPE:
				// Cancel download on ESC key if downloading
				if s.isDownloading {
					close(s.cancelDownload)
					s.cancelDownload = make(chan struct{})
					return true
				} else {
					// Go back to main menu if not downloading
					internal.GetSceneManager().SwitchTo("mainMenu")
					return true
				}
			case sdl.K_b: // Add keyboard B key support
				if s.isDownloading {
					close(s.cancelDownload)
					s.cancelDownload = make(chan struct{})
					return true
				} else if s.downloadComplete {
					internal.GetSceneManager().SwitchTo("mainMenu")
					return true
				}
			}
		}

	case *sdl.ControllerButtonEvent:
		if e.Type == sdl.CONTROLLERBUTTONDOWN {
			// Handle controller buttons
			switch e.Button {
			case ui.BrickButton_B: // B button (cancel/back)
				if s.isDownloading {
					close(s.cancelDownload)
					s.cancelDownload = make(chan struct{})
					return true
				} else if s.downloadComplete {
					internal.GetSceneManager().SwitchTo("mainMenu")
					return true
				} else {
					internal.GetSceneManager().SwitchTo("mainMenu")
					return true
				}
			}
		}
	}

	return false
}

func (s *DownloadScene) Destroy() error {
	return nil
}
