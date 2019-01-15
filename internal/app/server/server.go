package server

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pgstenberg/volleyball/internal/pkg/networking"
)

type GameServer struct {
	Bind string
}

const ServerTickRate int = 20

func (s *GameServer) Start() {

	upgrader := websocket.Upgrader{}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	addr := flag.String("addr", s.Bind, "http service address")

	gw := NewGameWorld(time.Duration(ServerTickRate))
	hub := networking.NewHub(gw.NetworkOutputChannel)

	go hub.Start()
	go gw.startStateLoop()

	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ws(gw, hub, &upgrader, w, r)
	})

	log.Printf("Starting server using %s.", s.Bind)

	gw.start()

	err := http.ListenAndServe(*addr, nil)

	if err != nil {
		log.Fatal(err)
	}

}

func ws(world *GameWorld, hub *networking.Hub, upgrader *websocket.Upgrader, w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// Create new client
	client := &networking.Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan []byte, 256),
		ID:   hub.CalcClientID()}

	fmt.Printf("ClientID: %d", client.ID)

	if nil == world.players[uint8(world.tick%stateBufferSize)] {
		world.players[uint8(world.tick%stateBufferSize)] = make(map[uint8]*player)
	}

	world.players[uint8(world.tick%stateBufferSize)][client.ID] = &player{
		sequenceNumber: 0,
	}

	world.players[uint8(world.tick%stateBufferSize)][client.ID].positionX = uint16(0)
	world.players[uint8(world.tick%stateBufferSize)][client.ID].positionY = uint16(0)
	world.players[uint8(world.tick%stateBufferSize)][client.ID].velocityX = float64(0)
	world.players[uint8(world.tick%stateBufferSize)][client.ID].velocityY = float64(0)

	client.Hub.Register <- client

	go client.Read(world.NetworkInputChannel)
	go client.Write()

	//client.Send <- []byte(client.ID.String())

}
