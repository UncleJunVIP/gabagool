package ui

import (
	"github.com/UncleJunVIP/gabagool/internal"
	"github.com/veandco/go-sdl2/sdl"
	"time"
)

type ProcessReturn struct {
	Success bool
	Result  interface{}
	Error   error
}

type blockingProcess struct {
	window       *internal.Window
	message      string
	isProcessing bool
	completeTime time.Time
}

func BlockingProcess(message string, fn func() (interface{}, error)) (ProcessReturn, error) {
	processor := &blockingProcess{
		window:       internal.GetWindow(),
		message:      message,
		isProcessing: true,
	}

	result := ProcessReturn{
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

			if time.Since(processor.completeTime) > 1*time.Second {
				running = false
			}
		}

		processor.render(renderer)
		renderer.Present()

		sdl.Delay(16)
	}

	return result, err
}

func (p *blockingProcess) render(renderer *sdl.Renderer) {
	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.Clear()

	font := internal.GetMediumFont()

	maxWidth := p.window.Width * 3 / 4
	renderMultilineText(renderer, p.message, font, maxWidth, p.window.Width/2, p.window.Height/2, sdl.Color{R: 255, G: 255, B: 255, A: 255})
}
