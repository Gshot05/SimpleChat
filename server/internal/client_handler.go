package internal

import (
	"fmt"
	"gochat/server/models"
	"log"

	"github.com/gorilla/websocket"
)

type ClientHandler struct {
	cl *models.Client
}

func NewClientHandler(cl *models.Client) *ClientHandler {
	return &ClientHandler{
		cl: cl,
	}
}

func (clHandler *ClientHandler) ReadPump(hub *models.Hub) {
	defer func() {
		hub.Broadcast <- &models.Message{Client: clHandler.cl, Message: []byte(fmt.Sprintf("%s покинул(а) чат.", clHandler.cl.Name))}
		hub.Unregister <- clHandler.cl
		clHandler.cl.Conn.Close()
	}()

	_, name, err := clHandler.cl.Conn.ReadMessage()
	if err != nil {
		log.Println("Ошибка получения имени клиента:", err)
		return
	}
	clHandler.cl.Name = string(name)
	log.Printf("Клиент присоединился с именем: %s", clHandler.cl.Name)

	for {
		_, message, err := clHandler.cl.Conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
				log.Printf("Пользователь %s покинул чат.", clHandler.cl.Name)
			} else {
				log.Printf("Ошибка чтения сообщения от %s: %v", clHandler.cl.Name, err)
			}
			break
		}
		hub.Broadcast <- &models.Message{Client: clHandler.cl, Message: message}
	}
}

func (clHandler *ClientHandler) WritePump() {
	for {
		select {
		case message, ok := <-clHandler.cl.Send:
			if !ok {
				clHandler.cl.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			clHandler.cl.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}
