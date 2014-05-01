package replicarpc

import "net/rpc"
//defines the RPC methods that can be called on a replica server

type AddServersArgs struct{
	ServerMap map [int] *rpc.Client
}
type AddServersReply struct{

}

type PrepareArgs struct{
	N float64
	CS int
	HostPort string
}

type AcceptArgs struct{
	N float64
	V []byte
	CS int	
	HostPort string
}

type CommitArgs struct{
	N float64
	V []byte
	hostport string
}

type Reply struct{
	OK bool
	Value []byte
	N float64
	CS int
	Committed bool
}

type CommitVal struct{
	N float64
	V []byte
	CS int
	Committed bool
}

type GetArgs struct{
}

type GetReply struct{
	V []byte
}

type GetListReply struct{
}

type PutArgs struct{
	V []byte
}

type PutReply struct{
}
