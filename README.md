# Task1 WB v2.0

Cервис на Go 1.24, который читает заказы из Kafka, сохраняет их в PostgreSQL (JSONB + индексы), кэширует в памяти (TTL), и предоставляет HTTP API и веб‑страницу для просмотра.

## Архитектура
```
[gofakeit producer] -> Kafka(topic: orders) -> App(consumer)
                                     |           |-> Postgres (orders)
                                     |           |-> In‑Memory Cache (TTL)
                                     |           \-> HTTP :8081 (/order/{id}, /healthz, /readyz, /)
```

## Технологии и версии
- Go 1.24, стандартный `net/http` + `chi`
- Kafka (bitnami/kafka:3.7) + Zookeeper (bitnami/zookeeper:3.9)
- PostgreSQL 15‑alpine, JSONB + GIN
- Миграции: `migrate/migrate:v4.17.0`
- Клиенты: `segmentio/kafka-go`, `pgx/v5`, `slog`

## Структура каталогов
```
./cmd/app            # основное приложение (HTTP + Kafka consumer)
./cmd/producer       # генератор заказов (gofakeit -> Kafka)
./internal/config    # конфигурация из ENV
./internal/model     # модели и валидация
./internal/repo      # репозиторий (Postgres/pgx)
./internal/cache     # TTL-кэш в памяти
./internal/service   # бизнес-логика (orchestration)
./internal/http      # HTTP-эндпоинты (net/http + chi)
./internal/kafka     # Kafka consumer
./migrations         # SQL миграции
./web                # простая HTML-страница
./deploy             # Dockerfile, docker-compose.yml
./scripts            # вспомогательные скрипты
./docs               # документация/диаграммы
```



## Тесты
```bash
go test ./...
```



---

## Быстрый старт с нуля (Ubuntu 22.04)

```bash


# 1) Сборка образа приложения и запуск инфраструктуры
docker compose -f deploy/docker-compose.yml build app
docker compose -f deploy/docker-compose.yml up -d zookeeper kafka postgres migrator app

# 2) Проверки
curl -s http://localhost:8081/healthz
curl -s http://localhost:8081/readyz

#3) Генерация тестовых заказов (10 шт.) и запрос одного из них
docker compose -f deploy/docker-compose.yml run --rm --entrypoint /producer app -n 10
curl -s http://localhost:8081/order/<ВСТАВЬТЕ_UID> | jq
