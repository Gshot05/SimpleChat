package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:8080" || origin == "http://127.0.0.1:8080"
	},
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
	name string
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
}

type Message struct {
	client  *Client
	message []byte
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) broadcastClients() {
	// Формируем список имен подключенных клиентов
	var clientList []string
	for client := range h.clients {
		clientList = append(clientList, client.name)
	}

	// Отправляем этот список всем подключенным клиентам
	for client := range h.clients {
		select {
		case client.send <- []byte("users:" + strings.Join(clientList, ",")): // Отправляем список клиентов в виде "users:client1,client2,client3"
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.broadcastClients() // Отправляем обновленный список клиентов
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.broadcastClients() // Отправляем обновленный список клиентов
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				if client != message.client {
					formattedMessage := fmt.Sprintf("Сообщение от %s: %s", message.client.name, message.message)
					select {
					case client.send <- []byte(formattedMessage):
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		}
	}
}

func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.broadcast <- &Message{client: c, message: []byte(fmt.Sprintf("%s покинул(а) чат.", c.name))}
		hub.unregister <- c
		c.conn.Close()
	}()

	_, name, err := c.conn.ReadMessage()
	if err != nil {
		log.Println("Ошибка получения имени клиента:", err)
		return
	}
	c.name = string(name)
	log.Printf("Клиент присоединился с именем: %s", c.name)

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
				log.Printf("Пользователь %s покинул чат.", c.name)
			} else {
				log.Printf("Ошибка чтения сообщения от %s: %v", c.name, err)
			}
			break
		}
		hub.broadcast <- &Message{client: c, message: message}
	}
}

func (c *Client) writePump() {
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func serveWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка обновления до WebSocket:", err)
		return
	}

	client := &Client{conn: conn, send: make(chan []byte)}
	hub.register <- client

	go client.writePump()
	client.readPump(hub)
}

func serveChatPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../static/index.html")
}

func main() {
	hub := newHub()
	go hub.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWS(hub, w, r)
	})

	http.HandleFunc("/chat", serveChatPage)

	fmt.Println("Сервер запущен на порту :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
