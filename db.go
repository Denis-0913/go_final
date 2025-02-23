package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	//_ "modernc.org/sqlite"
	_ "github.com/mattn/go-sqlite3" // в тесте проверяет дравйвер sqlite3 !!! в задании об этом ни слова
)

var db *sql.DB

// проверяем есть ли база данных, код из задания
func checkDB() (bool, string) {

	//// Получаем путь к исполняемому файлу приложения
	//appPath, err := os.Executable()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//// Определяем путь к файлу базы данных
	//dbFile := filepath.Join(filepath.Dir(appPath), "scheduler.db")
	//_, err = os.Stat(dbFile)
	// Получаем текущую рабочую директорию

	// код выше закоменчанный был дан в задании и он показывает путь некорректно
	// исправления переменных окружения на моем компьютере недоступно, поэтому
	// так как в условии сказано что БД должна лежать в деректории, файл внутри дериктории и буду определять

	//Реализуйте возможность определять путь к файлу базы данных через переменную окружения.
	//Для этого сервер должен получать значение переменной окружения TODO_DBFILE
	//и использовать его в качестве пути к базе данных, если это не пустая строка.

	currentDir := os.Getenv("TODO_DBFILE")
	if currentDir == "" {
		currentDir, _ = os.Getwd()
	}

	dbFile := filepath.Join(currentDir, "scheduler.db")
	_, err := os.Stat(dbFile)

	// Проверяем, существует ли файл базы данных
	if err == nil {
		return true, dbFile // файл сушеществует
	}
	return false, dbFile // файла нет
}

// Создаем BD и таблицу если ее нет
// функция с большой буквы, что бы ее можно было вызвать в других файлах данного пакета
func CreateDatabase() {

	checkDB, dbFile := checkDB()

	// Устанавливаем соединение с базой данных SQLite
	db, err := sql.Open("sqlite3", dbFile)
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
func InsertTask(date string, title string, comment string, repeat string) (id int, s string, err error) {

	// Устанавливаем соединение с базой данных SQLite
	// при попытке ее не закрывать в CreateDatabase() и тут не переоткрывать доступа к БД нет, хотя она и идет как переменная (даже если убираю defer db.Close())
	_, dbFile := checkDB()
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, date, title, comment, repeat)
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
