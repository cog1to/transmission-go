package utils

func MinInt(x, y int) int {
	if x < y {
		return x
	} else {
		return y
	}
}

func MaxInt(x, y int) int {
	if x > y {
		return x
	} else {
		return y
	}
}

func MaxFloat32(x, y float32) float32 {
	if x > y {
		return x
	} else {
		return y
	}
}
