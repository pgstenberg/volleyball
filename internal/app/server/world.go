package server

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"
)

type GameWorld struct {
	Tick int
	onUpdate func(*GameWorld, float64)
	ticker *time.Ticker
	state chan GameWorldState
	NetworkInputChannel chan []byte
	NetworkOutputChannel chan []byte
	mux sync.Mutex
	inputBuffer []PlayerInput
	player Player
}

type GameWorldState int

const (
	Stopped GameWorldState = iota
	Started
)

func handlePlayerInputs(world *GameWorld, delta float64){
	world.mux.Lock()

	xBefore := world.player.X

	for len(world.inputBuffer) > 0 {
		input := world.inputBuffer[0]
		dx := int(math.Round(float64(4) * delta))
		switch input.Value {
			case 1:
				world.player.X -= dx
		case 2:
				world.player.X += dx
		}
		world.inputBuffer = world.inputBuffer[1:]
	}

	if xBefore != world.player.X {
		snapshot := fmt.Sprintf("%d", world.player.X)
		world.NetworkOutputChannel <- []byte(snapshot)
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
		player: Player{
			Id:1,
			X:0, Y:0},
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

		i, err := strconv.Atoi(string(data[0:1]))

		if err != nil {
			fmt.Printf("Error %v\n", err)
			continue
		}

		world.mux.Lock()
		world.inputBuffer = append(world.inputBuffer, PlayerInput{
			Id: 1,
			Value: i,
		})
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
				world.Tick++

		}
	}

}