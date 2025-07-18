package gabagool

type MenuItem struct {
	Text               string
	Selected           bool
	Focused            bool
	NotMultiSelectable bool
	Metadata           interface{}
	ImageFilename      string
}

type ListReturn struct {
	Items           []MenuItem
	SelectedIndex   int
	SelectedItem    *MenuItem
	SelectedIndices []int
	SelectedItems   []*MenuItem
	VisiblePosition int
	LastPressedBtn  Button
	ActionTriggered bool
}

func (r *ListReturn) populateSingleSelection(index int, items []MenuItem, visibleStartIndex int) {
	r.SelectedIndex = index
	r.SelectedItem = &items[index]
	r.SelectedIndices = []int{index}
	r.SelectedItems = []*MenuItem{&items[index]}
	r.VisiblePosition = index - visibleStartIndex
}

func (r *ListReturn) populateMultiSelection(indices []int, items []MenuItem, visibleStartIndex int) {
	r.SelectedIndex = indices[0]
	r.SelectedItem = &items[indices[0]]
	r.SelectedIndices = indices
	r.SelectedItems = make([]*MenuItem, len(indices))
	r.VisiblePosition = indices[0] - visibleStartIndex
	for i, idx := range indices {
		r.SelectedItems[i] = &items[idx]
	}
}
