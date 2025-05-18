package gabagool

import (
	"github.com/veandco/go-sdl2/sdl"
	"time"
)

type ProcessMessageOptions struct {
	ShowBackground bool
}
type ProcessReturn struct {
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
}

func ProcessMessage(message string, options ProcessMessageOptions, fn func() (interface{}, error)) (ProcessReturn, error) {
	processor := &processMessage{
		window:       GetWindow(),
		showBG:       options.ShowBackground,
		message:      message,
		isProcessing: true,
	}

	result := ProcessReturn{
		Success: false,
		Result:  nil,
		Error:   nil,
	}

	window := GetWindow()
	renderer := window.Renderer

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
		}
		processor.render(renderer)
		renderer.Present()

		sdl.Delay(16)
	}

	return result, err
}

func (p *processMessage) render(renderer *sdl.Renderer) {

	if !p.showBG {
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()
	}

	font := fonts.smallFont

	maxWidth := p.window.Width * 3 / 4
	renderMultilineText(renderer, p.message, font, maxWidth, p.window.Width/2, p.window.Height/2, sdl.Color{R: 255, G: 255, B: 255, A: 255})
}
