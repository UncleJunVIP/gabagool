package internal

func getFontScale(width, height int32) int {
	if width == DefaultWindowWidth && height == DefaultWindowHeight {
		return 3
	}

	return 2
}
