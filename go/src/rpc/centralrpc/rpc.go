package centralrpc

import "rpc/replicarpc"

//contains the RPC methods that can be called on the centralserver

type StatusCode int

const (
	OK       StatusCode = iota + 1 // The RPC was a success.
	NotReady                       //sent if not all game servers have finished registering
	NotOK
)

type RegisterGSArgs struct {
	Port string
	ID   int
}

type RegisterGSReply struct {
	Status StatusCode
}

type RemoteCentralServer interface {
	RegisterGameServer(*RegisterGSArgs, *RegisterGSReply) error
	RegisterReplica(*replicarpc.RegisterArgs, *replicarpc.RegisterReply) error
	Get(*replicarpc.GetArgs, *replicarpc.GetReply) error
	Put(*replicarpc.PutArgs, *replicarpc.PutReply) error
}

type CentralServer struct {
	// Embed all methods into the struct. See the Effective Go section about
	// embedding for more details: golang.org/doc/effective_go.html#embedding
	RemoteCentralServer
}

// Wrap wraps l in a type-safe wrapper struct to ensure that only the desired
// LeaseCallbacks methods are exported to receive RPCs.
func Wrap(c RemoteCentralServer) RemoteCentralServer {
	return &CentralServer{c}
}
