package models

import "github.com/gorilla/websocket"

type Client struct {
	Conn *websocket.Conn
	Send chan []byte
	Name string
}

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan *Message
	Register   chan *Client
	Unregister chan *Client
}

type Message struct {
	Client  *Client
	Message []byte
}
