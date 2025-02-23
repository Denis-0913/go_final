package main

import (
	"fmt"
	"net/http"
	"os"
)

// константа для формата даты
const (
	DateFormat = "20060102"
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

	// правила повторения задач
	http.HandleFunc("/api/nextdate", HandleNextDate)

	http.HandleFunc("/api/task", HandleTask)

	// запуск сервера
	fmt.Println("Запускаем сервер c портом: " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Ошибка при запуске сервера: %s", err.Error())
		return
	}
}

func main() {
	// Вызываем функцию CreateDatabase из файла dbSetup.go
	CreateDatabase()
	// запуск сервера
	startServer()
	fmt.Println("Завершаем работу")
}
