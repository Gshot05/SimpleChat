package internal

import (
	"gochat/server/models"
	"strings"
)

type OnlineHandler struct {
	cl  *models.Client
	hub *models.Hub
}

func NewOnlinehandler(cl *models.Client, hub *models.Hub) *OnlineHandler {
	return &OnlineHandler{
		cl:  cl,
		hub: hub,
	}
}
func (h *OnlineHandler) BroadcastClients() {
	// Формируем список имен подключенных клиентов
	var clientList []string
	for client := range h.hub.Clients {
		clientList = append(clientList, client.Name)
	}

	// Отправляем этот список всем подключенным клиентам
	for client := range h.hub.Clients {
		go func(c *models.Client) {
			c.Send <- []byte("users:" + strings.Join(clientList, ","))
		}(client)
	}
}
