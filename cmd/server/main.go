package main

import (
	"github.com/pgstenberg/volleyball/internal/app/server"
)

func main(){
	s := server.GameServer{Bind:"localhost:8080"}
	s.Start()
}