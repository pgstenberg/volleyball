package server

import "math"

type player struct {
	PosX, PosY         int
	velX, velY         float64
	LastSequenceNumber uint32
	state              []bool
}

type playerInput struct {
	sequenceNumber uint32
	value          uint8
}

const stateMovingLeft int = 0
const stateMovingRight int = 1
const stateJumping int = 2

const playerSpeed float64 = 4 * float64(ServerTickRate)
const gravity float64 = 8 * float64(ServerTickRate)
const jumpSpeed = gravity * 3
const maxJumpHight = 150

func (p *player) proccessInput(value uint8) {

	switch value {
	case 1:
		p.state[stateMovingLeft] = true
	case 2:
		p.state[stateMovingRight] = true
	case 3:
		if p.PosY == 0 && !p.state[stateJumping] {
			p.state[stateJumping] = true
		}
	}

}

func (p *player) update(delta float64) {

	if p.state[stateMovingLeft] {
		p.velX = -playerSpeed
	} else if p.state[stateMovingRight] {
		p.velX = playerSpeed
	}

	if p.state[stateJumping] {
		p.velY = jumpSpeed
	}

	if p.PosY > 0 {
		p.velY -= gravity
	}

	p.PosX += int(math.Round(p.velX * delta))
	p.PosY += int(math.Round(p.velY * delta))

	if p.PosY < 0 {
		p.PosY = 0
		p.velY = 0
		p.state[stateJumping] = false
	} else if p.PosY > maxJumpHight {
		p.PosY = maxJumpHight
		p.state[stateJumping] = false
	}

	p.state[stateMovingRight] = false
	p.state[stateMovingLeft] = false
	p.velX = 0
}

func (p *player) copy() *player {
	return &player{
		PosX:               p.PosX,
		PosY:               p.PosY,
		velX:               p.velX,
		velY:               p.velY,
		LastSequenceNumber: p.LastSequenceNumber,
		state:              p.state,
	}
}
