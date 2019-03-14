package server

import (
	"encoding/binary"
	"sort"
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
	// State based on server tick. [TICK][PLAYERID]
	players map[uint8]map[uint8]*player
	// Server tick and input buffer. [TICK][PLAYERID][SEQID][INPUTS]
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
	prevTickIdx := uint8((world.tick - 1) % stateBufferSize)

	returnData := []uint8{}

	world.players[nextTickIdx] = make(map[uint8]*player)
	world.stateBuffer[nextTickIdx] = make(map[uint8]map[uint32][]bool)

	for id, p := range world.players[currTickIdx] {

		d := delta / 3

		if nil == world.stateBuffer[currTickIdx][id] {
			world.stateBuffer[currTickIdx][id] = make(map[uint32][]bool)
		}

		// Fill in empty sequenceNumbers from client
		for seq := p.lastProcessedSequenceNumber + 1; seq <= p.lastProcessedSequenceNumber+3; seq++ {
			if nil == world.stateBuffer[currTickIdx][id][seq] && len(world.stateBuffer[currTickIdx][id]) < 3 {
				world.stateBuffer[currTickIdx][id][seq] = emptyInputs()
			}
		}

		// Sort inputs and process them
		var keys []uint32
		for k := range world.stateBuffer[currTickIdx][id] {
			keys = append(keys, k)
		}

		sort.Slice(keys, func(i, j int) bool {
			return keys[i] < keys[j]
		})

		for _, k := range keys {
			p.process(world, id, world.stateBuffer[currTickIdx][id][k], k, d)
			if k > p.lastProcessedSequenceNumber {
				p.lastProcessedSequenceNumber = uint32(k)
			}
		}

		// Check if player state have changed, if that is the case send update to clients.
		if nil != world.players[prevTickIdx][id] {
			if world.players[prevTickIdx][id].positionX != p.positionX || world.players[prevTickIdx][id].positionY != p.positionY {
				returnData = append(returnData, id)
			}
		}

		world.players[nextTickIdx][id] = p.copy()
		world.stateBuffer[nextTickIdx][id] = make(map[uint32][]bool)

	}

	// Send updated state to clients.
	breturn := []byte{}
	for _, id := range returnData {

		breturn = append(breturn, id)

		a := make([]byte, 4)
		binary.LittleEndian.PutUint32(a, world.players[nextTickIdx][id].lastReceivedSequenceNumber)
		breturn = append(breturn, a...)

		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, uint32(world.players[nextTickIdx][id].positionX))
		breturn = append(breturn, b...)

		c := make([]byte, 4)
		binary.LittleEndian.PutUint32(c, uint32(world.players[nextTickIdx][id].positionY))
		breturn = append(breturn, c...)

	}
	if len(breturn) > 0 {
		world.NetworkOutputChannel <- breturn
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

		switch packageType {
		// Client Inputs
		case 1:

			seq := binary.LittleEndian.Uint32(data[5:9])

			inputs := []bool{false, false, false}

			for idx := 9; idx < len(data); idx++ {
				id := uint8(data[idx])
				inputs[id] = true
			}

			world.stateBuffer[currTickIdx][clientID][seq] = inputs

			// update sequence ID
			world.players[currTickIdx][clientID].lastReceivedSequenceNumber = seq

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
