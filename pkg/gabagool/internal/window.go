package internal

import (
	"os"
	"sync"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

type Window struct {
	Window            *sdl.Window
	Renderer          *sdl.Renderer
	Title             string
	FontSize          int
	SmallFontSize     int
	Background        *sdl.Texture
	DisplayBackground bool
	PowerButtonWG     sync.WaitGroup
}

func initWindow(title string, displayBackground bool) *Window {
	displayIndex := 0
	displayMode, err := sdl.GetCurrentDisplayMode(displayIndex)

	if err != nil {
		GetLogger().Error("Failed to Get display mode!", "error", err)
	}

	return initWindowWithSize(title, displayMode.W, displayMode.H, displayBackground)
}

func initWindowWithSize(title string, width, height int32, displayBackground bool) *Window {
	x, y := int32(0), int32(0)

	if IsDevMode() {
		width = 1024
		height = 768
	}

	var windowFlags uint32
	windowFlags = sdl.WINDOW_SHOWN

	if IsDevMode() {
		windowFlags = windowFlags | sdl.WINDOW_BORDERLESS
	}

	window, err := sdl.CreateWindow(title, x, y, width, height, windowFlags)
	if err != nil {
		panic(err)
	}

	var renderer *sdl.Renderer
	var lastErr error

	renderer, lastErr = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC|sdl.RENDERER_TARGETTEXTURE)

	if lastErr != nil {
		GetLogger().Error("Failed to create any renderer!", "final_error", lastErr)
		os.Exit(1)
	}

	renderer.SetLogicalSize(width, height)

	win := &Window{
		Window:            window,
		Renderer:          renderer,
		Title:             title,
		DisplayBackground: displayBackground,
	}

	win.loadBackground()

	return win
}

func (window *Window) initPowerButtonHandling() {
	window.PowerButtonWG.Add(1)
	go powerButtonHandler(&window.PowerButtonWG)
}

func (window *Window) loadBackground() {
	img.Init(img.INIT_PNG)

	theme := GetTheme()

	bgTexture, err := img.LoadTexture(window.Renderer, theme.BackgroundImagePath)
	if err == nil {
		window.Background = bgTexture
	} else {
		window.Background = nil
	}
}

func (window *Window) closeWindow() {
	if !IsDevMode() {
		window.PowerButtonWG.Done()
	}

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

func (window *Window) GetWidth() int32 {
	w, _ := window.Window.GetSize()
	return w
}

func (window *Window) GetHeight() int32 {
	_, h := window.Window.GetSize()
	return h
}

func (window *Window) RenderBackground() {
	if window.Background != nil {
		window.Renderer.Copy(window.Background, nil, &sdl.Rect{X: 0, Y: 0, W: window.GetWidth(), H: window.GetHeight()})
	}
}

func ResetBackground() {
	window.loadBackground()
}
