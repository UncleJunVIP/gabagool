package gabagool

import (
	"log/slog"
	"os"
	"sync"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
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
	PowerButtonWG     sync.WaitGroup
}

func initWindow(title string, displayBackground bool) *Window {
	displayIndex := 0
	displayMode, err := sdl.GetCurrentDisplayMode(displayIndex)

	if err != nil {
		slog.Error("Failed to get display mode!", "error", err)
	}

	return initWindowWithSize(title, displayMode.W, displayMode.H, displayBackground)
}

func initWindowWithSize(title string, width, height int32, displayBackground bool) *Window {
	x, y := int32(0), int32(0)

	if IsDev {
		x = width/2 - DefaultWindowWidth/2
		y = height/2 - DefaultWindowHeight/2
		width = DefaultWindowWidth
		height = DefaultWindowHeight
	}

	window, err := sdl.CreateWindow(title, x, y, width, height, sdl.WINDOW_SHOWN|sdl.WINDOW_BORDERLESS)
	if err != nil {
		panic(err)
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_TARGETTEXTURE|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		slog.Error("Failed to create renderer!", "error", err)
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

func (window *Window) initPowerButtonHandling() {
	window.PowerButtonWG.Add(1)
	go powerButtonHandler(&window.PowerButtonWG)
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

func (window *Window) closeWindow() {
	window.PowerButtonWG.Done()

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
