package models

import "github.com/veandco/go-sdl2/sdl"

type MenuItem struct {
	Text     string
	Selected bool
	Focused  bool
	Metadata interface{}
}

type ListReturn struct {
	SelectedIndex  int         // Index of the selected item
	SelectedItem   *MenuItem   // Pointer to the selected item
	LastPressedKey sdl.Keycode // Last pressed keyboard key
	LastPressedBtn uint8       // Last pressed controller button
	Cancelled      bool        // Whether the selection was cancelled
}
