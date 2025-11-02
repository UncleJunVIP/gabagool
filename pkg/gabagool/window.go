package gabagool

import (
	"log/slog"
	"os"
	"sync"

	"github.com/UncleJunVIP/gabagool/pkg/gabagool/core"
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
		slog.Error("Failed to get display mode!", "error", err)
	}

	return initWindowWithSize(title, displayMode.W, displayMode.H, displayBackground)
}

func initWindowWithSize(title string, width, height int32, displayBackground bool) *Window {
	x, y := int32(0), int32(0)

	//if core.IsDevMode() {
	//	width = width - 400
	//	height = height - 400
	//}

	//if core.IsDevMode() {
	//	width = 1024
	//	height = 768
	//}

	//if core.IsDevMode() {
	//	width = 1280
	//	height = 720
	//}

	if core.IsDevMode() {
		width = 640
		height = 480
	}

	window, err := sdl.CreateWindow(title, x, y, width, height, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}

	GetLoggerInstance().Info("Window created successfully")

	numDrivers, err := sdl.GetNumRenderDrivers()

	if err != nil {
		GetLoggerInstance().Error("Failed to get render drivers!", "error", err)
		os.Exit(1)
	}

	slog.Info("Available render drivers", "count", numDrivers)
	for i := 0; i < numDrivers; i++ {
		info := &sdl.RendererInfo{}
		sdl.GetRenderDriverInfo(i, info)
		GetLoggerInstance().Info("Render driver", "index", i, "name", info.Name)
	}

	var renderer *sdl.Renderer
	var lastErr error

	// Try 1: Default renderer with no flags (uses best available)
	GetLoggerInstance().Info("Attempting default renderer (no flags)")
	renderer, lastErr = sdl.CreateRenderer(window, -1, 0)
	if renderer != nil {
		GetLoggerInstance().Info("Successfully created default renderer")
	} else {
		GetLoggerInstance().Warn("Default renderer failed", "error", lastErr)

		// Try 2: Accelerated with PRESENTVSYNC
		GetLoggerInstance().Info("Attempting accelerated renderer with PRESENTVSYNC")
		renderer, lastErr = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
		if renderer != nil {
			GetLoggerInstance().Info("Successfully created accelerated renderer with PRESENTVSYNC")
		} else {
			GetLoggerInstance().Warn("Accelerated renderer with PRESENTVSYNC failed", "error", lastErr)

			// Try 3: Accelerated without PRESENTVSYNC
			GetLoggerInstance().Info("Attempting accelerated renderer without PRESENTVSYNC")
			renderer, lastErr = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
			if renderer != nil {
				GetLoggerInstance().Info("Successfully created accelerated renderer")
			} else {
				GetLoggerInstance().Warn("Accelerated renderer failed", "error", lastErr)

				// Try 4: Software renderer (last resort)
				GetLoggerInstance().Info("Attempting software renderer (fallback)")
				renderer, lastErr = sdl.CreateRenderer(window, -1, sdl.RENDERER_SOFTWARE)
				if renderer != nil {
					GetLoggerInstance().Info("Successfully created software renderer")
				}
			}
		}
	}

	if renderer == nil {
		GetLoggerInstance().Error("Failed to create any renderer!", "final_error", lastErr)
		os.Exit(1)
	}

	rendererInfo, _ := renderer.GetInfo()
	GetLoggerInstance().Info("Renderer info",
		"name", rendererInfo.Name,
		"flags", rendererInfo.Flags,
		"num_texture_formats", rendererInfo.NumTextureFormats,
	)

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

	theme := core.GetTheme()

	bgTexture, err := img.LoadTexture(window.Renderer, theme.BackgroundImagePath)
	if err == nil {
		window.Background = bgTexture
	} else {
		window.Background = nil
	}
}

func (window *Window) closeWindow() {
	if !core.IsDevMode() {
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
