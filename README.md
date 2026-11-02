# ДЗ 7. PostgreSQL и Redis в Ledger

Ledger теперь хранит транзакции и бюджеты в PostgreSQL, а отчет по тратам кеширует в Redis на 30 секунд. Gateway работает с теми же эндпоинтами, но данные приходят из базы и не пропадают после перезапуска. Как поднять базу и запустить, написано в [ledger/README.md](ledger/README.md) и [gateway/README.md](gateway/README.md).
