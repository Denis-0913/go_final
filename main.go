package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	//формат даты
	DateFormat = "20060102"
	//Лимит на кол-во строк при запросе всех задач
	LimitTask = 50
)

// создаем структуру задач для json и db
type Tasks struct {
	ID      string `db:"id" json:"id", omitempty`
	Date    string `db:"date" json:"date", omitempty`
	Title   string `db:"title" json:"title", omitempty`
	Comment string `db:"comment" json:"comment", omitempty`
	Repeat  string `db:"repeat" json:"repeat", omitempty`
}

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

	// запускаем подключение - работает до тех пор пока сервер не остановят
	DBConnect, err := newConnection()
	if err != nil {
		log.Fatal("Ошибка при создании подключения:", err)
	}
	defer DBConnect.Close()

	// правила повторения задач
	http.HandleFunc("/api/nextdate", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "DBConnect", DBConnect)
		r = r.WithContext(ctx)
		handleNextDate(w, r)
	})

	http.HandleFunc("/api/task/done", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "DBConnect", DBConnect)
		r = r.WithContext(ctx)
		handleTaskDone(w, r)
	})

	http.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "DBConnect", DBConnect)
		r = r.WithContext(ctx)
		handleTask(w, r)
	})

	http.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "DBConnect", DBConnect)
		r = r.WithContext(ctx)
		handleAllTasks(w, r)
	})

	// запуск сервера
	fmt.Println("Запускаем сервер c портом: " + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Ошибка при запуске сервера: %s", err.Error())
		return
	}
}

// Функция для получения подключения из контекста
func getDBFromContext(r *http.Request) *Database {
	db, ok := r.Context().Value("DBConnect").(*Database)
	if !ok {
		log.Fatal("Подключение к БД не найдено в контексте")
	}
	return db
}

func main() {
	// создаем БД с таблицей если ее нет и сразу закрываем подключение - внутри прописано defer db.Close() - используем database/sql - этот кусок переписывать не стал
	createDatabase()

	// запуск сервера
	startServer()
}
