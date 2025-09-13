package gabagool

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
