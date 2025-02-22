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
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/stretchr/testify v1.10.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
