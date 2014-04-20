package storageserver

import (
	"errors"
	"github.com/cmu440/tribbler/rpc/replicarpc"
)

type replicaServer struct {
	hostport string
	ID       uint32
	numNodes int
}

// NewReplicaServer creates and starts a new ReplicaServer.
// numNodes is the number of members in the Paxos quorum.
// port is the number of the port that the replicaserver should listen on.
// memberID is the ID the replicaserver should have
// For now, the address of all replicaservers are "localhost".
// This function should return only once all members have joined the quorum
// and should return a non-nil error if the replicaserver could not be started.
func NewReplicaServer(numNodes, port int, memberID uint32) (ReplicaServer, error) {
	return nil, errors.New("not implemented")
}

func (ss *replicaServer) Get(args *replicarpc.GetArgs, reply *replicarpc.GetReply) error {
	return errors.New("not implemented")
}

func (ss *replicaServer) GetList(args *replicarpc.GetArgs, reply *replicarpc.GetListReply) error {
	return errors.New("not implemented")
}

func (ss *replicaServer) Put(args *replicarpc.PutArgs, reply *replicarpc.PutReply) error {
	return errors.New("not implemented")
}

func (ss *replicaServer) AppendToList(args *replicarpc.PutArgs, reply *replicarpc.PutReply) error {
	return errors.New("not implemented")
}

func (ss *replicaServer) RemoveFromList(args *replicarpc.PutArgs, reply *replicarpc.PutReply) error {
	return errors.New("not implemented")
}
