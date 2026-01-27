package common

func ClampInt(value int, min int, max int) int {
	if value < min {
		return min
	}
	if max > 0 && value > max {
		return max
	}
	return value
}
