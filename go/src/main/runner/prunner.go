package main

import (
	"flag"
	"log"
	"replicaserver"
)

const defaultMasterPort = 9009

var (
	port           = flag.Int("port", defaultMasterPort, "port number to listen on")
	masterHostPort = flag.String("master", "", "master storage server host port (if non-empty then this storage server is a slave)")
	numNodes       = flag.Int("N", 1, "the number of nodes in the ring (including the master)")
)

func init() {
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
}

func main() {
	flag.Parse()
	if *masterHostPort == "" && *port == 0 {
		// If masterHostPort string is empty, then this storage server is the master.
		*port = defaultMasterPort
	}
	
	var err error
	// Create and start the ReplicaServer.
	_,err = replicaserver.NewReplicaServer(*masterHostPort,*port,*numNodes)
	if err != nil {
		log.Fatalln("Failed to create storage server:", err)
	}

	// Run the storage server forever.
	select {}
}
