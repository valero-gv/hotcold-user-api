# hotcold-user-api

Read‑through cache сервис: `GET /v1/users?user_id=`. Горячий путь — Redis, холодный — PostgreSQL. На промахе читаем из БД, кладём в Redis и отдаём ответ.

## Возможности
- Низкая латентность на кэш‑хитах, стабильный холодный путь
- Валидация `user_id` ([A-Za-z0-9_-]{1,20})
- Коды/сообщения: `200 ok:redis|ok:postgres`, `400`, `404`, `500`
- Swagger UI: `/swagger`
- Тюнинг пулов Redis/PG, `automaxprocs` для CPU лимитов контейнера
- singleflight дедупликация холодных запросов по `user_id`
- Dockerfile (distroless) + docker-compose (healthchecks)

## Стек
Go, Gin, pgx/v5, go-redis/v9, zap, envconfig, singleflight, Docker/Compose

## Схема БД
CREATE TABLE IF NOT EXISTS users (
  user_id       VARCHAR(20) PRIMARY KEY,
  deeplink      TEXT NOT NULL DEFAULT '',
  promo_message TEXT NOT NULL DEFAULT ''
);## Запуск (Docker Compose)
docker compose -f deploy/docker-compose.yml up -d --build
# применить схему:
docker compose exec -T db psql -U postgres -d app \
  -c "CREATE TABLE IF NOT EXISTS users (user_id VARCHAR(20) PRIMARY KEY, deeplink TEXT NOT NULL DEFAULT '', promo_message TEXT NOT NULL DEFAULT '');"
open http://localhost:8080/swagger## Переменные окружения (основные)
- PORT=8080
- DATABASE_URL=postgres://postgres:postgres@db:5432/app?sslmode=disable
- REDIS_ADDR=redis:6379
- CACHE_TTL_SECONDS=3600
- REQUEST_TIMEOUT_MS=300
- PG_MAX_CONNS=24
- (тюнинг) PG_HEALTHCHECK_PERIOD_SEC, PG_MAX_CONN_IDLE_TIME_SEC, PG_MAX_CONN_LIFETIME_SEC  
- (тюнинг) REDIS_POOL_SIZE, REDIS_MIN_IDLE_CONNS, REDIS_POOL_TIMEOUT_MS, REDIS_MAX_RETRIES, REDIS_CONN_MAX_IDLE_SEC

## Эндпоинты
- GET `/v1/users?user_id=ID`
  - 200: `{code:200, message: "ok:redis"|"ok:postgres", user_id, deeplink, promo_message, cache_hit, metadata}`
  - 400/404/500

## Сборка контейнера и публикация
docker buildx create --use --name builder || true
docker buildx build --platform linux/amd64,linux/arm64 \
  -t YOUR_DH_USER/hotcold-user-api:0.1.0 -t YOUR_DH_USER/hotcold-user-api:latest \
  --push .## Нагрузочное тестирование
wrk -t4 -c128 -d60s --latency 'http://127.0.0.1:8080/v1/users?user_id=someid'
# или
hey -z 60s -c 128 'http://127.0.0.1:8080/v1/users?user_id=someid'## Производительность
- Цель: ≥ 4000 req/min (≈67 rps) — достигается «из коробки»
- Рекомендации: держать p95 Redis < 2–3 ms, следить за PG MaxConns

## План улучшений
- Prometheus метрики, pprof
- golang-migrate миграции
- CI (lint+test+build+push)
- Интеграционные тесты (testcontainers)
- Rate limiting на вход и circuit breaker при необходимости
