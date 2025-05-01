package ui

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

func InitWindow(title string, width, height int32, fontSize int, smallFontSize int) *Window {
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
	CloseGameController()
	window.Renderer.Destroy()
	window.Window.Destroy()
}
