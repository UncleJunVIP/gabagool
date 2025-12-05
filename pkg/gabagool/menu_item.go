package gabagool

type MenuItem struct {
	Text               string
	Selected           bool
	Focused            bool
	NotMultiSelectable bool
	Metadata           interface{}
	ImageFilename      string
	BackgroundFilename string
}

// ListResult is the standardized return type for the List component
type ListResult struct {
	Items           []MenuItem
	Selected        []int      // Indices of selected items (always a slice, even for single selection)
	Action          ListAction // The action taken when exiting (Selected or Triggered)
	VisiblePosition int        // Position of first selected item relative to VisibleStartIndex (for scroll restoration)
}
