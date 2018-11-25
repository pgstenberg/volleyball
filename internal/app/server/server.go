package server

import (
	"flag"
	"github.com/gorilla/websocket"
	"github.com/pgstenberg/volleyball/internal/pkg/networking"
	"log"
	"net/http"
)

type GameServer struct {
	Bind string
}

func (s *GameServer) Start(){

	upgrader := websocket.Upgrader{}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	addr := flag.String("addr", s.Bind, "http service address")

	hub := networking.NewHub()

	go hub.Start()



	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ws(hub, &upgrader, w, r)
	})

	log.Fatal(http.ListenAndServe(*addr, nil))

}

func ws(hub *networking.Hub, upgrader *websocket.Upgrader, w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &networking.Client{Hub: hub, Conn: conn, Send: make(chan []byte, 256)}
	client.Hub.Register <- client

	go client.Read()
	go client.Write()
}