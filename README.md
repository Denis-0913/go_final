# Файлы для итогового задания

В директории `tests` находятся тесты для проверки API, которое должно быть реализовано в веб-сервере.

Директория `web` содержит файлы фронтенда.


Коментарии разработчики

Из за использование в тесте на создание БД sqlite3, вместо sqlite
пришлось выполнить в терминале 

Из-за ошибки: Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub
Постоянное обновление значения CGO_ENABLED: go env -w CGO_ENABLED=1

Из-за ошибки: C compiler "gcc" not found: exec: "gcc": executable file not found in %PATH%
В PowerShell выполнил:
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression
scoop install mingw