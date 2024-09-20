package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/websocket"
)

func main() {
	// Установите URL WebSocket-сервера
	url := "ws://localhost:8080/ws"

	// Подключаемся к WebSocket-серверу
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Ошибка подключения к серверу:", err)
	}
	defer conn.Close()

	// Запускаем горутину для получения и вывода сообщений от сервера
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("Ошибка получения сообщения:", err)
				return
			}
			fmt.Println(string(message)) // Вывод сообщения
		}
	}()

	// Запрашиваем имя пользователя
	fmt.Print("Введите свое имя: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		name := scanner.Text()
		if err := conn.WriteMessage(websocket.TextMessage, []byte(name)); err != nil {
			log.Println("Ошибка отправки имени:", err)
			return
		}
	}

	// Отправляем сообщения пользователю
	for scanner.Scan() {
		message := scanner.Text()
		if err := conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
			log.Println("Ошибка отправки сообщения:", err)
			return
		}
	}

	if scanner.Err() != nil {
		log.Println("Ошибка чтения ввода:", scanner.Err())
	}
}
