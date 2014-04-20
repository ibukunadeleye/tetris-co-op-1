package storageserver

import (
	"errors"
	"github.com/cmu440/tribbler/rpc/storagerpc"
)

type storageServer struct {
	hostport string
	ID       uint32
	numNodes int
}

// NewStorageServer creates and starts a new StorageServer.
// numNodes is the number of members in the Paxos quorum.
// port is the number of the port that the storageserver should listen on.
// memberID is the ID the storageserver should have
// For now, the address of all storageservers are "localhost".
// This function should return only once all members have joined the quorum
// and should return a non-nil error if the storageserver could not be started.
func NewStorageServer(numNodes, port int, memberID uint32) (StorageServer, error) {
	return nil, errors.New("not implemented")
}

func (ss *storageServer) Get(args *storagerpc.GetArgs, reply *storagerpc.GetReply) error {
	return errors.New("not implemented")
}

func (ss *storageServer) GetList(args *storagerpc.GetArgs, reply *storagerpc.GetListReply) error {
	return errors.New("not implemented")
}

func (ss *storageServer) Put(args *storagerpc.PutArgs, reply *storagerpc.PutReply) error {
	return errors.New("not implemented")
}

func (ss *storageServer) AppendToList(args *storagerpc.PutArgs, reply *storagerpc.PutReply) error {
	return errors.New("not implemented")
}

func (ss *storageServer) RemoveFromList(args *storagerpc.PutArgs, reply *storagerpc.PutReply) error {
	return errors.New("not implemented")
}

//implementation for storage server goes here
