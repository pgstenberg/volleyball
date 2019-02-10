package server

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

type GameWorld struct {
	onUpdate             func(*GameWorld, float64)
	ticker               *time.Ticker
	tick                 uint16
	state                chan GameWorldState
	NetworkInputChannel  chan []byte
	NetworkOutputChannel chan []byte
	mux                  sync.Mutex
	// State based on server tick.
	players map[uint8]map[uint8]*player
	// Server tick and input buffer.
	stateBuffer map[uint8]map[uint8]map[uint32][]bool
}

type GameWorldState int

const (
	Stopped GameWorldState = iota
	Started
)

const stateBufferSize = 100

func worldUpdate(world *GameWorld, delta float64) {
	world.mux.Lock()

	currTickIdx := uint8(world.tick % stateBufferSize)
	nextTickIdx := uint8((world.tick + 1) % stateBufferSize)

	world.players[nextTickIdx] = make(map[uint8]*player)
	world.stateBuffer[nextTickIdx] = make(map[uint8]map[uint32][]bool)

	for id, p := range world.players[currTickIdx] {
		fmt.Printf("UPDATE >>>>>> %d\n", world.tick)
		numInputs := len(world.stateBuffer[currTickIdx][id])
		for tid, i := range world.stateBuffer[currTickIdx][id] {
			p.process(world, id, i, tid, delta, numInputs)
		}
		fmt.Printf("<<<<<<<<<<<<<<<<<<<< \n")
		world.players[nextTickIdx][id] = p.copy()
		world.stateBuffer[nextTickIdx][id] = make(map[uint32][]bool)
	}
	world.tick++

	world.mux.Unlock()
}

func NewGameWorld(tickRate time.Duration) *GameWorld {
	world := GameWorld{
		onUpdate:             worldUpdate,
		ticker:               time.NewTicker(time.Second / tickRate),
		tick:                 1,
		state:                make(chan GameWorldState),
		NetworkInputChannel:  make(chan []byte),
		NetworkOutputChannel: make(chan []byte),
		players:              make(map[uint8]map[uint8]*player),
		stateBuffer:          make(map[uint8]map[uint8]map[uint32][]bool),
	}
	return &world
}

func (world *GameWorld) start() {
	world.state <- Started
}
func (world *GameWorld) stop() {
	world.state <- Stopped
}

func (world *GameWorld) startNetworkLoop() {
	for data := range world.NetworkInputChannel {

		world.mux.Lock()

		clientID := uint8(data[0])
		packageType := uint8(data[1])

		currTickIdx := uint8(world.tick % stateBufferSize)
		prevTickIdx := uint8((world.tick - 1) % stateBufferSize)

		switch packageType {
		case 1:

			seq := binary.LittleEndian.Uint32(data[5:9])

			inputs := []bool{false, false, false}

			for idx := 9; idx < len(data); idx++ {
				id := uint8(data[idx])
				inputs[id] = true
			}

			//
			// Input Buffer
			//
			/*
				if nil == world.stateBuffer[currTickIdx] {
					world.stateBuffer[currTickIdx] = make(map[uint8]map[uint32][]bool)
				}
				if nil == world.stateBuffer[currTickIdx][clientID] {
					world.stateBuffer[currTickIdx][clientID] = make(map[uint32][]bool)
				}
			*/
			// Update next server tick with inputs and last sequence number.

			world.stateBuffer[currTickIdx][clientID][seq] = inputs

			//
			// Player State
			//
			if nil == world.players[currTickIdx] {
				world.players[currTickIdx] = make(map[uint8]*player)
			}
			if nil == world.players[currTickIdx][clientID] {
				world.players[currTickIdx][clientID] = world.players[prevTickIdx][clientID].copy()
			}
			// update sequence ID
			world.players[currTickIdx][clientID].sequenceNumber = seq

			world.mux.Unlock()

		}

	}
}

func (world *GameWorld) startStateLoop() {
	for {
		select {
		case state := <-world.state:
			switch state {
			case Stopped:
				world.ticker.Stop()
				close(world.NetworkInputChannel)
				close(world.NetworkOutputChannel)
			case Started:
				go world.startGameLoop()
				go world.startNetworkLoop()
			}
		}
	}
}

func (world *GameWorld) startGameLoop() {

	t0 := time.Now().UnixNano()

	for {
		select {
		case <-world.ticker.C:
			t := time.Now().UnixNano()
			// DT in seconds
			delta := float64(t-t0) / 1000000000
			t0 = t

			world.onUpdate(world, delta)
		}
	}

}
