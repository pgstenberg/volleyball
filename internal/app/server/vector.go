package server

import "math"

type vector struct {
	x int
	y int
}

func (v *vector) set(x int, y int) {
	v.x = x
	v.y = y
}

func (v *vector) mul(f float64) {
	v.x = int(math.Round(float64(v.x) * f))
	v.y = int(math.Round(float64(v.y) * f))
}

func (v *vector) add(v2 *vector) {
	v.x += v2.x
	v.y += v2.y
}
