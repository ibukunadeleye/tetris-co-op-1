package gamerpc

import "github.com/gorilla/websocket"

//defines the RPC methods that can be called on the gameserver
// This file provides a type-safe wrapper that should be used to register
// the GameServer to receive RPCs from the CentralServer

type StatusCode int

const (
	OK StatusCode = iota + 1 // The RPC was a success.
	NotOK
)

type RunGameArgs struct {
	Player1 *websocket.Conn
	Player2 *websocket.Conn
}

type RunGameReply struct {
	Status StatusCode
}

type RemoteGameServer interface {
	RunGame(args *RunGameArgs, reply *RunGameReply) error
}

type GameServer struct {
	// Embed all methods into the struct. See the Effective Go section about
	// embedding for more details: golang.org/doc/effective_go.html#embedding
	RemoteGameServer
}

func Wrap(g RemoteGameServer) RemoteGameServer {
	return &GameServer{g}
}
