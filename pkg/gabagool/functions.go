package gabagool

func getFontScale(width, height int32) int {

	if width == DefaultWindowWidth && height == DefaultWindowHeight {
		return 4
	}

	return 3
}
