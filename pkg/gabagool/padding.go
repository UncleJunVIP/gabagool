package gabagool

type padding struct {
	Top    int32
	Right  int32
	Bottom int32
	Left   int32
}

func uniformPadding(value int32) padding {
	return padding{
		Top:    value,
		Right:  value,
		Bottom: value,
		Left:   value,
	}
}
