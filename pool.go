package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type Pool struct {
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	Broadcast  chan string
}

func NewPool() *Pool {
	return &Pool{
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan string),
	}
}

func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Register:
			pool.Clients[client] = true
		case client := <-pool.Unregister:
			delete(pool.Clients, client)
		case msg := <-pool.Broadcast:
			fmt.Printf("Received : %s\n", string(msg))
			for client := range pool.Clients {
				err := client.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%p", client)))
				if err != nil {
					fmt.Printf("Error in writemessage : %s\n", err)
					// log.Fatalln(err)
					// return
				}
			}
		}
	}
}
