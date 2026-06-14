# Укорачиватель ссылок

HTTP-сервис для создания коротких ссылок.

## HTTP API

### Создание короткой ссылки

`POST /links`

Запрос:

```json
{
  "url": "https://example.com/some/page"
}
```

Успешный ответ, `201 Created`:

```json
{
  "short_code": "abcDEF123_",
  "short_url": "http://localhost:8080/abcDEF123_"
}
```

Если один и тот же исходный URL отправить повторно, сервис вернет уже
существующий короткий код.

### Получение исходного URL в JSON

`GET /links/{shortCode}`

Успешный ответ, `200 OK`:

```json
{
  "url": "https://example.com/some/page"
}
```

### Редирект по короткой ссылке

`GET /{shortCode}`

Успешный ответ:

```http
302 Found
Location: https://example.com/some/page
```

## Ошибки

JSON API возвращает ошибки в едином формате:

```json
{
  "error": "описание ошибки"
}
```

Коды ответов:

- `400 Bad Request` - некорректный JSON, пустой URL, некорректный URL или некорректный короткий код.
- `404 Not Found` - короткий код не найден.
- `405 Method Not Allowed` - HTTP-метод не поддерживается.
- `409 Conflict` - конфликт уникальности не удалось разрешить повторной генерацией.
- `500 Internal Server Error` - непредвиденная ошибка сервиса или хранилища.

## Хранилища

- `memory` хранит ссылки в памяти приложения. После перезапуска процесса данные теряются.
- `postgres` хранит ссылки в PostgreSQL.

Тип хранилища выбирается параметром конфигурации при запуске сервиса.

## Конфигурация

- `HTTP_ADDR` - адрес, на котором запускается HTTP-сервер. Значение по умолчанию: `:8080`.
- `BASE_URL` - публичный адрес сервиса для сборки `short_url`. Значение по умолчанию: `http://localhost:8080`.
- `STORAGE_TYPE` - тип хранилища: `memory` или `postgres`. Значение по умолчанию: `memory`.
- `DATABASE_URL` - строка подключения к PostgreSQL для режима `postgres`.


Для режима `postgres` перед запуском нужно применить SQL из `migrations/000001_create_links.up.sql`.

## Docker Compose

Перед запуском через Docker Compose создайте локальный `.env`:

```bash
cp .env.example .env
```

Запуск приложения с in-memory хранилищем:

```bash
make run-memory
```

Запуск in-memory хранилища в фоне:

```bash
make run-memory-d
```

После запуска сервис доступен по адресу:

```text
http://localhost:8080
```

Запуск приложения с PostgreSQL:

```bash
make run-postgres
```

Запуск PostgreSQL-режима в фоне:

```bash
make run-postgres-d
```

В этом режиме Compose запускает два сервиса:

- `app-postgres` - HTTP-сервис укорачивателя ссылок.
- `postgres` - PostgreSQL с healthcheck.
- `migrate` - применяет SQL-миграции через `golang-migrate`.

Миграции из папки `migrations` применяются контейнером `migrate` перед запуском
`app-postgres`.
PostgreSQL доступен с локальной машины на `localhost:5432`.

Остановить контейнеры:

```bash
make stop
```

Применить миграцию вручную на запущенном PostgreSQL:

```bash
migrate -path migrations -database 'postgres://user:pass@localhost:5432/db?sslmode=disable' up
```

Откатить миграцию вручную:

```bash
migrate -path migrations -database 'postgres://user:pass@localhost:5432/db?sslmode=disable' down
```

## Тесты

Запуск unit-тестов:

```bash
make test
```

Покрытие с учетом вызовов между пакетами:

```bash
go test ./... -coverpkg=./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

HTML-отчет:

```bash
go tool cover -html=coverage.out
```

## Завершение работы

Сервис обрабатывает `SIGINT` и `SIGTERM`. При получении сигнала HTTP-сервер
корректно завершает активные запросы через graceful shutdown с таймаутом 5 секунд.
