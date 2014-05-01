package centralserver

import "rpc/centralrpc"

type CentralServer interface {
	RegisterGameServer(*centralrpc.RegisterGSArgs, *centralrpc.RegisterGSReply) error
}
