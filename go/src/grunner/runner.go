package main

import (
	"flag"
	"fmt"
	"gameserver"
)

const defaultPort = "9010"

var (
	gamePort    = flag.String("game_port", defaultPort, "port number that game server will listen on")
	id          = flag.Int("id", 0, "ID of the game server")
	centralPort = flag.String("central_port", defaultPort, "port number that central server will listen on")
)

func main() {
	flag.Parse()
	//call newGameServer with at the designated host:port and id
	_, err := gameserver.NewGameServer(*gamePort, *centralPort, *id)
	if err != nil {
		fmt.Println("Grunner Error: ", err)
	} else {
		fmt.Println("NewGameServer() exited")
	}

	select {}
}
