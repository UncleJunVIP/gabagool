package gabagool

func getFontScale(width, height int32) int {

	if width == DefaultWindowWidth && height == DefaultWindowHeight {
		return 3
	}

	return 2
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func min32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
func max32(a, b int32) int32 {
	if a > b {
		return a
	}
	return b
}
func abs32(x int32) int32 {
	if x < 0 {
		return -x
	}
	return x
}
