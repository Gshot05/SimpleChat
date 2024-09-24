package main

import (
	server "gochat/server/cmd"
	"log"
)

func main() {
	// Создаем HubManager
	hubManager := server.NewHub() // Вызов функции NewHub() из пакета server
	go hubManager.Run()           // Запуск хаба

	// Настраиваем маршруты и запускаем сервер
	server.SetupRoutes(hubManager)

	log.Println("Сервер запущен на порту :8080")
	log.Fatal(server.StartServer(":8080")) // Запуск сервера на порту 8080
}
