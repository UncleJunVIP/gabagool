package types

import "github.com/veandco/go-sdl2/sdl"

type Scene interface {
	Init() error
	Activate() error
	Deactivate() error
	HandleEvent(sdl.Event) bool
	Update() error
	Render() error
	Destroy() error
}
