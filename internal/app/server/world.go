package server

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

type GameWorld struct {
	onUpdate func(*GameWorld, float64)
	ticker *time.Ticker
	state chan GameWorldState
	NetworkInputChannel chan []byte
	NetworkOutputChannel chan []byte
	mux sync.Mutex
	inputBuffer []PlayerInput
	Snapshot Snapshot
}

type GameWorldState int

const (
	Stopped GameWorldState = iota
	Started
)

type Snapshot struct {
	Players[] Player
	LastSequenceNumber[] uint32
}

func copySnapshot(snapshot *Snapshot) *Snapshot{
	newSnapshot := Snapshot{
		Players: make([]Player, len(snapshot.Players)),
	}
	copy(newSnapshot.Players, snapshot.Players)
	return &newSnapshot
}

func diffSnapshot(snapshot0 *Snapshot, snapshot *Snapshot) bool {
	for index, player := range snapshot.Players {
		if player.X != snapshot0.Players[index].X || player.Y != snapshot0.Players[index].Y{
			return true
		}
	}
	return false
}

func handlePlayerInputs(world *GameWorld, delta float64){
	world.mux.Lock()

	snapshot0 := copySnapshot(&world.Snapshot)

	for len(world.inputBuffer) > 0 {
		input := world.inputBuffer[0]
		dx := int(math.Round(float64(40) * delta))
		switch input.Value {
			case 1:
				world.Snapshot.Players[input.Id].X -= dx
			case 2:
				world.Snapshot.Players[input.Id].X += dx
		}
		if world.Snapshot.LastSequenceNumber[input.Id] < input.SequenceNumber {
			world.Snapshot.LastSequenceNumber[input.Id] = input.SequenceNumber
		}
		world.inputBuffer = world.inputBuffer[1:]
	}

	if diffSnapshot(snapshot0, &world.Snapshot) {
		b, err := json.Marshal(world.Snapshot)
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
		onUpdate: handlePlayerInputs,
		ticker: time.NewTicker(time.Second / tickRate),
		state: make(chan GameWorldState),
		NetworkInputChannel: make(chan []byte),
		NetworkOutputChannel: make(chan []byte),
		Snapshot: Snapshot{
			Players: []Player{
				{Id: 0, X: 0, Y:0 },
				{Id: 1, X: 0, Y:0 },
			},
			LastSequenceNumber: []uint32{
				0,
				0,
			},
		},
	}
	return &world
}

func (world *GameWorld) Start(){
	world.state <- Started
}
func (world *GameWorld) Stop(){
	world.state <- Stopped
}


func (world *GameWorld) startNetworkingInoutLoop(){
	for data := range world.NetworkInputChannel {

		s := strings.Split(string(data), ";")

		sequenceNumber, err := strconv.Atoi(s[0])

		if err != nil {
			continue
		}

		id, err := strconv.Atoi(string(s[1][0:1]))

		if err != nil {
			continue
		}

		value, err := strconv.Atoi(string(s[1][1:2]))

		if err != nil {
			continue
		}

		playerInput := PlayerInput{
			SequenceNumber: uint32(sequenceNumber),
			Id: uint8(id),
			Value: uint8(value),
		}

		world.mux.Lock()
		world.inputBuffer = append(world.inputBuffer, playerInput)
		world.mux.Unlock()
	}
}

func (world *GameWorld) startStateLoop(){
	for {
		select {
			case state := <- world.state:
				switch state {
					case Stopped:
						world.ticker.Stop()
						close(world.NetworkInputChannel)
						close(world.NetworkOutputChannel)
					case Started:
						go world.startGameLoop()
						go world.startNetworkingInoutLoop()
				}
		}
	}
}

func (world *GameWorld) startGameLoop(){

	t0 := time.Now().UnixNano()

	for {
		select {
			case <- world.ticker.C:
				t := time.Now().UnixNano()
				// DT in seconds
				delta := float64(t-t0) / 1000000000
				t0 = t

				world.onUpdate(world, delta)
		}
	}

}