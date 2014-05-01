package centralserver

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"net/rpc"
	"rpc/centralrpc"
	"sync"
	"time"
)

type Game struct {
	GameServerID int
	Player1      string
	Player2      string
	GameOver     bool //true if game is over
}

type GameServers_Synced struct {
	//Map: keys are integer IDs of the game server
	//and the values are the hostport of the game sever
	sync.RWMutex
	Map map[int]string
}

type Players_Synced struct {
	//Map: keys are the remote address of the player client
	sync.RWMutex
	Map map[string]*websocket.Conn
}

type centralServer struct {
	HostPort    string
	GameServers *GameServers_Synced
	Players     *Players_Synced
	//	Games       []*Game
	TotalGS int //number of game servers expected
}

func checkOrigin(r *http.Request) bool {
	return true
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

func (cs *centralServer) Handler(w http.ResponseWriter, r *http.Request) {
	id := r.RemoteAddr
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	//only supporting 2 players for now
	cs.Players.Lock()
	cs.Players.Map[id] = conn
	fmt.Println("Registered player from ", id)
	cs.Players.Unlock()

	//wait for at least one game server to register
	cs.GameServers.Lock()
	for len(cs.GameServers.Map) == 0 {
		cs.GameServers.Unlock()
		time.Sleep(time.Second)
		cs.GameServers.Lock()
	}
	gs := cs.GameServers.Map[1]
	cs.GameServers.Unlock()
	conn.WriteMessage(1, []byte(gs))
	fmt.Println("http handler exited")
	return
}

//numGS is the number of game servers the central server should expect
func NewCentralServer(hostport string, numGS int) (CentralServer, error) {

	fmt.Println("Creating central server at ", hostport)

	newCentralServer := &centralServer{
		HostPort:    hostport,
		GameServers: &GameServers_Synced{Map: make(map[int]string)},
		Players:     &Players_Synced{Map: make(map[string]*websocket.Conn)},
		TotalGS:     numGS}

	http.HandleFunc("/", newCentralServer.Handler)
	go http.ListenAndServe(hostport, nil)

	//register new central server to receive RPCs
	listener, err := net.Listen("tcp", hostport)
	if err != nil {
		fmt.Println("CS: ", err)
		return nil, err
	}

	err = rpc.RegisterName("CentralServer", centralrpc.Wrap(newCentralServer))
	if err != nil {
		fmt.Println("CS: ", err)
		return nil, err
	}

	rpc.HandleHTTP()
	go http.Serve(listener, nil)
	fmt.Println("Created central server successfully")
	return newCentralServer, nil
}

func (cs *centralServer) RegisterGameServer(args *centralrpc.RegisterGSArgs, reply *centralrpc.RegisterGSReply) error {
	cs.GameServers.Lock()
	_, ok := cs.GameServers.Map[args.ID]

	if ok {
		if len(cs.GameServers.Map) == cs.TotalGS {
			cs.GameServers.Unlock()
			reply.Status = centralrpc.OK
			return nil
		} else {
			cs.GameServers.Unlock()
			reply.Status = centralrpc.NotReady
			return nil
		}
	} else {
		cs.GameServers.Map[args.ID] = args.HostPort

		if len(cs.GameServers.Map) == cs.TotalGS {
			cs.GameServers.Unlock()
			reply.Status = centralrpc.OK
			fmt.Printf("CS: Registered game server #%d, all game servers registered\n", args.ID)
			return nil
		} else {
			cs.GameServers.Unlock()
			reply.Status = centralrpc.NotReady
			fmt.Printf("CS: Registered game server #%d\n", args.ID)
			return nil
		}
	}
	return nil
}
