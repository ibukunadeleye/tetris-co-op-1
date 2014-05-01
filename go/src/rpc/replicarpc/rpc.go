package replicarpc

type RemoteStorageServer interface {
	Get(*GetArgs, *GetReply) error

	GetList(*GetArgs, *GetListReply) error

	Put(*PutArgs, *PutReply) error

	AppendToList(*PutArgs, *PutReply) error

	RemoveFromList(*PutArgs, *PutReply) error
	
	AddServers(*AddServersArgs, *AddServersReply) error
	
	Prepare(*PrepareArgs, *Reply) error
	
	Accept(*AcceptArgs, *Reply) error
	
	Commit(*CommitVal, *Reply)error
	
}

type ReplicaServer struct {
	// Embed all methods into the struct. See the Effective Go section about
	// embedding for more details: golang.org/doc/effective_go.html#embedding
	RemoteStorageServer
}

// Wrap wraps s in a type-safe wrapper struct to ensure that only the desired
// StorageServer methods are exported to receive RPCs.
func Wrap(s RemoteStorageServer) RemoteStorageServer {
	return &ReplicaServer{s}
}