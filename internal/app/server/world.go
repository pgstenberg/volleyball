package server

import (
	"encoding/json"
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

	for id := range world.inputBuffer {
		for len(world.inputBuffer[id]) > 0 {
			input := world.inputBuffer[id][0]
			world.snapshot.Players[id].proccessInput(input.value)
			if world.snapshot.Players[id].LastSequenceNumber < input.sequenceNumber {
				world.snapshot.Players[id].LastSequenceNumber = input.sequenceNumber
			}
			world.inputBuffer[id] = world.inputBuffer[id][1:]
		}
	}

	for _, p := range world.snapshot.Players {
		p.update(delta)
	}

	dSnapshot := diffSnapshot(snapshot0, &world.snapshot)

	if len(dSnapshot.Players) > 0 {
		b, err := json.Marshal(&dSnapshot)
		if err != nil {
			log.Fatalf("Unable to marshal snapshot, err: %s", err)
		} else {
			world.NetworkOutputChannel <- b
		}
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
			Players: make(map[uuid.UUID]*player),
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

		world.mux.Lock()
		for _, input := range s[1:] {
			value, err := strconv.ParseUint(input, 10, 8)

			if err != nil {
				log.Fatal(err)
				continue
			}

			world.inputBuffer[id] = append(world.inputBuffer[id], playerInput{
				sequenceNumber: uint32(sequenceNumber),
				value:          uint8(value),
			})
		}
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
