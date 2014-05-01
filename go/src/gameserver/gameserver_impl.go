package gameserver

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	//	"net"
	"net/http"
	"net/rpc"
	"rpc/centralrpc"
	"rpc/replicarpc"
	"tetris"
	"time"
)

type Player struct {
	Socket *websocket.Conn
	ID     string
}

type GameData struct {
	Row1   int
	Col1   int
	Row2   int
	Col2   int
	Landed []([]int)
}

//only supporting two players for now
type gameServer struct {
	ID          int
	GamePort    string
	CentralPort string
	P1          *Player
	P2          *Player
}

func checkOrigin(r *http.Request) bool {
	return true
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

func (gs *gameServer) Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request from ", r.RemoteAddr)
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	player := &Player{
		ID:     r.RemoteAddr,
		Socket: conn}

	if gs.P1 == nil {
		fmt.Print("No player 1 yet")
		gs.P1 = player
		return
	} else {
		gs.P2 = player
		go gs.RunGame()
		return
	}
}

func NewGameServer(myPort string, centralPort string, id int) (GameServer, error) {
	fmt.Printf("creating new game server #%d\n", id)
	//register rpc handler for game server
	newGameServer := &gameServer{
		ID:          id,
		GamePort:    myPort,
		CentralPort: centralPort}

	http.HandleFunc("/", newGameServer.Handler)
	go http.ListenAndServe(":"+myPort, nil)

	centralHostPort := "localhost:" + centralPort

	centralServer, err := rpc.DialHTTP("tcp", centralHostPort)
	for err != nil {
		fmt.Println(err)
		time.Sleep(time.Second)
		centralServer, err = rpc.DialHTTP("tcp", centralHostPort)
	}

	args := &centralrpc.RegisterGSArgs{
		Port: myPort,
		ID:   id}

	reply := new(centralrpc.RegisterGSReply)

	for {
		err = centralServer.Call("CentralServer.RegisterGameServer", args, reply)
		if err != nil {
			fmt.Println(err)
			duration := time.Duration(1) * time.Second
			time.Sleep(duration)
			continue
		}
		if reply.Status == centralrpc.NotReady {
			fmt.Printf("GS #%d: Central Server said NotReady, trying again\n", id)
			duration := time.Duration(1) * time.Second
			time.Sleep(duration)
			continue
		}
		if reply.Status == centralrpc.OK {
			fmt.Printf("GS #%d: Game server created successfully", id)
			return newGameServer, nil
		}
	}
	return newGameServer, nil
}

func copy2D(input []([]int)) []([]int) {
	Copy := make([]([]int), len(input))
	for i := 0; i < len(input); i++ {
		Copy[i] = (input[i])[:]
	}
	return Copy
}

func (gs *gameServer) RunGame() {
	//sanity check
	if gs.P1 == nil || gs.P2 == nil {
		fmt.Println("GS Error: Not enough players")
		return
	}

	fmt.Printf("Running game with player 1 from %s and player 2 from %s\n", gs.P1.ID, gs.P2.ID)

	conn1 := gs.P1.Socket
	conn2 := gs.P2.Socket

	initCoord1 := tetris.Coord{Row: 0, Col: 1}
	initCoord2 := tetris.Coord{Row: 0, Col: 4}

	board := tetris.NewBoard(6, 6, initCoord1, initCoord2)

	var init1 = tetris.Update{
		Pos:   initCoord1,
		Value: 1}

	var init2 = tetris.Update{
		Pos:   initCoord2,
		Value: 1}

	//initial commit to storage replica
	centralHostPort := "localhost:" + gs.CentralPort
	centralServer, err := rpc.DialHTTP("tcp", centralHostPort)
	for err != nil {
		fmt.Println(err)
		time.Sleep(time.Second)
		centralServer, err = rpc.DialHTTP("tcp", centralHostPort)
	}
	reply := new(replicarpc.PutReply)

	//copy Landed array in board over to store in replica
	state := GameData{
		Row1:   board.CurrPiece1.Pos.Row,
		Col1:   board.CurrPiece1.Pos.Col,
		Row2:   board.CurrPiece2.Pos.Col,
		Col2:   board.CurrPiece2.Pos.Col,
		Landed: copy2D(board.Landed)}

	stateBytes, _ := json.Marshal(state)

	args := &replicarpc.PutArgs{V: stateBytes}

	err = centralServer.Call("CentralServer.Put", args, reply)
	if err != nil {
		fmt.Println("Put RPC to central server failed")
	}

	//initially display blocks on screen
	updates := []tetris.Update{init1, init2}
	bytes, _ := json.Marshal(updates)
	conn1.WriteMessage(1, bytes)
	conn2.WriteMessage(1, bytes)

	readerChan1 := make(chan string)
	readerChan2 := make(chan string)
	timer := time.NewTicker(time.Second)

	go readFromConn(conn1, readerChan1)
	go readFromConn(conn2, readerChan2)

	for {
		select {
		case <-timer.C:
			updates1, ok1 := board.MoveDown(1)
			updates2, ok2 := board.MoveDown(2)

			//commit game state to storage server
			state = GameData{
				Row1:   board.CurrPiece1.Pos.Row,
				Col1:   board.CurrPiece1.Pos.Col,
				Row2:   board.CurrPiece2.Pos.Col,
				Col2:   board.CurrPiece2.Pos.Col,
				Landed: copy2D(board.Landed)}

			stateBytes, _ = json.Marshal(state)
			err = centralServer.Call("CentralServer.Put", args, reply)
			if err != nil {
				fmt.Println("Put RPC to central server failed")
			}

			args := &replicarpc.PutArgs{V: stateBytes}

			err = centralServer.Call("CentralServer.Put", args, reply)
			if err != nil {
				fmt.Println("Put RPC to central server failed")
			}

			if len(updates) > 0 {
				bytes, err := json.Marshal(append(updates1, updates2...))
				if err != nil {
					fmt.Println(err)
				}
				conn1.WriteMessage(1, bytes)
				conn2.WriteMessage(1, bytes)
			}
			if !ok1 || !ok2 {
				timer.Stop()
				fmt.Println("Stopped timer")
				conn1.WriteMessage(1, []byte("GameOver"))
				conn2.WriteMessage(1, []byte("GameOver"))
			}

		case move := <-readerChan1:
			fmt.Println("received move", move)
			switch move {
			case "Right":
				bytes, err := json.Marshal(board.MoveRight(1))
				if err != nil {
					fmt.Println(err)
				}
				//commit game state to storage server
				state = GameData{
					Row1:   board.CurrPiece1.Pos.Row,
					Col1:   board.CurrPiece1.Pos.Col,
					Row2:   board.CurrPiece2.Pos.Col,
					Col2:   board.CurrPiece2.Pos.Col,
					Landed: copy2D(board.Landed)}

				stateBytes, _ = json.Marshal(state)
				err = centralServer.Call("CentralServer.Put", args, reply)
				if err != nil {
					fmt.Println("Put RPC to central server failed")
				}

				args := &replicarpc.PutArgs{V: stateBytes}
				err = centralServer.Call("CentralServer.Put", args, reply)
				if err != nil {
					fmt.Println("Put RPC to central server failed")
				}

				conn1.WriteMessage(1, bytes)
				conn2.WriteMessage(1, bytes)

			case "Left":
				bytes, err = json.Marshal(board.MoveLeft(1))
				if err != nil {
					fmt.Println(err)
				}

				//commit game state to storage server
				state = GameData{
					Row1:   board.CurrPiece1.Pos.Row,
					Col1:   board.CurrPiece1.Pos.Col,
					Row2:   board.CurrPiece2.Pos.Col,
					Col2:   board.CurrPiece2.Pos.Col,
					Landed: copy2D(board.Landed)}

				stateBytes, _ = json.Marshal(state)
				err = centralServer.Call("CentralServer.Put", args, reply)
				if err != nil {
					fmt.Println("Put RPC to central server failed")
				}

				args := &replicarpc.PutArgs{V: stateBytes}
				err = centralServer.Call("CentralServer.Put", args, reply)
				if err != nil {
					fmt.Println("Put RPC to central server failed")
				}

				conn1.WriteMessage(1, bytes)
				conn2.WriteMessage(1, bytes)
			}

		case move := <-readerChan2:
			fmt.Println("received move", move)
			switch move {
			case "Right":
				bytes, err := json.Marshal(board.MoveRight(2))
				if err != nil {
					fmt.Println(err)
				}
				//commit game state to storage server
				state = GameData{
					Row1:   board.CurrPiece1.Pos.Row,
					Col1:   board.CurrPiece1.Pos.Col,
					Row2:   board.CurrPiece2.Pos.Col,
					Col2:   board.CurrPiece2.Pos.Col,
					Landed: copy2D(board.Landed)}

				stateBytes, _ = json.Marshal(state)
				err = centralServer.Call("CentralServer.Put", args, reply)
				if err != nil {
					fmt.Println("Put RPC to central server failed")
				}

				args := &replicarpc.PutArgs{V: stateBytes}
				err = centralServer.Call("CentralServer.Put", args, reply)
				if err != nil {
					fmt.Println("Put RPC to central server failed")
				}

				conn1.WriteMessage(1, bytes)
				conn2.WriteMessage(1, bytes)

			case "Left":
				bytes, err := json.Marshal(board.MoveLeft(2))
				if err != nil {
					fmt.Println(err)
				}
				//commit game state to storage server
				state = GameData{
					Row1:   board.CurrPiece1.Pos.Row,
					Col1:   board.CurrPiece1.Pos.Col,
					Row2:   board.CurrPiece2.Pos.Col,
					Col2:   board.CurrPiece2.Pos.Col,
					Landed: copy2D(board.Landed)}

				stateBytes, _ = json.Marshal(state)
				err = centralServer.Call("CentralServer.Put", args, reply)
				if err != nil {
					fmt.Println("Put RPC to central server failed")
				}

				args := &replicarpc.PutArgs{V: stateBytes}
				err = centralServer.Call("CentralServer.Put", args, reply)
				if err != nil {
					fmt.Println("Put RPC to central server failed")
				}

				conn1.WriteMessage(1, bytes)
				conn2.WriteMessage(1, bytes)
			}
		}
	}
}

func readFromConn(c *websocket.Conn, ch chan string) {
	for {
		//only expecting text messages
		_, bytes, err := c.ReadMessage()
		if err != nil {
			fmt.Println("Read Error: ", err)
			return
		}
		if len(bytes) > 0 {
			ch <- string(bytes)
		}
	}
}
