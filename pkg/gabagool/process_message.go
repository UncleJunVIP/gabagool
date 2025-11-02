package gabagool

import (
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

type ProcessMessageOptions struct {
	Image               string
	ImageWidth          int32
	ImageHeight         int32
	ShowThemeBackground bool
}
type ProcessMessageReturn struct {
	Success bool
	Result  interface{}
	Error   error
}

type processMessage struct {
	window       *Window
	showBG       bool
	message      string
	isProcessing bool
	completeTime time.Time
	imageTexture *sdl.Texture
	imageWidth   int32
	imageHeight  int32
}

func ProcessMessage(message string, options ProcessMessageOptions, fn func() (interface{}, error)) (ProcessMessageReturn, error) {
	processor := &processMessage{
		window:       GetWindow(),
		showBG:       options.ShowThemeBackground,
		imageWidth:   options.ImageWidth,
		imageHeight:  options.ImageHeight,
		message:      message,
		isProcessing: true,
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

	window := GetWindow()
	renderer := window.Renderer

	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.Clear()

	if processor.showBG {
		window.RenderBackground()
	}

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

		if processor.showBG {
			window.RenderBackground()
		} else {
			renderer.SetDrawColor(0, 0, 0, 255)
			renderer.Clear()
		}

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

	if !p.showBG && p.imageTexture == nil {
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

	font := fonts.smallFont

	maxWidth := p.window.GetWidth() * 3 / 4
	renderMultilineText(renderer, p.message, font, maxWidth, p.window.GetWidth()/2, p.window.GetHeight()/2, sdl.Color{R: 255, G: 255, B: 255, A: 255})
}
