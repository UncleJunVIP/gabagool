package gabagool

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/UncleJunVIP/gabagool/pkg/gabagool/internal"
	"github.com/veandco/go-sdl2/ttf"

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
	window             *internal.Window
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

	scrollOffset int32

	headers       map[string]string
	lastInputTime time.Time
	inputDelay    time.Duration
}

func newDownloadManager(downloads []Download, headers map[string]string) *downloadManager {
	window := internal.GetWindow()

	// Calculate responsive progress bar width
	responsiveBarWidth := window.GetWidth() * 3 / 4
	if responsiveBarWidth > 900 {
		responsiveBarWidth = 900
	}
	progressBarHeight := int32(30)
	progressBarX := (window.GetWidth() - responsiveBarWidth) / 2

	return &downloadManager{
		window:             window,
		downloads:          downloads,
		downloadQueue:      []*downloadJob{},
		activeJobs:         []*downloadJob{},
		completedDownloads: []Download{},
		failedDownloads:    []Download{},
		errors:             []error{},
		isAllComplete:      false,
		maxActiveJobs:      3,
		headers:            headers,
		progressBarWidth:   responsiveBarWidth,
		progressBarHeight:  progressBarHeight,
		progressBarX:       progressBarX,
		scrollOffset:       0,
		lastInputTime:      time.Now(),
		inputDelay:         internal.DefaultInputDelay,
	}
}

func DownloadManager(downloads []Download, headers map[string]string, autoContinue bool) (DownloadReturn, error) {
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

	window := internal.GetWindow()
	renderer := window.Renderer
	processor := internal.GetInputProcessor()

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
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
				err = sdl.GetError()
				downloadManager.cancelAllDownloads()
				result.Cancelled = true

			case *sdl.KeyboardEvent, *sdl.ControllerButtonEvent, *sdl.ControllerAxisEvent, *sdl.JoyButtonEvent, *sdl.JoyAxisEvent, *sdl.JoyHatEvent:
				inputEvent := processor.ProcessSDLEvent(event.(sdl.Event))
				if inputEvent == nil || !inputEvent.Pressed {
					continue
				}

				if !downloadManager.isInputAllowed() {
					continue
				}
				downloadManager.lastInputTime = time.Now()

				result.LastPressedKey = sdl.Keycode(inputEvent.RawCode)
				result.LastPressedBtn = uint8(inputEvent.RawCode)

				if downloadManager.isAllComplete {
					running = false
					continue
				}

				// Cancel on Y button
				if inputEvent.Button == internal.VirtualButtonY {
					downloadManager.cancelAllDownloads()
					result.Cancelled = true
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

			if autoContinue && len(downloadManager.failedDownloads) == 0 {
				running = false
				continue
			}
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

func (dm *downloadManager) isInputAllowed() bool {
	return time.Since(dm.lastInputTime) >= dm.inputDelay
}

func (dm *downloadManager) startNextDownloads() {
	availableSlots := dm.maxActiveJobs - len(dm.activeJobs)
	if availableSlots <= 0 {
		return
	}

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
		Timeout: 120 * time.Minute,
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
	surface, _ := font.RenderUTF8Blended(filename, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if surface == nil {
		return filename
	}
	defer surface.Free()

	if surface.W <= maxWidth {
		return filename
	}

	ellipsis := "..."
	for len(filename) > 5 {
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

	font := internal.Fonts.SmallFont
	windowWidth := dm.window.GetWidth()
	windowHeight := dm.window.GetHeight()

	// No title, no footer - use full screen for content
	contentAreaStart := int32(20)
	contentAreaHeight := windowHeight - 20

	if len(dm.activeJobs) == 0 && dm.isAllComplete {
		var completeColor sdl.Color
		var completeText string

		var downloadText string
		if len(dm.downloads) > 1 {
			downloadText = "All Downloads"
		} else {
			downloadText = "Download"
		}

		if dm.failedDownloads != nil && len(dm.failedDownloads) > 0 {
			if dm.errors != nil && len(dm.errors) > 0 && dm.errors[0] != nil && dm.errors[0].Error() != "download cancelled by user" {
				completeText = fmt.Sprintf("%s Failed!", downloadText)
				completeColor = sdl.Color{R: 255, G: 0, B: 0, A: 255}
			} else {
				completeText = fmt.Sprintf("%s Canceled!", downloadText)
				completeColor = sdl.Color{R: 255, G: 0, B: 0, A: 255}
			}
		} else {
			completeText = fmt.Sprintf("%s Completed!", downloadText)
			completeColor = sdl.Color{R: 100, G: 255, B: 100, A: 255}
		}

		completeSurface, err := font.RenderUTF8Blended(completeText, completeColor)
		if err == nil && completeSurface != nil {
			completeTexture, err := renderer.CreateTextureFromSurface(completeSurface)
			if err == nil {
				centerY := (windowHeight - completeSurface.H) / 2
				completeRect := &sdl.Rect{
					X: (windowWidth - completeSurface.W) / 2,
					Y: centerY,
					W: completeSurface.W,
					H: completeSurface.H,
				}
				renderer.Copy(completeTexture, nil, completeRect)
				completeTexture.Destroy()
			}
			completeSurface.Free()
		}
	} else {
		// Measure text heights
		maxFilenameSurface, _ := font.RenderUTF8Blended("Sample", sdl.Color{R: 255, G: 255, B: 255, A: 255})
		filenameHeight := int32(0)
		if maxFilenameSurface != nil {
			filenameHeight = maxFilenameSurface.H
			maxFilenameSurface.Free()
		}

		spacingBetweenFilenameAndBar := int32(5)
		spacingBetweenDownloads := int32(25)

		// More compact: filename + bar only (no separate percentage below)
		singleDownloadHeight := filenameHeight + spacingBetweenFilenameAndBar + dm.progressBarHeight

		if len(dm.activeJobs) > 0 {
			// Check if we have no queued downloads and 1-3 active jobs
			hasNoQueue := len(dm.downloadQueue) == 0

			if hasNoQueue && len(dm.activeJobs) <= 3 {
				// Center 1-3 downloads vertically when no queue
				footerHeight := int32(80)
				availableHeight := contentAreaHeight - footerHeight

				totalHeight := int32(len(dm.activeJobs))*singleDownloadHeight + int32(len(dm.activeJobs)-1)*spacingBetweenDownloads
				startY := contentAreaStart + (availableHeight-totalHeight)/2
				if startY < contentAreaStart {
					startY = contentAreaStart + 10
				}

				for i, job := range dm.activeJobs {
					itemY := startY + int32(i)*(singleDownloadHeight+spacingBetweenDownloads)
					dm.renderDownloadItem(renderer, job, windowWidth, itemY, filenameHeight, spacingBetweenFilenameAndBar)
				}
			} else {
				// Multiple downloads with queue - use the multi-download layout
				dm.renderMultipleDownloads(renderer, windowWidth, contentAreaStart, contentAreaHeight, filenameHeight, spacingBetweenFilenameAndBar, spacingBetweenDownloads, singleDownloadHeight)
			}
		}
	}

	var footerHelpItems []FooterHelpItem
	if dm.isAllComplete {
		footerHelpItems = append(footerHelpItems, FooterHelpItem{ButtonName: "A", HelpText: "CloseLogger"})
	} else {
		helpText := "Cancel Download"
		if len(dm.downloads) > 1 {
			helpText = "Cancel All Downloads"
		}
		footerHelpItems = append(footerHelpItems, FooterHelpItem{ButtonName: "Y", HelpText: helpText})
	}

	renderFooter(renderer, internal.Fonts.SmallFont, footerHelpItems, 20, true)
}

func (dm *downloadManager) renderMultipleDownloads(renderer *sdl.Renderer, windowWidth int32, contentAreaStart int32, contentAreaHeight int32, filenameHeight int32, spacingBetweenFilenameAndBar int32, spacingBetweenDownloads int32, singleDownloadHeight int32) {
	// Always show max 3 concurrent downloads
	maxVisibleDownloads := 3

	// Measure remaining text height if needed
	remainingTextHeight := int32(0)
	totalRemaining := len(dm.activeJobs) - maxVisibleDownloads + len(dm.downloadQueue)
	if totalRemaining > 0 {
		remainingSurface, _ := internal.Fonts.SmallFont.RenderUTF8Blended("Sample", sdl.Color{R: 150, G: 150, B: 150, A: 255})
		if remainingSurface != nil {
			remainingTextHeight = remainingSurface.H + 15 // Text height + spacing
			remainingSurface.Free()
		}
	}

	// Calculate total height needed for visible downloads + remaining text
	totalHeight := int32(maxVisibleDownloads)*singleDownloadHeight + int32(maxVisibleDownloads-1)*spacingBetweenDownloads + remainingTextHeight

	// Reserve space for footer and center content
	footerHeight := int32(80)
	availableHeight := contentAreaHeight - footerHeight

	// Center downloads vertically within available space
	startY := contentAreaStart + (availableHeight-totalHeight)/2
	if startY < contentAreaStart {
		startY = contentAreaStart + 10 // Don't go above content area
	}

	// Render up to 3 active downloads
	renderCount := 0
	for _, job := range dm.activeJobs {
		if renderCount >= maxVisibleDownloads {
			break
		}

		itemY := startY + int32(renderCount)*(singleDownloadHeight+spacingBetweenDownloads)
		dm.renderDownloadItem(renderer, job, windowWidth, itemY, filenameHeight, spacingBetweenFilenameAndBar)
		renderCount++
	}

	// Show remaining downloads info
	if totalRemaining > 0 {
		remainingText := fmt.Sprintf("%d Additional Download%s Queued", totalRemaining, func() string {
			if totalRemaining == 1 {
				return ""
			}
			return "s"
		}())

		remainingSurface, err := internal.Fonts.SmallFont.RenderUTF8Blended(remainingText, sdl.Color{R: 150, G: 150, B: 150, A: 255})
		if err == nil && remainingSurface != nil {
			remainingTexture, err := renderer.CreateTextureFromSurface(remainingSurface)
			if err == nil {
				remainingY := startY + int32(maxVisibleDownloads)*(singleDownloadHeight+spacingBetweenDownloads) + 10
				remainingRect := &sdl.Rect{
					X: (windowWidth - remainingSurface.W) / 2,
					Y: remainingY,
					W: remainingSurface.W,
					H: remainingSurface.H,
				}
				renderer.Copy(remainingTexture, nil, remainingRect)
				remainingTexture.Destroy()
			}
			remainingSurface.Free()
		}
	}
}

func (dm *downloadManager) renderDownloadItem(renderer *sdl.Renderer, job *downloadJob, windowWidth int32, startY int32, filenameHeight int32, spacingBetweenFilenameAndBar int32) {
	font := internal.Fonts.SmallFont

	// Get display name with truncation
	var displayText string
	if job.download.DisplayName != "" {
		displayText = job.download.DisplayName
	} else {
		displayText = filepath.Base(job.download.Location)
	}

	// Truncate filename to fit responsively
	maxWidth := windowWidth * 3 / 4
	if maxWidth > 900 {
		maxWidth = 900
	}
	displayText = truncateFilename(displayText, maxWidth, font)

	filenameSurface, err := font.RenderUTF8Blended(displayText, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err == nil && filenameSurface != nil {
		filenameTexture, err := renderer.CreateTextureFromSurface(filenameSurface)
		if err == nil {
			filenameRect := &sdl.Rect{
				X: (windowWidth - filenameSurface.W) / 2,
				Y: startY,
				W: filenameSurface.W,
				H: filenameSurface.H,
			}
			renderer.Copy(filenameTexture, nil, filenameRect)
			filenameTexture.Destroy()
		}
		filenameSurface.Free()
	}

	// Calculate progress bar Y position
	progressBarY := startY + filenameHeight + spacingBetweenFilenameAndBar

	// Render progress bar background
	renderer.SetDrawColor(50, 50, 50, 255)
	progressBarBg := sdl.Rect{
		X: dm.progressBarX,
		Y: progressBarY,
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
			Y: progressBarY,
			W: progressWidth,
			H: dm.progressBarHeight,
		}
		renderer.FillRect(&progressBarFill)
	}

	// Render percentage text INSIDE the progress bar
	percentText := fmt.Sprintf("%.0f%%", job.progress*100)
	if job.totalSize > 0 {
		downloadedMB := float64(job.downloadedSize) / 1048576.0
		totalMB := float64(job.totalSize) / 1048576.0
		percentText = fmt.Sprintf("%.0f%% (%.1fMB/%.1fMB)", job.progress*100, downloadedMB, totalMB)
	}

	percentSurface, err := font.RenderUTF8Blended(percentText, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err == nil && percentSurface != nil {
		percentTexture, err := renderer.CreateTextureFromSurface(percentSurface)
		if err == nil {
			// Center text inside progress bar
			textX := dm.progressBarX + (dm.progressBarWidth-percentSurface.W)/2
			textY := progressBarY + (dm.progressBarHeight-percentSurface.H)/2

			percentRect := &sdl.Rect{
				X: textX,
				Y: textY,
				W: percentSurface.W,
				H: percentSurface.H,
			}
			renderer.Copy(percentTexture, nil, percentRect)
			percentTexture.Destroy()
		}
		percentSurface.Free()
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
