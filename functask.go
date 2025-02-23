package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

// создаем структуру задачи
// так как какое то поле может не понадобиться добавляем omitempty
type Tasks struct {
	ID      string `json:"id,omitempty"`
	Date    string `json:"date,omitempty"`
	Title   string `json:"title,omitempty"`
	Comment string `json:"comment,omitempty"`
	Repeat  string `json:"repeat,omitempty"`
}

// Обработчик должен возвращать JSON с полем id или error
type ResponseJSON struct {
	Id    int    `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

// JSON ошибки
func WriteErrorJSON(w http.ResponseWriter, s string, err error) {
	responseError := ResponseJSON{}

	// если был передана ошибка error, то добавляем ее к ответу
	responseError.Error = s
	if err != nil {
		responseError.Error += ": " + err.Error()
	}

	// сериализуем в JSON данные в тело ответа
	jsonResponse, err := json.Marshal(responseError)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// записываем сериализованные в JSON данные в тело ответа
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

}

// AddTask Обработчик для добавления задачи
func AddTask(w http.ResponseWriter, r *http.Request) {

	//Чтение данных из тела запроса r.Body в буфер buf
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		WriteErrorJSON(w, "Ошибка чтения из тела запроса r.Body в буфер buf", err)
		return
	}

	//десериализуем данные из буфера buf в структуру Tasks.
	var task Tasks
	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		WriteErrorJSON(w, "Ошибка десериализации JSON", err)
		return
	}

	//Поле title обязательное
	if task.Title == "" {
		WriteErrorJSON(w, "не указан заголовок задачи", err)
		return
	}

	//Если поле date не указано или содержит пустую строку, берётся сегодняшнее число.
	now := time.Now()
	nowStr := now.Format(DateFormat)
	if task.Date == "" {
		task.Date = nowStr
	}

	// Еще обязательно проверьте, что дата указана в формате 20060102 и что функция time.Parse() корректно её распознаёт.
	date, err := time.Parse(DateFormat, task.Date)
	if err != nil {
		WriteErrorJSON(w, "дата представлена в формате, отличном от 20060102", err)
		return
	}

	dateStr := date.Format(DateFormat) // понадобится ниже

	//Если дата меньше сегодняшнего числа, есть два варианта:
	//1) если правило повторения не указано или равно пустой строке, подставляется сегодняшнее число;
	//2) при указанном правиле повторения вам нужно вычислить и записать в таблицу дату выполнения,
	// которая будет больше сегодняшнего числа. Для этого используйте функцию NextDate(), которую вы уже написали раньше.

	if dateStr < nowStr {
		if task.Repeat == "" {
			task.Date = nowStr
		} else {
			nextDate, err := NextDate(now, task.Date, task.Repeat)
			if err != nil {
				WriteErrorJSON(w, "Ошибка NextDate: ", err) //правило повторения указано в неправильном формате в том чисде
				return
			}
			task.Date = nextDate
		}
	}

	//Вставляем данные в базу
	id, s, err := InsertTask(task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		WriteErrorJSON(w, s, err)
		return
	}

	responseOk := ResponseJSON{}
	responseOk.Id = id

	// сериализуем в JSON данные в тело ответа
	jsonResponse, err := json.Marshal(responseOk)
	if err != nil {
		WriteErrorJSON(w, "Ошибка сериализации: ", err)
		return
	}

	// записываем сериализованные в JSON данные в тело ответа
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
