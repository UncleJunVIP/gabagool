package gabagool

import (
	"fmt"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"os"
)

type Window struct {
	Window            *sdl.Window
	Renderer          *sdl.Renderer
	Title             string
	Width             int32
	Height            int32
	FontSize          int
	SmallFontSize     int
	Background        *sdl.Texture
	DisplayBackground bool
}

func initWindow(title string, displayBackground bool) *Window {
	displayIndex := 0
	displayMode, err := sdl.GetCurrentDisplayMode(displayIndex)

	width := DefaultWindowWidth
	height := DefaultWindowHeight

	if IsDev {
		width = DefaultWindowWidth
		height = DefaultWindowHeight
	} else if err == nil {
		width = displayMode.W
		height = displayMode.H
	} else {
		fmt.Fprintf(os.Stderr, "Failed to get display mode: %s\n", err)
	}

	return initWindowWithSize(title, width, height, displayBackground)
}

func initWindowWithSize(title string, width, height int32, displayBackground bool) *Window {
	window, err := sdl.CreateWindow(title, 0, 0, width, height, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_TARGETTEXTURE|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		os.Exit(1)
	}

	initFonts(getFontScale(width, height))

	win := &Window{
		Window:            window,
		Renderer:          renderer,
		Title:             title,
		Width:             width,
		Height:            height,
		DisplayBackground: displayBackground,
	}

	win.loadBackground()

	return win
}

func (window *Window) loadBackground() {
	img.Init(img.INIT_PNG)

	bgPath := NextUIBackgroundPath

	if IsDev {
		bgPath = os.Getenv(EnvBackgroundPath)
	}

	bgTexture, err := img.LoadTexture(window.Renderer, bgPath)
	if err == nil {
		window.Background = bgTexture
	} else {
		window.Background = nil
	}
}

func (window Window) closeWindow() {
	if window.Background != nil {
		window.Background.Destroy()
	}
	window.Renderer.Destroy()
	window.Window.Destroy()

	img.Quit()
}

func GetWindow() *Window {
	return window
}

func (window *Window) RenderBackground() {
	if window.Background != nil {
		window.Renderer.Copy(window.Background, nil, &sdl.Rect{X: 0, Y: 0, W: window.Width, H: window.Height})
	}
}
