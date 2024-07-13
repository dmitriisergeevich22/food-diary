# Сервис food-diary.
[GitLab]()

Food-diary представляет собой сервис, отвечающий за контроль калорийности рациона и сбор статистики самочувствия и после употребления.



## Запуск приложения локально.
### Запуск необходимой инфрастуктуры.
Для запуска необходимой инфрастуктуры приложения, достаточно иметь установленный docker и запустить команду:

`docker-compose up -d`

### Запуск сервиса.
Для запуска сервиса используется команда:

`go run cmd/app/main.go -envPath=config/local.env`

## Тестирование.

### End2End тесты.
Для запуска end2end необходимо установка docker, goose на локальной машине.
Команда запуска end2end тестов:

`go test -timeout 30s -run ^TestEnd2End$ template/test`

## Генерация proto service и swagger.
Для генерации proto service и swagger достаточно запустить команду `make generate-proto-and-swagger` 