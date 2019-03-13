package server

import (
	"math"
)

type player struct {
	positionX                   int
	positionY                   int
	velocityX                   float64
	velocityY                   float64
	lastReceivedSequenceNumber  uint32
	lastProcessedSequenceNumber uint32
	numJumpInputs               uint8
	onGround                    bool
	jumping                     bool
}

/*
*	CONSTANTS
 */
const stateMovingLeft int = 0
const stateMovingRight int = 1
const stateJumping int = 2

const playerSpeed float64 = 4 * float64(60)
const gravity float64 = 4 * float64(60)
const jumpSpeed float64 = 4 * 3 * float64(60)
const maxJumpHight = 150

func emptyInputs() []bool {
	return []bool{false, false, false}
}

func (p *player) process(world *GameWorld, clientID uint8, input []bool, sequenceNumber uint32, delta float64) {

	if p.onGround && !input[stateJumping] {
		p.numJumpInputs = 0
		p.jumping = false
	}

	if input[stateMovingLeft] {
		p.velocityX -= playerSpeed
	} else if input[stateMovingRight] {
		p.velocityX += playerSpeed
	}

	if input[stateJumping] && (!p.jumping || p.numJumpInputs < 3) {
		p.velocityY += jumpSpeed
		p.numJumpInputs++
		p.onGround = false
		p.jumping = true
	}

	if !p.onGround {
		p.velocityY -= gravity
	}

	dx := int(math.Round(p.velocityX * delta))
	dy := int(math.Round(p.velocityY * delta))

	p.positionX = p.positionX + dx
	p.velocityX = 0

	if p.positionY+dy < 0 {
		p.positionY = 0
		p.velocityY = 0
		p.onGround = true
	} else {
		p.positionY = p.positionY + dy
	}

}

func (p *player) copy() *player {
	clone := *p
	return &clone
}
