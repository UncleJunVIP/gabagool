package models

type MenuItem struct {
	Text     string
	Selected bool
	Focused  bool
	Metadata interface{}
}
