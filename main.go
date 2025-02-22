package main

import (
	"fmt"
	"net/http"
	"os"
)

// Зупуск сервера
func startServer() {
	// Определение порта из переменной окружения или по умолчанию 7540
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
	}

	// Настройка файлового сервера для директории ./web
	webDir := "web"
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	// запуск сервера
	fmt.Println("Запускаем сервер c портом: " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Ошибка при запуске сервера: %s", err.Error())
		return
	}
}

func main() {
	startServer()
	fmt.Println("Завершаем работу")
}
