package ui

import (
	"github.com/UncleJunVIP/gabagool/internal"
	"github.com/veandco/go-sdl2/sdl"
	"time"
)

// ProcessReturn contains the result of the blocking process
type ProcessReturn struct {
	Success bool
	Result  interface{}
	Error   error
}

// BlockingProcess manages a UI that shows a message while running a function in a goroutine
type BlockingProcess struct {
	window       *internal.Window
	message      string
	isProcessing bool
	completeTime time.Time
}

// NewBlockingProcess creates a UI that shows a message while a function runs
// and returns the result once complete
func NewBlockingProcess(message string, fn func() (interface{}, error)) (ProcessReturn, error) {
	processor := &BlockingProcess{
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

	// Setup and start the process
	processor.render(renderer)
	renderer.Present()

	// Create a channel for the function result
	resultChan := make(chan struct {
		success bool
		result  interface{}
		err     error
	}, 1)

	// Start the function in a goroutine
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

	// Main loop to keep displaying the message
	running := true
	functionComplete := false
	var err error

	for running {
		// Handle SDL events (just to keep the window responsive)
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			if _, ok := event.(*sdl.QuitEvent); ok {
				running = false
				err = sdl.GetError()
			}
		}

		// Check if function has completed
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
				// Continue waiting
			}
		} else {
			// Function is complete, check if we should close
			if time.Since(processor.completeTime) > 1*time.Second {
				running = false
			}
		}

		// Update display
		processor.render(renderer)
		renderer.Present()

		sdl.Delay(16) // ~60 FPS
	}

	return result, err
}

func (p *BlockingProcess) render(renderer *sdl.Renderer) {
	renderer.SetDrawColor(0, 0, 0, 255)
	renderer.Clear()

	font := internal.GetSmallFont()

	// Only render the message, centered vertically and horizontally
	maxWidth := p.window.Width * 3 / 4
	renderMultilineText(renderer, p.message, font, maxWidth, p.window.Width/2, p.window.Height/2, sdl.Color{R: 255, G: 255, B: 255, A: 255})
}
