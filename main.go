package main

import (
	"AutoM/config"
	"AutoM/routes"
	"log"
	"net/http"
)

func main() {
	config.InitStore()
	config.InitDB()
	defer config.CloseDB()

	router := routes.RegisterRoutes()
	log.Println("Сервер запущен на порту :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Ошибка запуска сервера: %v", err)
	}
}
