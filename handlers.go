package main

import (
	"net/http"
	"strings"
	"time"
)

func HandleNextDate(w http.ResponseWriter, req *http.Request) {

	// Получаем параметры запроса
	r := req.URL.Query()

	nowStr := r.Get("now")
	dateStr := r.Get("date")
	repeatStr := r.Get("repeat")

	// Проверяем, что все параметры были переданы
	missingParams := []string{}
	if nowStr == "" {
		missingParams = append(missingParams, "now")
	}
	if dateStr == "" {
		missingParams = append(missingParams, "date")
	}
	if repeatStr == "" {
		missingParams = append(missingParams, "repeat")
	}

	if len(missingParams) > 0 {
		http.Error(w, "Не переданы следующие параметры: "+strings.Join(missingParams, ","), http.StatusBadRequest)
		return
	}

	// приводим now к нужному формату time.Time
	now, err := time.Parse(DateFormat, nowStr)
	if err != nil {
		http.Error(w, "Некорректная дата now", http.StatusBadRequest)
		return
	}

	nextDate, err := NextDate(now, dateStr, repeatStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	// в заголовок записываем тип контента, у нас это данные в формате JSON
	w.Header().Set("Content-Type", "application/json")
	// так как все успешно, то статус OK
	w.WriteHeader(http.StatusOK)
	// записываем сериализованные в JSON данные в тело ответа
	w.Write([]byte(nextDate))
}

// Обратите внимание, что вы должны добавлять задачу только в том случае,
// если запрос был отправлен методом POST.
// Дело в том, что к одному запросу /api/task будет привязано несколько действий,
// которые используют разные методы. Например, получение информации о задаче — GET-запрос /api/task?id=<число>,
// удаление задачи — DELETE-запрос /api/task?id=<число>.
// Поэтому в функции обработчика /api/task можно вставить switch,
// который будет вызывать разные функции в зависимости от используемого HTTP-метода.
func HandleTask(w http.ResponseWriter, req *http.Request) {
	switch {
	case req.Method == "POST":
		AddTask(w, req)
	case req.Method == "GET":
		FindTaskById(w, req)
	case req.Method == "PUT":
		ChangeTaskById(w, req)
	}
}

func HandleAllTasks(w http.ResponseWriter, req *http.Request) {
	FindTasks(w, req)
}
