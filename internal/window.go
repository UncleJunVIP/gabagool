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
)

func InitWindow(title string) *Window {
	displayIndex := 0
	displayMode, err := sdl.GetCurrentDisplayMode(displayIndex)

	width := DefaultWindowWidth
	height := DefaultWindowHeight

	if os.Getenv("DEVELOPMENT") == "true" {
		width = DefaultWindowWidth
		height = DefaultWindowHeight
	} else if err == nil {
		width = displayMode.W
		height = displayMode.H
	} else {
		fmt.Fprintf(os.Stderr, "Failed to get display mode: %s\n", err)
	}

	return InitWindowWithSize(title, width, height)
}

func InitWindowWithSize(title string, width, height int32) *Window {
	window, err := sdl.CreateWindow(title, 0, 0, width, height, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
		os.Exit(1)
	}

	InitFonts(getFontScale(width, height))

	return &Window{
		Window:   window,
		Renderer: renderer,
		Title:    title,
		Width:    width,
		Height:   height,
	}
}

func getFontScale(width, height int32) int {
	if width == DefaultWindowWidth && height == DefaultWindowHeight {
		return 3
	}

	return 2
}

func (window Window) CloseWindow() {
	CloseFonts()
	window.Renderer.Destroy()
	window.Window.Destroy()
}
