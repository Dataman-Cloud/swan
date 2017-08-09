package utils

import (
	"math"
)

func ComputeWeight(n, t, c float64) float64 {
	x := (t - n) * 100 * c
	y := n * (c - 1)

	if y == 0 || x == 0 {
		return 100
	}

	return math.Floor(-x/y + 0.5)
}
