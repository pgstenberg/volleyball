package main

import (
	"github.com/pgstenberg/volleyball/internal/app/server"
)

func main(){
	s := server.GameServer{Bind:"0.0.0.0:8080"}
	s.Start()
}
