# Ledger

Бюджеты лежат в таблице `budgets`, транзакции в `expenses`. Отчет по тратам за период кешируется в Redis на 30 секунд.

## База, Redis и миграции

```bash
docker run -d --name cashcraft-postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=cashapp -p 5432:5432 postgres:16-alpine
docker run -d --name cashcraft-redis -p 6379:6379 redis:7-alpine
go install github.com/pressly/goose/v3/cmd/goose@latest
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable"
goose -dir ./migrations postgres "$DATABASE_URL" up
```

Откатить: `goose -dir ./migrations postgres "$DATABASE_URL" down`.

## Запуск

Подключение берется из `DATABASE_URL` (или из `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`, `DB_NAME`), Redis из `REDIS_ADDR`, `REDIS_DB`, `REDIS_PASSWORD`.

```bash
go run ./cmd
```

Программа подключается к базе и Redis, записывает бюджеты из `budgets.json` и печатает, что лежит в базе. Отчет строится два раза, в логе видно `cache miss`, потом `cache hit`.

## Тесты

Тесту с базой нужна отдельная база `cashapp_test` с миграциями, без `TEST_DATABASE_URL` он пропускается:

```bash
TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/cashapp_test?sslmode=disable" go test ./...
```
