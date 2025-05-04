package models

import "github.com/veandco/go-sdl2/sdl"

type MenuItem struct {
	Text     string
	Selected bool
	Focused  bool
	Metadata interface{}
}

type ListReturn struct {
	SelectedIndex  int
	SelectedItem   *MenuItem
	LastPressedKey sdl.Keycode
	LastPressedBtn uint8
	Cancelled      bool
}
