# Order Service
Backend-сервис для управления заказами (Текущий контроль № 1).

## Запуск
go mod tidy
go run ./cmd/api
Сервер запускается на порту 8080.
Swagger UI: http://localhost:8080/swagger/index.html

## Проверка и тесты
go test ./...
go test ./internal/service/... -cover
