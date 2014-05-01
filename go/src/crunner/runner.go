package main

import (
	"centralserver"
	"flag"
)

const defaultPort = "9009"

var (
	port     = flag.String("port", defaultPort, "port number to listen on")
	numGS    = flag.Int("numGS", 2, "number of game servers")
	numNodes = flag.Int("numNodes", 2, "number of Paxos nodes")
)

func main() {
	flag.Parse()
	centralserver.NewCentralServer(*port, *numGS, *numNodes)
	select {}
}
