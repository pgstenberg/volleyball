package server

import "math"

type entity struct {
	positionX, positionY uint16
	velocityX, velocityY float64
}

func (e *entity) update(delta float64) {
	e.positionX += uint16(math.Round(e.velocityX * delta))
	e.positionY += uint16(math.Round(e.velocityY * delta))
}
