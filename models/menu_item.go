package models

type MenuItem struct {
	Text     string
	Selected bool
	Focused  bool
	Metadata interface{}
}

type ListReturn struct {
	SelectedIndex   int
	SelectedItem    *MenuItem
	SelectedIndices []int
	SelectedItems   []*MenuItem
	LastPressedBtn  uint8
	ActionTriggered bool
	Cancelled       bool
}

func (r *ListReturn) PopulateSingleSelection(index int, items []MenuItem) {
	r.SelectedIndex = index
	r.SelectedItem = &items[index]
	r.SelectedIndices = []int{index}
	r.SelectedItems = []*MenuItem{&items[index]}
}

// PopulateMultiSelection populates the result with multiple selections
func (r *ListReturn) PopulateMultiSelection(indices []int, items []MenuItem) {
	r.SelectedIndex = indices[0]
	r.SelectedItem = &items[indices[0]]
	r.SelectedIndices = indices
	r.SelectedItems = make([]*MenuItem, len(indices))
	for i, idx := range indices {
		r.SelectedItems[i] = &items[idx]
	}
}
