package internal

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"os"
)

type Window struct {
	Window        *sdl.Window
	Renderer      *sdl.Renderer
	Title         string
	Width         int32
	Height        int32
	FontSize      int
	SmallFontSize int
}

const (
	DefaultWindowWidth  = int32(1024)
	DefaultWindowHeight = int32(768)
	FontSize            = 50
	SmallFontSize       = 40
)

func InitWindow(title string) *Window {
	displayIndex := 0
	displayMode, err := sdl.GetCurrentDisplayMode(displayIndex)

	width := DefaultWindowWidth
	height := DefaultWindowHeight

	if err == nil {
		width = displayMode.W
		height = displayMode.H
	} else {
		fmt.Fprintf(os.Stderr, "Failed to get display mode: %s\n", err)
	}

	Logger.Info("Window size", "width", width, "height", height)
	return InitWindowWithSize(title, 1024, 768, FontSize, SmallFontSize)
}

func InitWindowWithSize(title string, width, height int32, fontSize int, smallFontSize int) *Window {
	window, err := sdl.CreateWindow(title, 0, 0, width, height, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		os.Exit(1)
	}

	InitFonts(fontSize+10, fontSize, smallFontSize)

	return &Window{
		Window:        window,
		Renderer:      renderer,
		Title:         title,
		Width:         width,
		Height:        height,
		FontSize:      fontSize,
		SmallFontSize: smallFontSize,
	}
}

func (window Window) CloseWindow() {
	CloseFonts()
	window.Renderer.Destroy()
	window.Window.Destroy()
}
