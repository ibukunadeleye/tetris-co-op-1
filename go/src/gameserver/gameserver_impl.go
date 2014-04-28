package gameserver

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"tetris"
	"time"
)

func checkOrigin(r *http.Request) bool {
	return true
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

func Handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Host)
	fmt.Println(r.RemoteAddr)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	go handleConn(conn)
	fmt.Println("http handler exited")
}

func handleConn(c *websocket.Conn) {
	initCoord := tetris.Coord{Row: 0, Col: 1}

	board := tetris.NewBoard(4, 4)

	var initUpdate = tetris.Update{
		Pos:   initCoord,
		Value: 1}

	//initially display block on screen
	updates := []tetris.Update{initUpdate}
	bytes, _ := json.Marshal(updates)
	c.WriteMessage(1, bytes)

	readerChan := make(chan string)
	timer := time.NewTicker(time.Second)

	go readFromConn(c, readerChan)
	for {
		select {
		case <-timer.C:
			updates, ok := board.MoveDown()
			if len(updates) > 0 {
				bytes, err := json.Marshal(updates)
				if err != nil {
					fmt.Println(err)
				}
				c.WriteMessage(1, bytes)
			}
			if !ok {
				timer.Stop()
				fmt.Println("Stopped timer")
				c.WriteMessage(1, []byte("GameOver"))
				return
			}

		case move := <-readerChan:
			fmt.Println("received move", move)
			switch move {
			case "Right":
				bytes, err := json.Marshal(board.MoveRight())
				if err != nil {
					fmt.Println(err)
				}
				c.WriteMessage(1, bytes)

			case "Left":
				bytes, err := json.Marshal(board.MoveLeft())
				if err != nil {
					fmt.Println(err)
				}
				c.WriteMessage(1, bytes)
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
