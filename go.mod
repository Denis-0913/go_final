// модуль создан командой go mod init github.com/Denis-0913/go_final

module github.com/Denis-0913/go_final

go 1.23.1

// добавили командой go get -u github.com/go-chi/chi/v5
require github.com/go-chi/chi/v5 v5.2.1 // indirect

// первый запуск теста при старте сервера выдал ошибку
// добавили модуль из теста командой go get github.com/jmoiron/sqlx
require github.com/jmoiron/sqlx v1.4.0

//go get -t github.com/Denis-0913/go_final/tests
require (
	github.com/mattn/go-sqlite3 v1.14.24
	github.com/stretchr/testify v1.10.0
)

// для создания БД устанавливаем драйвер go get modernc.org/sqlite
// go get -u github.com/mattn/go-sqlite3
require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/exp v0.0.0-20230315142452-642cacee5cc0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	modernc.org/libc v1.61.13 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.8.2 // indirect
	modernc.org/sqlite v1.35.0 // indirect
)
