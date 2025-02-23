package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
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
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

// AddTask Обработчик для добавления задачи

// нужно реализовать обработчик для GET-запроса /api/tasks.
// Он должен возвращать список ближайших задач в формате JSON в виде списка в поле tasks.
// Задачи должны быть отсортированы по дате в сторону увеличения.
// Каждая задача должна содержать все поля таблицы scheduler в виде строк.
// Дата представлена в уже знакомом вам формате 20060102.
func FindTasks(w http.ResponseWriter, r *http.Request) {

	//считываем 50 строк отсортированных по возрастанию даты
	tasks, s, err := SelectTasks()
	if err != nil {
		WriteErrorJSON(w, s, err)
		return
	}

	response := map[string]interface{}{
		"tasks": tasks,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		WriteErrorJSON(w, "Ошибка преобразования задач в JSON: ", err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

//реализуйте обработчик GET-запроса /api/task?id=<идентификатор>. Запрос должен возвращать JSON-объект со всеми полями задачи.

func FindTaskById(w http.ResponseWriter, r *http.Request) {

	// считываем идентификатор
	id := r.URL.Query().Get("id")
	if id == "" {
		WriteErrorJSON(w, "Не указан идентификатор!", nil)
		return
	}

	// проверяем идентификатор, что он состоит из цифр
	_, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		WriteErrorJSON(w, "Не верный формат идентификатора", err)
		return
	}

	//считываем задачу
	task, s, err := GetTaskById(id)
	if err != nil {
		WriteErrorJSON(w, s, err)
		return
	}

	jsonData, err := json.Marshal(task)
	if err != nil {
		WriteErrorJSON(w, "Ошибка преобразования задач в JSON: ", err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

//реализуйте обработчик PUT-запроса /api/task
// Если пользователь изменит какое-либо значение, в диалоговом окне появится кнопка Сохранить. П
// ри нажатии на неё фронтенд отправляет значение всех полей методом PUT по адресу /api/task.
// Данные передаются в виде JSON-объекта, как при добавлении задачи, но с полем id:
// Добавьте обработку PUT-запроса в хендлер для /api/task.
// При этом данные нужно проверять так же, как при добавлении задачи.
// В случае успешного изменения должен возвращаться пустой JSON {}, а в случае ошибки, она записывается в поле error.

func ChangeTaskById(w http.ResponseWriter, r *http.Request) {

	// При этом данные нужно проверять так же, как при добавлении задачи.
	//Чтение данных из тела запроса r.Body в буфер buf
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		WriteErrorJSON(w, "Ошибка чтения из тела запроса r.Body в буфер buf", err)
		return
	}

	//десериализуем данные из буфера buf в структуру Tasks.
	var task TasksDB
	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		WriteErrorJSON(w, "Ошибка десериализации JSON", err)
		return
	}

	//Поле title обязательное
	if task.Title == "" {
		WriteErrorJSON(w, "не указан заголовок задачи", err)
		return
	}

	//Поле id обязательное
	if task.ID == "" {
		WriteErrorJSON(w, "не указан идентификатор задачи", err)
		return
	}

	// проверяем идентификатор, что он состоит из цифр
	_, err = strconv.ParseInt(task.ID, 10, 32)
	if err != nil {
		WriteErrorJSON(w, "Не верный формат идентификатора", err)
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

	//Обновляем данные в базе
	s, err := UpdateTaskInDb(task.ID, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		WriteErrorJSON(w, s, err)
		return
	}

	//В случае успешного изменения должен возвращаться пустой JSON {}
	// Объявление пустой карты
	responseOk := make(map[string]interface{})

	// сериализуем в JSON данные в тело ответа
	jsonResponse, err := json.Marshal(responseOk)
	if err != nil {
		WriteErrorJSON(w, "Ошибка при преобразовании в JSON:", err)
		return
	}

	// записываем сериализованные в JSON данные в тело ответа
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

}

// Напишите обработчик для POST-запроса /api/task/done, который делает задачу выполненной.
// Для периодической задачи нужно рассчитать и поменять дату следующего выполнения.
// Одноразовая задача с пустым полем repeat удаляется.
// Идентификатор задачи передаётся в самом запросе /api/task/done?id=<идентификатор>.
// В случае успешного удаления возвращается пустой JSON {}, а в случае ошибки, она должна быть указана в поле error.
// Для расчёта следующей даты используйте функцию NextDate() из начала итогового задания. Самое главное, не забудьте изменить у задачи значение колонки date на новую дату.

func WriteOK(w http.ResponseWriter) {
	//В случае успешного изменения должен возвращаться пустой JSON {}
	// Объявление пустой карты
	responseOk := make(map[string]interface{})

	// сериализуем в JSON данные в тело ответа
	jsonResponse, err := json.Marshal(responseOk)
	if err != nil {
		WriteErrorJSON(w, "Ошибка при преобразовании в JSON:", err)
		return
	}

	// записываем сериализованные в JSON данные в тело ответа
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func TaskDone(w http.ResponseWriter, r *http.Request) {

	// считываем идентификатор
	id := r.URL.Query().Get("id")
	if id == "" {
		WriteErrorJSON(w, "Не указан идентификатор!", nil)
		return
	}

	// проверяем идентификатор, что он состоит из цифр
	_, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		WriteErrorJSON(w, "Не верный формат идентификатора", err)
		return
	}

	//считываем задачу
	task, s, err := GetTaskById(id)
	if err != nil {
		WriteErrorJSON(w, s, err)
		return
	}

	// Одноразовая задача с пустым полем repeat удаляется.
	if task.Repeat == "" {
		_, err = DeleteTaskById(id)
		if err != nil {
			WriteErrorJSON(w, s, err)
			return
		}
		WriteOK(w)
		return
	}

	// Для расчёта следующей даты используйте функцию NextDate() из начала итогового задания. Самое главное, не забудьте изменить у задачи значение колонки date на новую дату.
	date, err := NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		WriteErrorJSON(w, "Ошибка при получении NextDate: ", err)
		return
	}

	//Обновляем данные в базе
	s, err = UpdateTaskInDb(id, date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		WriteErrorJSON(w, s, err)
		return
	}
	WriteOK(w)
}

// Удаление задачи
// Бывают ситуации, когда задача потеряла актуальность и её нужно просто удалить.
// На фронтенде в браузере для этого есть иконка в виде корзины. Добавьте в хендлер /api/task обработку запроса с методом DELETE - /api/task/done?id=<идентификатор>.
// В этом случае нужно удалить из таблицы scheduler задачу с указанным идентификатором.
// Ответ сервера должен быть аналогичен ответу на запрос /api/task/done. Нужно возвращать {} или, в случае ошибки, JSON с полем error.
// Проверьте функцию удаления задач, как обычно, с помощью тестов go test -run ^TestDelTask$ ./tests, а затем в браузере.
func DeleteTask(w http.ResponseWriter, r *http.Request) {

	// считываем идентификатор
	id := r.URL.Query().Get("id")
	if id == "" {
		WriteErrorJSON(w, "Не указан идентификатор!", nil)
		return
	}

	// проверяем идентификатор, что он состоит из цифр
	_, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		WriteErrorJSON(w, "Не верный формат идентификатора", err)
		return
	}

	//проверяем задачу на наличие
	_, s, err := GetTaskById(id)
	if err != nil {
		WriteErrorJSON(w, s, err)
		return
	}

	_, err = DeleteTaskById(id)
	if err != nil {
		WriteErrorJSON(w, s, err)
		return
	}

	WriteOK(w)
}
