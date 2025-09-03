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


## Быстрый старт (Ubuntu / Windows + Docker Desktop)
```bash
cp .env.example .env
docker compose -f deploy/docker-compose.yml up -d --build

# Проверка
curl -s http://localhost:8081/healthz
curl -s http://localhost:8081/readyz

# Генерация 10 заказов
docker compose -f deploy/docker-compose.yml run --rm producer -n 10

# Возьмите один из выведенных UID и проверьте:
curl -s http://localhost:8081/order/<order_uid> | jq
# или откройте http://localhost:8081/ и введите UID
```

## Тесты
```bash
go test ./...
```

## Принятые решения
- **JSONB**: оставлен для гибкости структуры заказа, индексы добавлены.
- **Кэш**: простой TTL map+mutex с фоновой очисткой; инвалидация по TTL.
- **Ошибки**: `errors.Is` по месту; контекст через `context.Context`.
- **Моки**: ключевые интерфейсы (репозиторий/кэш) даны, в тестах — простые стабы без внешних генераторов.


## Лицензия
MIT

---

## Быстрый старт с нуля (Ubuntu 22.04)

```bash
# 0) Установите Docker + Compose (если ещё не установлены)
#   см. https://docs.docker.com/engine/install/ubuntu/  и  docker --version

# 1) Распаковка
unzip task1_new-main-fixed.zip
cd task1_new-main-fixed

# 2) Подготовка окружения
cp .env.example .env

# 3) Сборка образа приложения и запуск инфраструктуры
docker compose -f deploy/docker-compose.yml build app
docker compose -f deploy/docker-compose.yml up -d zookeeper kafka postgres migrator app

# 4) Проверки
curl -s http://localhost:8081/healthz
curl -s http://localhost:8081/readyz

# 5) Генерация тестовых заказов (10 шт.) и запрос одного из них
docker compose -f deploy/docker-compose.yml run --rm --entrypoint /producer app -n 10
curl -s http://localhost:8081/order/<ВСТАВЬТЕ_UID> | jq

# 6) Веб-страница
# откройте http://localhost:8081/ и введите order_uid
```

### Почему так?
Сервис `producer` использует образ `wb-order-service:1.0.0`. Если его не собрать локально заранее,
Docker попробует скачать его из реестра и получит `pull access denied`. Поэтому сначала `build app`,
а затем либо `up app`, либо запускайте продюсер через существующий образ `app` с `--entrypoint /producer`.
