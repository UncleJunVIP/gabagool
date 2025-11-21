package gabagool

import (
	"fmt"
	"time"

	"github.com/UncleJunVIP/gabagool/pkg/gabagool/internal"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"go.uber.org/atomic"
)

type ProcessMessageOptions struct {
	Image               string
	ImageWidth          int32
	ImageHeight         int32
	ShowThemeBackground bool
	ShowProgressBar     bool
	Progress            *atomic.Float64
}

type ProcessMessageReturn struct {
	Success bool
	Result  interface{}
	Error   error
}

type processMessage struct {
	window          *internal.Window
	showBG          bool
	message         string
	isProcessing    bool
	completeTime    time.Time
	imageTexture    *sdl.Texture
	imageWidth      int32
	imageHeight     int32
	showProgressBar bool
	progress        *atomic.Float64
}

func ProcessMessage(message string, options ProcessMessageOptions, fn func() (interface{}, error)) (ProcessMessageReturn, error) {
	processor := &processMessage{
		window:          internal.GetWindow(),
		showBG:          options.ShowThemeBackground,
		imageWidth:      options.ImageWidth,
		imageHeight:     options.ImageHeight,
		message:         message,
		isProcessing:    true,
		showProgressBar: options.ShowProgressBar,
		progress:        options.Progress,
	}

	if options.Image != "" {
		img.Init(img.INIT_PNG)
		texture, err := img.LoadTexture(processor.window.Renderer, options.Image)
		if err == nil {
			processor.imageTexture = texture
		}
	}

	result := ProcessMessageReturn{
		Success: false,
		Result:  nil,
		Error:   nil,
	}

	window := internal.GetWindow()
	renderer := window.Renderer

	processor.render(renderer)
	renderer.Present()

	resultChan := make(chan struct {
		success bool
		result  interface{}
		err     error
	}, 1)

	go func() {
		result, err := fn()
		if err != nil {
			resultChan <- struct {
				success bool
				result  interface{}
				err     error
			}{success: false, result: nil, err: err}
		} else {
			resultChan <- struct {
				success bool
				result  interface{}
				err     error
			}{success: true, result: result, err: nil}
		}
	}()

	running := true
	functionComplete := false
	var err error

	for running {

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			if _, ok := event.(*sdl.QuitEvent); ok {
				running = false
				err = sdl.GetError()
			}
		}

		if !functionComplete {
			select {
			case processResult := <-resultChan:
				result.Success = processResult.success
				result.Result = processResult.result
				result.Error = processResult.err
				functionComplete = true
				processor.isProcessing = false
				processor.completeTime = time.Now()
			default:

			}
		} else {

			if time.Since(processor.completeTime) > 350*time.Millisecond {
				running = false
			}
		}

		// Render the process message with current progress
		processor.render(renderer)
		renderer.Present()

		sdl.Delay(16)
	}

	if processor.imageTexture != nil {
		processor.imageTexture.Destroy()
	}

	if result.Error != nil {
		// If the go routine returned an error, use this one
		return result, result.Error
	}

	// return the SDL error if any
	return result, err
}

func (p *processMessage) render(renderer *sdl.Renderer) {

	if p.showBG && internal.GetWindow().Background != nil {
		internal.GetWindow().RenderBackground()
	} else {
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()
	}

	if p.imageTexture != nil {
		width := p.imageWidth
		height := p.imageHeight

		if width == 0 {
			width = p.window.GetWidth()
		}

		if height == 0 {
			height = p.window.GetHeight()
		}

		x := (p.window.GetWidth() - width) / 2
		y := (p.window.GetHeight() - height) / 2

		renderer.Copy(p.imageTexture, nil, &sdl.Rect{X: x, Y: y, W: width, H: height})
	}

	font := internal.Fonts.SmallFont

	maxWidth := p.window.GetWidth() * 3 / 4
	internal.RenderMultilineText(renderer, p.message, font, maxWidth, p.window.GetWidth()/2, p.window.GetHeight()/2, sdl.Color{R: 255, G: 255, B: 255, A: 255})

	// Add progress bar if requested
	if p.showProgressBar {
		p.renderProgressBar(renderer)
	}
}

func (p *processMessage) renderProgressBar(renderer *sdl.Renderer) {
	// Calculate progress bar dimensions
	windowWidth := p.window.GetWidth()
	windowHeight := p.window.GetHeight()

	// Progress bar dimensions
	barWidth := windowWidth * 3 / 4
	if barWidth > 900 {
		barWidth = 900
	}
	barHeight := int32(40)
	barX := (windowWidth - barWidth) / 2
	barY := (windowHeight - barHeight + (int32(internal.Fonts.SmallFont.Height()) * 2)) / 2

	// Progress bar background
	renderer.SetDrawColor(50, 50, 50, 255)
	progressBarBg := sdl.Rect{
		X: barX,
		Y: barY,
		W: barWidth,
		H: barHeight,
	}
	renderer.FillRect(&progressBarBg)

	// Progress bar fill
	progressWidth := int32(float64(barWidth) * p.progress.Load())

	if progressWidth > 0 {
		renderer.SetDrawColor(100, 150, 255, 255)
		progressBarFill := sdl.Rect{
			X: barX,
			Y: barY,
			W: progressWidth,
			H: barHeight,
		}
		renderer.FillRect(&progressBarFill)
	}

	percentText := fmt.Sprintf("%.0f%%", p.progress.Load()*100)

	percentSurface, err := internal.Fonts.SmallFont.RenderUTF8Blended(percentText, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err == nil {
		percentTexture, err := renderer.CreateTextureFromSurface(percentSurface)
		if err == nil {
			// Center text inside progress bar
			textX := barX + (barWidth-percentSurface.W)/2
			textY := barY + (barHeight-percentSurface.H)/2

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
