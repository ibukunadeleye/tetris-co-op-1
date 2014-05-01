package centralserver

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"net/rpc"
	"rpc/centralrpc"
	"rpc/replicarpc"
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
	//and the values are the ports of the game sever
	sync.RWMutex
	Map map[int]string
}

type Players_Synced struct {
	//Map: keys are the remote address of the player client
	sync.RWMutex
	Map map[string]*websocket.Conn
}

type CurrGS_Synced struct {
	sync.RWMutex
	ID  int //id of the active game server
	Num int //number of clients the game server has been assigned to
}

type centralServer struct {
	Port        string
	CurrGS      *CurrGS_Synced
	GameServers *GameServers_Synced
	Players     *Players_Synced
	//	Games       []*Game
	TotalGS int //number of game servers expected

	//fields used for Paxos
	replicas        map[int]*rpc.Client
	replicaHostPort map[int]string
	replicaMutex    *sync.RWMutex
	currentAcceptor int
	numNodes        int
	masterHostPort  string
	ready           bool
	readyChan       chan int
}

func checkOrigin(r *http.Request) bool {
	return true
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

func ReadFromClient (cs *centralServer, c *websocket.Conn, addr string) {
	fmt.Println("Reading from client at ", addr)
	for {
                //only expecting text messages
		_, bytes, err := c.ReadMessage()
		if err != nil {
			fmt.Println("Read Error: ", err)
			return
		}
		if len(bytes) > 0 {
			fmt.Println("received message from player at ", addr)
			fmt.Println("Player from ", addr, " re-registered")
		//if client is already registered and is requesting the central server
		//that means that a gameserver died, so grab a new game server port
		// and send it to the client
		if cs.CurrGS.Num == 0 {
			fmt.Println("CS Error: CurrGS shouldn't be empty")
			return
		}
		cs.GameServers.Lock()
		cs.CurrGS.Lock()
		var newPort string

		//check to see if this is the first or second player re-registering

		if cs.CurrGS.Num == 2 { //first player re-registering
			delete(cs.GameServers.Map, cs.CurrGS.ID)
			if len(cs.GameServers.Map) == 0 {
				fmt.Println("CS: No more game servers available")
				return
			}

			for gs_id, port := range cs.GameServers.Map {
				cs.CurrGS.ID = gs_id
				cs.CurrGS.Num = 1
				newPort = port
				break
			}
		} else { //cs.CurrGS.Num == 1 (second player re-registering)
			cs.CurrGS.Num = 2
			newPort = cs.GameServers.Map[cs.CurrGS.ID]
		}

		cs.CurrGS.Unlock()
		cs.GameServers.Unlock()
		c.WriteMessage(1, []byte(newPort))
		}

	}
}
func (cs *centralServer) Handler(w http.ResponseWriter, r *http.Request) {
	id := r.RemoteAddr
	fmt.Println("received message from ", id)
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

	//check to see if this is the first player to register ever
	//if so, send him to gameserver #1
	var gsPort string
	cs.Players.Lock()
	if len(cs.Players.Map) == 0 { //very first player
		fmt.Println("Registered Player 1 from ", id, " for the first time")
		gsPort = cs.GameServers.Map[1]
		cs.CurrGS.Lock()
		cs.CurrGS.ID = 1
		cs.CurrGS.ID = 1
		cs.CurrGS.Unlock()
	} else {
		cs.CurrGS.Lock()
		if cs.CurrGS.Num == 0 {
			fmt.Println("CS Error: CurrGS should be initialized already")
			return
		}

		fmt.Println("Registered Player 2 from ", id, " for the first time")
		cs.CurrGS.Num = 2
		gsPort = cs.GameServers.Map[cs.CurrGS.ID]
		cs.CurrGS.Unlock()
	}
	cs.Players.Unlock()
	cs.GameServers.Unlock()
	conn.WriteMessage(1, []byte(gsPort))
	go ReadFromClient(cs, conn, id)
	return
}

//numGS is the number of game servers the central server should expect
func NewCentralServer(port string, numGS, numNodes int) (CentralServer, error) {

	fmt.Println("Creating central server at localhost:", port)

	newCentralServer := &centralServer{
		Port:            port,
		CurrGS:          &CurrGS_Synced{ID: -1, Num: 0},
		GameServers:     &GameServers_Synced{Map: make(map[int]string)},
		Players:         &Players_Synced{Map: make(map[string]*websocket.Conn)},
		TotalGS:         numGS,
		numNodes:        numNodes,
		replicas:        make(map[int]*rpc.Client),
		replicaHostPort: make(map[int]string),
		replicaMutex:    new(sync.RWMutex),
		ready:           false,
		readyChan:       make(chan int)}

	http.HandleFunc("/", newCentralServer.Handler)
	go http.ListenAndServe(":"+port, nil)

	//register new central server to receive RPCs
	listener, err := net.Listen("tcp", ":"+port)
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

	go newCentralServer.waitReady()

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
		cs.GameServers.Map[args.ID] = args.Port

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

func (cs *centralServer) waitReady() {
	select {
	case <-cs.readyChan:
		cs.ready = true
		for _, replicaRPC := range cs.replicas {
			args := &replicarpc.AddServersArgs{
				ServerMap: cs.replicaHostPort,
			}
			reply := new(replicarpc.AddServersReply)
			fmt.Println("rpc addServers called")
			err := replicaRPC.Call("ReplicaServer.AddServers", args, reply)
			if err != nil {
				fmt.Println("got error from call AddServers:", err)
			}
		}
	}
}

func (cs *centralServer) RegisterReplica(args *replicarpc.RegisterArgs, reply *replicarpc.RegisterReply) error {
	hostport := args.Hostport
	fmt.Println("registering Server:", hostport)
	replicaRPC, err := rpc.DialHTTP("tcp", hostport)
	cs.replicaMutex.Lock()
	nodeID := len(cs.replicas)
	if nodeID < cs.numNodes {
		cs.replicas[nodeID] = replicaRPC
		cs.replicaHostPort[nodeID] = hostport
	}
	cs.replicaMutex.Unlock()

	fmt.Println("regisering nodeID:", nodeID)
	reply.NodeID = nodeID

	cs.replicaMutex.RLock()
	if cs.numNodes == len(cs.replicas) {
		cs.readyChan <- 1
	}
	cs.replicaMutex.RUnlock()
	fmt.Println("registered Server:", hostport)
	return err
}

func (cs *centralServer) Get(args *replicarpc.GetArgs, reply *replicarpc.GetReply) error {
	replicaRPC := cs.replicas[cs.currentAcceptor]
	call := replicaRPC.Go("ReplicaServer.Get", args, reply, nil)
	func() {
		seconds := 3
		duration := time.Duration(seconds) * time.Second
		select {
		case <-time.After(duration):
			cs.currentAcceptor = (cs.currentAcceptor + 1) % (len(cs.replicas))
			fmt.Println("node failure switching nodes")
			cs.Get(args, reply)

		case <-call.Done:
			fmt.Println("(in ss) got back:", reply.V)
		}
	}()
	return nil
}

func (cs *centralServer) Put(args *replicarpc.PutArgs, reply *replicarpc.PutReply) error {
	replicaRPC := cs.replicas[cs.currentAcceptor]
	call := replicaRPC.Go("ReplicaServer.Put", args, reply, nil)
	go func() {
		seconds := 2
		duration := time.Duration(seconds) * time.Second
		select {
		case <-time.After(duration):
			if call.Error != nil {
				fmt.Println("1got error from call.Go err:", call.Error)
			}
			cs.currentAcceptor = (cs.currentAcceptor + 1) % (len(cs.replicas))
			fmt.Println("node failure switching nodes")
			cs.Put(args, reply)

		case <-call.Done:
			fmt.Println("put was completed")
			if call.Error != nil {
				fmt.Println("2got error from call.Go err:", call.Error)
				cs.currentAcceptor = (cs.currentAcceptor + 1) % (len(cs.replicas))
				fmt.Println("node failure switching nodes")
				cs.Put(args, reply)
			}
		}
	}()
	return nil
}
