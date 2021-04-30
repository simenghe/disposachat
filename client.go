package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn *websocket.Conn
	ID   string
	Pool *Pool
}

func (c *Client) Read() {

	defer func() {
		c.Pool.Unregister <- c
		c.Conn.Close()
	}()

	ws := c.Conn
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
		}
		c.Pool.Broadcast <- string(msg)
	}
}
