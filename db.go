package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type Database struct {
	DBConnect *sqlx.DB
}

func newDBConnect(db *sqlx.DB) Database {
	return Database{DBConnect: db}
}

// NewDatabase создает новое подключение к базе данных
func newConnection() (*Database, error) {

	_, dbFile := checkDB()

	db, err := sqlx.Connect("sqlite", dbFile)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Возвращаем структуру подключения
	return &Database{DBConnect: db}, nil
}

func (d *Database) Close() error {
	return d.DBConnect.Close()
}

// проверяем есть ли база данных, код из задания
func checkDB() (bool, string) {

	//Реализуйте возможность определять путь к файлу базы данных через переменную окружения.
	//Для этого сервер должен получать значение переменной окружения TODO_DBFILE
	//и использовать его в качестве пути к базе данных, если это не пустая строка.
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile, _ = os.Getwd()
		dbFile = filepath.Join(dbFile, "scheduler.db")
	}

	// Проверяем, существует ли файл базы данных
	_, err := os.Stat(dbFile)
	if err == nil {
		return true, dbFile // файл сушеществует
	}
	return false, dbFile // файла нет
}

// Создаем BD и таблицу если ее нет
// функция с большой буквы, что бы ее можно было вызвать в других файлах данного пакета
func createDatabase() {

	checkDB, dbFile := checkDB()

	// Устанавливаем соединение с базой данных SQLite
	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if checkDB {
		fmt.Println("База данных уже существует.")
	} else {
		// Создаем таблицу scheduler
		//  id — автоинкрементный идентификатор;
		//  date — дата задачи, которая будет хранится в формате YYYYMMDD или в Go-представлении 20060102;
		//  title — заголовок задачи;
		//  comment — комментарий к задаче;
		//  repeat — строковое поле не более 128 символов, которое будет содержать правила повторений для задачи. Формат правил будет описан в следующем шаге
		createTableQuery := `
			CREATE TABLE IF NOT EXISTS scheduler (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				date 	NVARCHAR(8),
				title 	NVARCHAR(255),
				comment NVARCHAR(1000),
				repeat 	NVARCHAR(128))`

		_, err = db.Exec(createTableQuery)
		if err != nil {
			log.Fatal(err)
		}
		//Задачи должны будут возвращаться отсортированными по дате, поэтому не забудьте создать индекс по полю date.
		createIndexQuery := `CREATE INDEX idx_date ON scheduler(date)`
		_, err = db.Exec(createIndexQuery)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Таблица и индекс созданы успешно!")
	}
}

// вставка задачи
func (d *Database) insertTask(date string, title string, comment string, repeat string) (id int, s string, err error) {

	// Устанавливаем соединение с базой данных SQLite
	// при попытке ее не закрывать в CreateDatabase() и тут не переоткрывать доступа к БД нет, хотя она и идет как переменная (даже если убираю defer db.Close())
	//_, dbFile := checkDB()
	//db, err := sql.Open("sqlite3", dbFile)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer db.Close()

	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := d.DBConnect.Exec(query, date, title, comment, repeat)
	if err != nil {
		return -1, "Ошибка вставки Task в БД", err
	}

	//Обработчик должен возвращать JSON с полем id или error
	lastInsertId, err := res.LastInsertId()
	if err != nil {
		return -1, "Ошибка получения id вставки", err
	}

	return int(lastInsertId), "", nil
}

// считываем 50 задач отсортировав по возрастанию
func (d *Database) selectTasks() ([]Tasks, string, error) {

	var tasks []Tasks

	// считываем 50 задач отсортировав по возрастанию
	err := d.DBConnect.Select(&tasks, "SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?", LimitTask)
	if err != nil {
		return nil, "Ошибка чтения Task из БД", err
	}

	//если строк нет вообще
	if tasks == nil {
		tasks = []Tasks{}
	}

	return tasks, "", nil
}

// находим задачу по id
func (d *Database) getTaskById(id string) (*Tasks, string, error) {

	var task Tasks

	// считываем 50 задач отсортировав по возрастанию
	err := d.DBConnect.Get(&task, "SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id)
	if err != nil {
		return nil, "Ошибка чтения Task из БД", err
	}

	if &task == nil {
		return nil, "Задача не найдена", err
	}

	return &task, "", nil
}

// находим задачу по id
// Если пользователь изменит какое-либо значение, в диалоговом окне появится кнопка Сохранить.
// При нажатии на неё фронтенд отправляет значение всех полей методом PUT по адресу /api/task.
// анные передаются в виде JSON-объекта, как при добавлении задачи, но с полем id:
func (d *Database) updateTaskInDb(id string, date string, title string, comment string, repeat string) (string, error) {

	sqlStmt := `
	UPDATE scheduler
	SET date = ?, title = ?, comment = ?, repeat = ? 
	WHERE id = ?`
	_, err := d.DBConnect.Exec(sqlStmt, date, title, comment, repeat, id)
	if err != nil {
		return ("Ошибка при обновлении задачи " + id), err
	}

	return "", nil
}

// Удаляем задачу по id
func (d *Database) deleteTaskById(id string) (string, error) {

	sqlStmt := `DELETE FROM scheduler WHERE id = ?`
	_, err := d.DBConnect.Exec(sqlStmt, id)
	if err != nil {
		return ("Ошибка при обновлении задачи " + id), err
	}

	return "", nil
}
