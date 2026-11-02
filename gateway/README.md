# Gateway

Эндпоинты те же, что раньше, но Ledger берет данные из PostgreSQL. Добавился `GET /api/reports/summary?from=YYYY-MM-DD&to=YYYY-MM-DD`, отчет кешируется в Redis.

## Запуск

Сначала нужно поднять базу и применить миграции, см. [../ledger/README.md](../ledger/README.md).

```bash
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/cashapp?sslmode=disable"
export REDIS_ADDR=localhost:6379
go run .
```

## Проверка через cURL

```bash
curl -X POST http://localhost:8080/api/budgets \
  -H "Content-Type: application/json" \
  -d '{"category":"еда","limit":1000}'
```

```bash
curl -X POST http://localhost:8080/api/transactions \
  -H "Content-Type: application/json" \
  -d '{"amount":450,"category":"еда","description":"ланч","date":"2026-11-01"}'
```

Еще 700 на еду не влезут в лимит, придет 409:

```bash
curl -X POST http://localhost:8080/api/transactions \
  -H "Content-Type: application/json" \
  -d '{"amount":700,"category":"еда","description":"ужин","date":"2026-11-02"}'
```

```bash
curl http://localhost:8080/api/transactions
curl http://localhost:8080/api/budgets
```

Отчет два раза подряд, в логе Gateway будет `cache miss`, потом `cache hit`:

```bash
curl "http://localhost:8080/api/reports/summary?from=2026-11-01&to=2026-11-30"
curl "http://localhost:8080/api/reports/summary?from=2026-11-01&to=2026-11-30"
```

После перезапуска Gateway данные остаются.
