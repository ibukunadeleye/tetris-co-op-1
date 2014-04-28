package main

import (
	"gameserver"
	"net/http"
)

func main() {
	http.HandleFunc("/", gameserver.Handler)
	http.ListenAndServe("localhost:8080", nil)
	return
}
