package handler

import (
	"fmt"
	"golang.org/x/net/websocket"
	"log"
)

func NewWebSocketHandler() websocket.Handler {
	return websocket.Handler(echoHandler)
}
func echoHandler(ws *websocket.Conn) {
	msg := make([]byte, 512)
	for {
		n, err := ws.Read(msg)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Receive: %s\n", msg[:n])
	}
	/*
		send_msg := "[" + string(msg[:n]) + "]"
		m, err := ws.Write([]byte(send_msg))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Send: %s\n", msg[:m])
	*/
}
