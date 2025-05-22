package gabagool

import (
	"fmt"
	"github.com/veandco/go-sdl2/ttf"
	"io"
	"net/http"
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

type DownloadReturn struct {
	CompletedDownloads []Download
	FailedDownloads    []Download
	Errors             []error
	LastPressedKey     sdl.Keycode
	LastPressedBtn     uint8
	Cancelled          bool
}

type downloadJob struct {
	download       Download
	progress       float64
	totalSize      int64
	downloadedSize int64
	isComplete     bool
	hasError       bool
	error          error
	cancelChan     chan struct{}
}

type downloadManager struct {
	window             *Window
	downloads          []Download
	downloadQueue      []*downloadJob
	activeJobs         []*downloadJob
	completedDownloads []Download
	failedDownloads    []Download
	errors             []error
	isAllComplete      bool
	maxActiveJobs      int
	cancellationError  error

	progressBarWidth  int32
	progressBarHeight int32
	progressBarX      int32

	headers map[string]string
}

func newDownloadManager(downloads []Download, headers map[string]string) *downloadManager {
	window := GetWindow()

	progressBarWidth := window.Width * 3 / 4
	progressBarHeight := int32(30)
	progressBarX := (window.Width - progressBarWidth) / 2

	return &downloadManager{
		window:             window,
		downloads:          downloads,
		downloadQueue:      []*downloadJob{},
		activeJobs:         []*downloadJob{},
		completedDownloads: []Download{},
		failedDownloads:    []Download{},
		errors:             []error{},
		isAllComplete:      false,
		maxActiveJobs:      4,
		headers:            headers,
		progressBarWidth:   progressBarWidth,
		progressBarHeight:  progressBarHeight,
		progressBarX:       progressBarX,
	}
}

func DownloadManager(downloads []Download, headers map[string]string) (DownloadReturn, error) {
	downloadManager := newDownloadManager(downloads, headers)

	result := DownloadReturn{
		CompletedDownloads: []Download{},
		FailedDownloads:    []Download{},
		Errors:             []error{},
		LastPressedKey:     0,
		LastPressedBtn:     0,
		Cancelled:          false,
	}

	if len(downloads) == 0 {
		return result, nil
	}

	window := GetWindow()
	renderer := window.Renderer

	// Initialize the download queue
	for _, download := range downloads {
		job := &downloadJob{
			download:   download,
			progress:   0,
			isComplete: false,
			hasError:   false,
			cancelChan: make(chan struct{}),
		}
		downloadManager.downloadQueue = append(downloadManager.downloadQueue, job)
	}

	// Start initial downloads
	downloadManager.startNextDownloads()

	downloadManager.render(renderer)
	renderer.Present()

	sdl.Delay(100)

	running := true
	var err error

	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
				err = sdl.GetError()
				downloadManager.cancelAllDownloads()
				result.Cancelled = true

			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {
					result.LastPressedKey = e.Keysym.Sym

					if downloadManager.isAllComplete {
						running = false
						continue
					}

					if e.Keysym.Sym == sdl.K_ESCAPE {
						downloadManager.cancelAllDownloads()
						result.Cancelled = true
					}
				}

			case *sdl.ControllerButtonEvent:
				if e.Type == sdl.CONTROLLERBUTTONDOWN {
					result.LastPressedBtn = e.Button

					if downloadManager.isAllComplete {
						running = false
						continue
					}

					if Button(e.Button) == ButtonY {
						downloadManager.cancelAllDownloads()
						result.Cancelled = true
					}
				}
			}
		}

		// Check for completed or failed downloads
		downloadManager.updateJobStatus()

		// Start new downloads if there's room
		if len(downloadManager.activeJobs) < downloadManager.maxActiveJobs && len(downloadManager.downloadQueue) > 0 {
			downloadManager.startNextDownloads()
		}

		// Check if all downloads are complete
		if len(downloadManager.activeJobs) == 0 && len(downloadManager.downloadQueue) == 0 && !downloadManager.isAllComplete {
			downloadManager.isAllComplete = true
		}

		downloadManager.render(renderer)
		renderer.Present()

		sdl.Delay(16)
	}

	result.CompletedDownloads = downloadManager.completedDownloads
	result.FailedDownloads = downloadManager.failedDownloads
	result.Errors = downloadManager.errors

	return result, err
}

func (dm *downloadManager) startNextDownloads() {
	availableSlots := dm.maxActiveJobs - len(dm.activeJobs)
	if availableSlots <= 0 {
		return
	}

	// Start as many downloads as we have slots for
	for i := 0; i < availableSlots && len(dm.downloadQueue) > 0; i++ {
		job := dm.downloadQueue[0]
		dm.downloadQueue = dm.downloadQueue[1:]
		dm.activeJobs = append(dm.activeJobs, job)

		go dm.downloadFile(job)
	}
}

func (dm *downloadManager) updateJobStatus() {
	remaining := []*downloadJob{}

	for _, job := range dm.activeJobs {
		if job.isComplete {
			dm.completedDownloads = append(dm.completedDownloads, job.download)
		} else if job.hasError {
			dm.failedDownloads = append(dm.failedDownloads, job.download)
			dm.errors = append(dm.errors, job.error)
		} else {
			remaining = append(remaining, job)
		}
	}

	dm.activeJobs = remaining
}

func (dm *downloadManager) cancelAllDownloads() {
	// Cancel all active downloads
	for _, job := range dm.activeJobs {
		close(job.cancelChan)
		if !job.isComplete && !job.hasError {
			job.hasError = true
			job.error = fmt.Errorf("download cancelled by user")
			dm.failedDownloads = append(dm.failedDownloads, job.download)
			dm.errors = append(dm.errors, job.error)
		}
	}

	// Mark all queued downloads as failed
	for _, job := range dm.downloadQueue {
		job.hasError = true
		job.error = fmt.Errorf("download cancelled by user")
		dm.failedDownloads = append(dm.failedDownloads, job.download)
		dm.errors = append(dm.errors, job.error)
	}

	dm.activeJobs = []*downloadJob{}
	dm.downloadQueue = []*downloadJob{}
	dm.isAllComplete = true
}

func (dm *downloadManager) downloadFile(job *downloadJob) {
	url := job.download.URL
	filePath := job.download.Location

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		job.hasError = true
		job.error = err
		return
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		job.hasError = true
		job.error = err
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
		job.hasError = true
		job.error = err
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		job.hasError = true
		job.error = fmt.Errorf("bad status: %s", resp.Status)
		return
	}

	job.totalSize = resp.ContentLength

	out, err := os.Create(filePath)
	if err != nil {
		job.hasError = true
		job.error = err
		return
	}
	defer out.Close()

	reader := &progressReader{
		reader: resp.Body,
		onProgress: func(bytesRead int64) {
			job.downloadedSize = bytesRead
			if job.totalSize > 0 {
				job.progress = float64(bytesRead) / float64(job.totalSize)
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
			job.hasError = true
			job.error = err
		} else {
			job.isComplete = true
		}
	case <-job.cancelChan:
		job.hasError = true
		job.error = fmt.Errorf("download canceled")
	}
}

func truncateFilename(filename string, maxWidth int32, font *ttf.Font) string {
	// Check if filename needs truncation
	surface, _ := font.RenderUTF8Blended(filename, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if surface == nil {
		return filename // If surface creation failed, return original
	}
	defer surface.Free()

	if surface.W <= maxWidth {
		return filename // No truncation needed
	}

	// Truncate with ellipsis
	ellipsis := "..."
	for len(filename) > 5 { // Keep at least 1 char + ellipsis
		filename = filename[:len(filename)-1]
		surface, _ := font.RenderUTF8Blended(filename+ellipsis, sdl.Color{R: 255, G: 255, B: 255, A: 255})
		if surface == nil {
			break
		}
		if surface.W <= maxWidth {
			surface.Free()
			return filename + ellipsis
		}
		surface.Free()
	}

	return filename + ellipsis
}

func (dm *downloadManager) render(renderer *sdl.Renderer) {
	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.Clear()

	font := fonts.tinyFont

	// Render title
	titleText := "Download Manager"
	titleSurface, err := fonts.mediumFont.RenderUTF8Blended(titleText, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err == nil {
		titleTexture, err := renderer.CreateTextureFromSurface(titleSurface)
		if err == nil {
			titleRect := &sdl.Rect{
				X: (dm.window.Width - titleSurface.W) / 2,
				Y: 30,
				W: titleSurface.W,
				H: titleSurface.H,
			}
			renderer.Copy(titleTexture, nil, titleRect)
			titleTexture.Destroy()
		}
		titleSurface.Free()
	}

	if len(dm.activeJobs) == 0 && dm.isAllComplete {
		// All downloads complete screen
		completeText := "All Downloads Complete"
		completeSurface, err := font.RenderUTF8Blended(completeText, sdl.Color{R: 100, G: 255, B: 100, A: 255})
		if err == nil {
			completeTexture, err := renderer.CreateTextureFromSurface(completeSurface)
			if err == nil {
				completeRect := &sdl.Rect{
					X: (dm.window.Width - completeSurface.W) / 2,
					Y: dm.window.Height/2 - completeSurface.H/2,
					W: completeSurface.W,
					H: completeSurface.H,
				}
				renderer.Copy(completeTexture, nil, completeRect)
				completeTexture.Destroy()
			}
			completeSurface.Free()
		}
	} else {
		// Render each active download
		titlePadding := int32(90)
		if titleSurface != nil {
			titlePadding = titleSurface.H
		}

		baseY := int32(150) + titlePadding
		spacing := dm.progressBarHeight + 100 // Increased spacing between downloads

		if len(dm.downloads) == 1 {
			baseY += dm.window.Height/5 + 75
			spacing = 0
		}

		for i, job := range dm.activeJobs {
			y := baseY + int32(i)*spacing

			// Get display name with truncation
			var displayText string
			if job.download.DisplayName != "" {
				displayText = job.download.DisplayName
			} else {
				displayText = filepath.Base(job.download.Location)
			}

			// Truncate filename to fit on a single line
			maxWidth := dm.window.Width * 3 / 4
			displayText = truncateFilename(displayText, maxWidth, font)

			// Render filename
			filenameSurface, err := font.RenderUTF8Blended(displayText, sdl.Color{R: 255, G: 255, B: 255, A: 255})
			if err == nil {
				filenameTexture, err := renderer.CreateTextureFromSurface(filenameSurface)
				if err == nil {
					filenameRect := &sdl.Rect{
						X: (dm.window.Width - filenameSurface.W) / 2,
						Y: y - 40,
						W: filenameSurface.W,
						H: filenameSurface.H,
					}
					renderer.Copy(filenameTexture, nil, filenameRect)
					filenameTexture.Destroy()
				}
				filenameSurface.Free()
			}

			// Render progress bar background
			renderer.SetDrawColor(50, 50, 50, 255)
			progressBarBg := sdl.Rect{
				X: dm.progressBarX,
				Y: y,
				W: dm.progressBarWidth,
				H: dm.progressBarHeight,
			}
			renderer.FillRect(&progressBarBg)

			// Render progress bar fill
			progressWidth := int32(float64(dm.progressBarWidth) * job.progress)
			if progressWidth > 0 {
				renderer.SetDrawColor(100, 150, 255, 255)
				progressBarFill := sdl.Rect{
					X: dm.progressBarX,
					Y: y,
					W: progressWidth,
					H: dm.progressBarHeight,
				}
				renderer.FillRect(&progressBarFill)
			}

			// Render percentage text
			percentText := fmt.Sprintf("%.1f%%", job.progress*100)
			if job.totalSize > 0 {
				downloadedMB := float64(job.downloadedSize) / 1048576.0
				totalMB := float64(job.totalSize) / 1048576.0
				percentText = fmt.Sprintf("%.1f%% (%.2f MB / %.2f MB)", job.progress*100, downloadedMB, totalMB)
			}

			percentSurface, err := font.RenderUTF8Blended(percentText, sdl.Color{R: 255, G: 255, B: 255, A: 255})
			if err == nil {
				percentTexture, err := renderer.CreateTextureFromSurface(percentSurface)
				if err == nil {
					percentRect := &sdl.Rect{
						X: (dm.window.Width - percentSurface.W) / 2,
						Y: y + dm.progressBarHeight + 10,
						W: percentSurface.W,
						H: percentSurface.H,
					}
					renderer.Copy(percentTexture, nil, percentRect)
					percentTexture.Destroy()
				}
				percentSurface.Free()
			}
		}

		// Display queue info
		if len(dm.downloadQueue) > 0 {
			queueText := fmt.Sprintf("Downloads in queue: %d", len(dm.downloadQueue))
			queueSurface, err := font.RenderUTF8Blended(queueText, sdl.Color{R: 180, G: 180, B: 180, A: 255})
			if err == nil {
				queueTexture, err := renderer.CreateTextureFromSurface(queueSurface)
				if err == nil {
					queueRect := &sdl.Rect{
						X: (dm.window.Width - queueSurface.W) / 2,
						Y: baseY + int32(len(dm.activeJobs))*spacing + 20,
						W: queueSurface.W,
						H: queueSurface.H,
					}
					renderer.Copy(queueTexture, nil, queueRect)
					queueTexture.Destroy()
				}
				queueSurface.Free()
			}
		}
	}

	// Render status text
	var statusText string
	if dm.isAllComplete {
		statusText = "Press Any Button To Continue"
	} else {
		statusText = "ESC: Cancel All Downloads"
	}

	statusSurface, err := font.RenderUTF8Blended(statusText, sdl.Color{R: 180, G: 180, B: 180, A: 255})
	if err == nil {
		statusTexture, err := renderer.CreateTextureFromSurface(statusSurface)
		if err == nil {
			statusRect := &sdl.Rect{
				X: (dm.window.Width - statusSurface.W) / 2,
				Y: dm.window.Height - statusSurface.H - 20,
				W: statusSurface.W,
				H: statusSurface.H,
			}
			renderer.Copy(statusTexture, nil, statusRect)
			statusTexture.Destroy()
		}
		statusSurface.Free()
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
