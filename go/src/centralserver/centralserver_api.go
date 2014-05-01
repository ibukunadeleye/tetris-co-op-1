package centralserver

import "rpc/centralrpc"
import "rpc/replicarpc"

type CentralServer interface {
	RegisterGameServer(*centralrpc.RegisterGSArgs, *centralrpc.RegisterGSReply) error
	//this is used to register the replicas to the starter Server
	RegisterReplica(*replicarpc.RegisterArgs, *replicarpc.RegisterReply) error

	//this is a normal Get arguement retrieves the last state of the game
	Get(*replicarpc.GetArgs, *replicarpc.GetReply) error

	//this is the method that places new updates to the system
	Put(*replicarpc.PutArgs, *replicarpc.PutReply) error
}
