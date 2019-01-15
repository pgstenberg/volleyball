package server

import (
	"fmt"
	"math"
	"time"
)

type player struct {
	positionX      uint16
	positionY      uint16
	velocityX      float64
	velocityY      float64
	sequenceNumber uint32
}

/*
*	CONSTANTS
 */
const stateMovingLeft int = 0
const stateMovingRight int = 1
const stateJumping int = 2

const playerSpeed float64 = 4 * float64(ServerTickRate)
const gravity float64 = 4 * float64(ServerTickRate)
const jumpSpeed = gravity * 3
const maxJumpHight = 150

func (p *player) getConsecutiveInputs(numSequences uint32, world *GameWorld, clientID uint8) [][]bool {
	input := make([][]bool, 0)
	foundInLastTick := true

	t := world.tick
	s := p.sequenceNumber

	for s >= (p.sequenceNumber-(numSequences-1)) && s >= 0 && t > 0 {
		bInput := world.stateBuffer[uint8(t%stateBufferSize)][clientID][s]

		if len(bInput) == 0 {
			// If we did not find any in the last server tick, return
			if !foundInLastTick {
				return input
			}
			// Continue to next server tick
			foundInLastTick = false
			t--
			continue
		}

		input = append(input, bInput)
		// Get to next sequence number
		s--
		foundInLastTick = true
	}

	return input
}

func (p *player) process(world *GameWorld, clientID uint8, input []bool) {

	//fmt.Printf("Input: %s\n", input)

	if input[stateMovingLeft] {
		p.velocityX -= playerSpeed
	} else if input[stateMovingRight] {
		p.velocityX += playerSpeed
	}

	start := time.Now()
	l0 := p.getConsecutiveInputs(12, world, clientID)

	fmt.Printf("(%d) LAST INPUTS [%d]: %s\n", time.Since(start), len(l0), l0)

	//JUMPING
}

func (p *player) update(delta float64) {

	//fmt.Printf("X: %d\n", p.positionX)

	p.positionX = p.positionX + uint16(math.Round(p.velocityX*delta))
	p.positionY = p.positionY + uint16(math.Round(p.velocityY*delta))

	p.velocityX = 0
}

func (p *player) copy() *player {
	clone := *p
	return &clone
}
