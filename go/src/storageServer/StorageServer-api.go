package StorageServer

import "github.com/cmu440/tribbler/rpc/storagerpc"

// StorageServer defines the set of methods that can be invoked
//remotely via RPCs.
type StorageServer interface {

	// Get retrieves the specified key from the data store and replies with
	// the key's value. If the key is not found, it should reply with status
	// KeyNotFound.
	Get(*storagerpc.GetArgs, *storagerpc.GetReply) error

	// GetList retrieves the specified key from the data store and replies with
	// the key's list value. If the key is not found, it should reply with status
	// KeyNotFound.
	GetList(*storagerpc.GetArgs, *storagerpc.GetListReply) error

	// Put inserts the specified key/value pair into the data store.
	Put(*storagerpc.PutArgs, *storagerpc.PutReply) error

	// AppendToList retrieves the specified key from the data store and appends
	// the specified value to its list. If the value is already contained in the
	//list, it should reply with status ItemExists.
	AppendToList(*storagerpc.PutArgs, *storagerpc.PutReply) error

	// RemoveFromList retrieves the specified key from the data store and removes
	// the specified value from its list. If the specified value is not already
	//contained in the list, it should reply with status ItemNotFound.
	RemoveFromList(*storagerpc.PutArgs, *storagerpc.PutReply) error
}
