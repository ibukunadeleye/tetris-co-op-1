package main

import (
	"centralserver"
	"flag"
)

const defaultPort = "9009"

var (
	host  = flag.String("host", "localhost", "host address for server")
	port  = flag.String("port", defaultPort, "port number to listen on")
	numGS = flag.Int("numGS", 2, "number of game servers")
)

func main() {
	flag.Parse()
	centralserver.NewCentralServer((*host + ":" + *port), *numGS)
	select {}
}
