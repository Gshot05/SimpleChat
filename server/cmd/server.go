package server

import (
	"fmt"
	"gochat/server/internal"
	"gochat/server/models"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:8080" || origin == "http://127.0.0.1:8080"
	},
}

type HubManager struct {
	Hub *models.Hub
}

func NewHub() *HubManager {
	return &HubManager{
		Hub: &models.Hub{
			Clients:    make(map[*models.Client]bool),
			Broadcast:  make(chan *models.Message),
			Register:   make(chan *models.Client),
			Unregister: make(chan *models.Client),
		},
	}
}

func (h *HubManager) Run() {
	for {
		select {
		case client := <-h.Hub.Register:
			h.Hub.Clients[client] = true
			onlineHandler := internal.NewOnlinehandler(client, h.Hub) // Используем новый OnlineHandler
			onlineHandler.BroadcastClients()                          // Отправляем обновленный список клиентов
		case client := <-h.Hub.Unregister:
			if _, ok := h.Hub.Clients[client]; ok {
				delete(h.Hub.Clients, client)
				close(client.Send)
				onlineHandler := internal.NewOnlinehandler(client, h.Hub) // Используем новый OnlineHandler
				onlineHandler.BroadcastClients()                          // Отправляем обновленный список клиентов
			}
		case message := <-h.Hub.Broadcast:
			for client := range h.Hub.Clients {
				if client != message.Client {
					formattedMessage := fmt.Sprintf("Сообщение от %s: %s", message.Client.Name, message.Message)
					select {
					case client.Send <- []byte(formattedMessage):
					default:
						close(client.Send)
						delete(h.Hub.Clients, client)
					}
				}
			}
		}
	}
}

func serveWS(hub *models.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка обновления до WebSocket:", err)
		return
	}

	client := &models.Client{Conn: conn, Send: make(chan []byte)}
	handler := internal.NewClientHandler(client)

	hub.Register <- client

	go handler.WritePump()
	handler.ReadPump(hub)
}

func serveChatPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func SetupRoutes(hubManager *HubManager) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWS(hubManager.Hub, w, r)
	})
	http.HandleFunc("/chat", serveChatPage)
}

func StartServer(addr string) error {
	return http.ListenAndServe(addr, nil)
}
