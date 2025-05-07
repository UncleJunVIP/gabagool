package models

type MenuItem struct {
	Text     string
	Selected bool
	Focused  bool
	Metadata interface{}
}

type ListReturn struct {
	SelectedIndex  int
	SelectedItem   *MenuItem
	LastPressedBtn uint8
	Cancelled      bool
}
