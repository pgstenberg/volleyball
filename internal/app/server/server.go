package server

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pgstenberg/volleyball/internal/pkg/networking"
	uuid "github.com/satori/go.uuid"
)

type GameServer struct {
	Bind string

	tick     int
	tickRate time.Duration
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
		ID:   uuid.Must(uuid.NewV4())}

	world.snapshot.Players[client.ID] = &player{
		PosX:               0,
		PosY:               0,
		velX:               0,
		velY:               0,
		LastSequenceNumber: 0,
		state:              make([]bool, 3),
		jumpInputs:         0,
	}

	client.Hub.Register <- client

	go client.Read(world.NetworkInputChannel)
	go client.Write()

	client.Send <- []byte(client.ID.String())

}
