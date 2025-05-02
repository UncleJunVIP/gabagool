package ui

type Padding struct {
	Top    int32
	Right  int32
	Bottom int32
	Left   int32
}

func UniformPadding(value int32) Padding {
	return Padding{
		Top:    value,
		Right:  value,
		Bottom: value,
		Left:   value,
	}
}

func HVPadding(horizontal, vertical int32) Padding {
	return Padding{
		Top:    vertical,
		Right:  horizontal,
		Bottom: vertical,
		Left:   horizontal,
	}
}
