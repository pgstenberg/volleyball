package server

import (
	"fmt"
	"math"
	"strconv"
)

type player struct {
	positionX                   int
	positionY                   int
	velocityX                   float64
	velocityY                   float64
	lastReceivedSequenceNumber  uint32
	lastProcessedSequenceNumber uint32
}

/*
*	CONSTANTS
 */
const stateMovingLeft int = 0
const stateMovingRight int = 1
const stateJumping int = 2

const playerSpeed float64 = 4 * float64(60)
const gravity float64 = 4 * float64(20)
const jumpSpeed float64 = 4 * 3 * float64(20)
const maxJumpHight = 150

func emptyInputs() []bool {
	return []bool{false, false, false}
}

func (p *player) getLastConsecutiveInputs(numSequences uint32, world *GameWorld, clientID uint8) [][]bool {
	return p.getConsecutiveInputs(p.lastProcessedSequenceNumber, numSequences, world, clientID)
}

func (p *player) getConsecutiveInputs(startSequenceNumber uint32, numSequences uint32, world *GameWorld, clientID uint8) [][]bool {
	input := make([][]bool, 0)
	foundInLastTick := true

	t := world.tick
	s := startSequenceNumber

	for s >= (startSequenceNumber-numSequences) && s >= 0 && t > 0 {
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

func (p *player) process(world *GameWorld, clientID uint8, input []bool, sequenceNumber uint32, delta float64) {

	//fmt.Printf("Input: %s\n", input)

	if input[stateMovingLeft] {
		p.velocityX -= playerSpeed
	} else if input[stateMovingRight] {
		p.velocityX += playerSpeed
	}

	if input[stateJumping] {

		validJump := true
		linputs := p.getConsecutiveInputs(sequenceNumber, 5, world, clientID)

		if len(linputs) == 5 {
			for _, i := range linputs {
				validJump = !i[stateJumping]
			}
		}

		fmt.Printf("> Seq: %d Inputs [%d] <<%s>>: %s\n", sequenceNumber, len(linputs), strconv.FormatBool(validJump), linputs)

		if validJump {
			p.velocityY += jumpSpeed
		}

		if p.positionY == 0 {
			p.velocityY += jumpSpeed
		}
	}

	if p.velocityY > 0 || p.positionY > 0 {
		p.velocityY -= gravity
	}

	dx := int(math.Round(p.velocityX * delta))
	dy := int(math.Round(p.velocityY * delta))

	p.positionX = p.positionX + dx
	p.velocityX = 0

	if p.positionY+dy < 0 {
		p.positionY = 0
		p.velocityY = 0
	} else {
		p.positionY = p.positionY + dy
	}

	fmt.Printf("Seq: %d, X: %d, Y: %d, VelX: %d, VelY: %d\n", sequenceNumber, p.positionX, p.positionY, p.velocityX, p.velocityY)

}

/*
func (p *player) update(delta float64) {

	if p.velocityY > 0 || p.positionY > 0 {
		p.velocityY -= gravity
	}

	dx := int(math.Round(p.velocityX * delta))
	dy := int(math.Round(p.velocityY * delta))

	p.positionX = p.positionX + dx
	p.velocityX = 0

	if p.positionY+dy < 0 {
		p.positionY = 0
		p.velocityY = 0
	} else {
		p.positionY = p.positionY + dy
	}

	fmt.Printf("Seq: %d, X: %d, Y: %d, VelX: %d, VelY: %d\n", p.sequenceNumber, p.positionX, p.positionY, p.velocityX, p.velocityY)

}
*/

func (p *player) copy() *player {
	clone := *p
	return &clone
}
