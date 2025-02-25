package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// создаем структуру задачи
// так как какое то поле может не понадобиться добавляем omitempty

// Обработчик должен возвращать JSON с полем id или error
type ResponseJSON struct {
	Id    int    `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

func handleNextDate(w http.ResponseWriter, req *http.Request) {

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

	nextDate, err := nextDate(now, dateStr, repeatStr)
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
func handleTask(w http.ResponseWriter, req *http.Request) {
	switch {
	case req.Method == "POST":
		addTask(w, req)
	case req.Method == "GET":
		findTaskById(w, req)
	case req.Method == "PUT":
		changeTaskById(w, req)
	case req.Method == "DELETE":
		deleteTask(w, req)
	}

}

func handleAllTasks(w http.ResponseWriter, req *http.Request) {
	findTasks(w, req)
}

func handleTaskDone(w http.ResponseWriter, req *http.Request) {
	taskDone(w, req)
}

// JSON ошибки
func writeErrorJSON(w http.ResponseWriter, s string, err error) {
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
func addTask(w http.ResponseWriter, r *http.Request) {

	//Чтение данных из тела запроса r.Body в буфер buf
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		writeErrorJSON(w, "Ошибка чтения из тела запроса r.Body в буфер buf", err)
		return
	}

	//десериализуем данные из буфера buf в структуру Tasks.
	var task Tasks
	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		writeErrorJSON(w, "Ошибка десериализации JSON", err)
		return
	}

	//Поле title обязательное
	if task.Title == "" {
		writeErrorJSON(w, "не указан заголовок задачи", err)
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
		writeErrorJSON(w, "дата представлена в формате, отличном от 20060102", err)
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
			nextDate, err := nextDate(now, task.Date, task.Repeat)
			if err != nil {
				writeErrorJSON(w, "Ошибка NextDate: ", err) //правило повторения указано в неправильном формате в том чисде
				return
			}
			task.Date = nextDate
		}
	}

	// Получаем подключение к БД из окружения
	DBConnect := getDBFromContext(r)

	//Вставляем данные в базу
	id, s, err := DBConnect.insertTask(task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		writeErrorJSON(w, s, err)
		return
	}

	responseOk := ResponseJSON{}
	responseOk.Id = id

	// сериализуем в JSON данные в тело ответа
	jsonResponse, err := json.Marshal(responseOk)
	if err != nil {
		writeErrorJSON(w, "Ошибка сериализации: ", err)
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
func findTasks(w http.ResponseWriter, r *http.Request) {

	// Получаем подключение к БД из окружения
	DBConnect := getDBFromContext(r)

	//считываем 50 строк отсортированных по возрастанию даты
	tasks, s, err := DBConnect.selectTasks()
	if err != nil {
		writeErrorJSON(w, s, err)
		return
	}

	response := map[string]interface{}{
		"tasks": tasks,
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		writeErrorJSON(w, "Ошибка преобразования задач в JSON: ", err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

//реализуйте обработчик GET-запроса /api/task?id=<идентификатор>. Запрос должен возвращать JSON-объект со всеми полями задачи.

func findTaskById(w http.ResponseWriter, r *http.Request) {

	// считываем идентификатор
	id := r.URL.Query().Get("id")
	if id == "" {
		writeErrorJSON(w, "Не указан идентификатор!", nil)
		return
	}

	// проверяем идентификатор, что он состоит из цифр
	_, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		writeErrorJSON(w, "Не верный формат идентификатора", err)
		return
	}

	// Получаем подключение к БД из окружения
	DBConnect := getDBFromContext(r)

	//считываем задачу
	task, s, err := DBConnect.getTaskById(id)
	if err != nil {
		writeErrorJSON(w, s, err)
		return
	}

	jsonData, err := json.Marshal(task)
	if err != nil {
		writeErrorJSON(w, "Ошибка преобразования задач в JSON: ", err)
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

func changeTaskById(w http.ResponseWriter, r *http.Request) {

	// При этом данные нужно проверять так же, как при добавлении задачи.
	//Чтение данных из тела запроса r.Body в буфер buf
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		writeErrorJSON(w, "Ошибка чтения из тела запроса r.Body в буфер buf", err)
		return
	}

	//десериализуем данные из буфера buf в структуру Tasks.
	var task Tasks
	if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
		writeErrorJSON(w, "Ошибка десериализации JSON", err)
		return
	}

	//Поле title обязательное
	if task.Title == "" {
		writeErrorJSON(w, "не указан заголовок задачи", err)
		return
	}

	//Поле id обязательное
	if task.ID == "" {
		writeErrorJSON(w, "не указан идентификатор задачи", err)
		return
	}

	// проверяем идентификатор, что он состоит из цифр
	_, err = strconv.ParseInt(task.ID, 10, 32)
	if err != nil {
		writeErrorJSON(w, "Не верный формат идентификатора", err)
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
		writeErrorJSON(w, "дата представлена в формате, отличном от 20060102", err)
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
			nextDate, err := nextDate(now, task.Date, task.Repeat)
			if err != nil {
				writeErrorJSON(w, "Ошибка NextDate: ", err) //правило повторения указано в неправильном формате в том чисде
				return
			}
			task.Date = nextDate
		}
	}

	// Получаем подключение к БД из окружения
	DBConnect := getDBFromContext(r)

	//Обновляем данные в базе
	s, err := DBConnect.updateTaskInDb(task.ID, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		writeErrorJSON(w, s, err)
		return
	}

	//В случае успешного изменения должен возвращаться пустой JSON {}
	// Объявление пустой карты
	responseOk := make(map[string]interface{})

	// сериализуем в JSON данные в тело ответа
	jsonResponse, err := json.Marshal(responseOk)
	if err != nil {
		writeErrorJSON(w, "Ошибка при преобразовании в JSON:", err)
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

func writeOK(w http.ResponseWriter) {
	//В случае успешного изменения должен возвращаться пустой JSON {}
	// Объявление пустой карты
	responseOk := make(map[string]interface{})

	// сериализуем в JSON данные в тело ответа
	jsonResponse, err := json.Marshal(responseOk)
	if err != nil {
		writeErrorJSON(w, "Ошибка при преобразовании в JSON:", err)
		return
	}

	// записываем сериализованные в JSON данные в тело ответа
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func taskDone(w http.ResponseWriter, r *http.Request) {

	// считываем идентификатор
	id := r.URL.Query().Get("id")
	if id == "" {
		writeErrorJSON(w, "Не указан идентификатор!", nil)
		return
	}

	// проверяем идентификатор, что он состоит из цифр
	_, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		writeErrorJSON(w, "Не верный формат идентификатора", err)
		return
	}

	// Получаем подключение к БД из окружения
	DBConnect := getDBFromContext(r)

	//считываем задачу
	task, s, err := DBConnect.getTaskById(id)
	if err != nil {
		writeErrorJSON(w, s, err)
		return
	}

	// Одноразовая задача с пустым полем repeat удаляется.
	if task.Repeat == "" {
		_, err = DBConnect.deleteTaskById(id)
		if err != nil {
			writeErrorJSON(w, s, err)
			return
		}
		writeOK(w)
		return
	}

	// Для расчёта следующей даты используйте функцию NextDate() из начала итогового задания. Самое главное, не забудьте изменить у задачи значение колонки date на новую дату.
	date, err := nextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		writeErrorJSON(w, "Ошибка при получении NextDate: ", err)
		return
	}

	//Обновляем данные в базе
	s, err = DBConnect.updateTaskInDb(id, date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		writeErrorJSON(w, s, err)
		return
	}
	writeOK(w)
}

// Удаление задачи
// Бывают ситуации, когда задача потеряла актуальность и её нужно просто удалить.
// На фронтенде в браузере для этого есть иконка в виде корзины. Добавьте в хендлер /api/task обработку запроса с методом DELETE - /api/task/done?id=<идентификатор>.
// В этом случае нужно удалить из таблицы scheduler задачу с указанным идентификатором.
// Ответ сервера должен быть аналогичен ответу на запрос /api/task/done. Нужно возвращать {} или, в случае ошибки, JSON с полем error.
// Проверьте функцию удаления задач, как обычно, с помощью тестов go test -run ^TestDelTask$ ./tests, а затем в браузере.
func deleteTask(w http.ResponseWriter, r *http.Request) {

	// считываем идентификатор
	id := r.URL.Query().Get("id")
	if id == "" {
		writeErrorJSON(w, "Не указан идентификатор!", nil)
		return
	}

	// проверяем идентификатор, что он состоит из цифр
	_, err := strconv.ParseInt(id, 10, 32)
	if err != nil {
		writeErrorJSON(w, "Не верный формат идентификатора", err)
		return
	}

	// Получаем подключение к БД из окружения
	DBConnect := getDBFromContext(r)

	//проверяем задачу на наличие
	_, s, err := DBConnect.getTaskById(id)
	if err != nil {
		writeErrorJSON(w, s, err)
		return
	}

	_, err = DBConnect.deleteTaskById(id)
	if err != nil {
		writeErrorJSON(w, s, err)
		return
	}

	writeOK(w)
}
