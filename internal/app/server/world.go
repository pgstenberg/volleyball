package server

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
)

type GameWorld struct {
	onUpdate             func(*GameWorld, float64)
	ticker               *time.Ticker
	state                chan GameWorldState
	NetworkInputChannel  chan []byte
	NetworkOutputChannel chan []byte
	mux                  sync.Mutex
	inputBuffer          map[uuid.UUID][]playerInput
	snapshot             snapshot
}

type GameWorldState int

const (
	Stopped GameWorldState = iota
	Started
)

func worldUpdate(world *GameWorld, delta float64) {
	world.mux.Lock()

	snapshot0 := copySnapshot(&world.snapshot)

	for id, playerInput := range world.inputBuffer {
		for len(playerInput) > 0 {
			input := world.inputBuffer[id][0]

			world.snapshot.players[id].proccessInput(input.value, delta)

			if world.snapshot.lastSequenceNumber[id] < input.sequenceNumber {
				world.snapshot.lastSequenceNumber[id] = input.sequenceNumber
			}
			world.inputBuffer[id] = world.inputBuffer[id][1:]
		}
	}

	dSnapshot := diffSnapshot(snapshot0, &world.snapshot)

	if len(dSnapshot.players) > 0 {
		b, err := json.Marshal(dSnapshot)
		if err != nil {
			fmt.Println(err)
			return
		}
		world.NetworkOutputChannel <- b
	}

	world.mux.Unlock()
}

func NewGameWorld(tickRate time.Duration) *GameWorld {
	world := GameWorld{
		onUpdate:             worldUpdate,
		ticker:               time.NewTicker(time.Second / tickRate),
		state:                make(chan GameWorldState),
		NetworkInputChannel:  make(chan []byte),
		NetworkOutputChannel: make(chan []byte),
		snapshot: snapshot{
			players:            make(map[uuid.UUID]player),
			lastSequenceNumber: make(map[uuid.UUID]uint32),
		},
		inputBuffer: make(map[uuid.UUID][]playerInput),
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

		fmt.Printf("Network Input: %s\n", string(data))

		id, err := uuid.FromBytes(data[0:16])
		if err != nil {
			log.Fatal(err)
			continue
		}

		payload := data[16:]

		s := strings.Split(string(payload), ";")

		sequenceNumber, err := strconv.ParseUint(s[0], 10, 32)

		if err != nil {
			log.Fatal(err)
			continue
		}

		value, err := strconv.ParseUint(s[1], 10, 8)

		if err != nil {
			log.Fatal(err)
			continue
		}

		fmt.Printf("ID: %s, seqNum: %d, val: %d\n", id.String(), sequenceNumber, value)

		pInput := playerInput{
			sequenceNumber: uint32(sequenceNumber),
			value:          uint8(value),
		}

		world.mux.Lock()
		world.inputBuffer[id] = append(world.inputBuffer[id], pInput)
		world.mux.Unlock()
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
