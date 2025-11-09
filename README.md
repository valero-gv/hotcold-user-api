# hotcold-user-api

Read-through cache service for user lookups. Hot path is Redis; on a miss, it fetches from PostgreSQL (cold path), stores into Redis, and returns the response.

## Features
- GET /v1/users?user_id=ID
- Hot/cold paths with clear response messages:
  - 200 ok:redis (cache hit)
  - 200 ok:postgres (cold path)
  - 400 invalid user_id, 404 not found, 500 internal error
- Input validation: `[A-Za-z0-9_-]{1,20}`
- Swagger UI embedded at /swagger
- pgxpool + go-redis with tuned pools/timeouts
- singleflight to deduplicate concurrent cold lookups
- Graceful shutdown, production Dockerfile (distroless), docker-compose with healthchecks

## Quick start (Docker Compose)
```bash
docker compose -f deploy/docker-compose.yml up -d --build
# Apply schema once:
docker compose exec -T db psql -U postgres -d app \
  -c "CREATE TABLE IF NOT EXISTS users (user_id VARCHAR(20) PRIMARY KEY, deeplink TEXT NOT NULL DEFAULT '', promo_message TEXT NOT NULL DEFAULT '');"
open http://localhost:8080/swagger
```

## Endpoint
GET /v1/users?user_id=ID

- 200: `{code:200, message:"ok:redis"|"ok:postgres", user_id, deeplink, promo_message, cache_hit, metadata}`
- 400: `{code:400, message:"invalid user_id", cache_hit:false, metadata}`
- 404: `{code:404, message:"not found", cache_hit:false, metadata}`
- 500: `{code:500, message:"internal error", cache_hit:<bool>, metadata}`

## Environment variables
- PORT=8080
- DATABASE_URL=postgres://postgres:postgres@db:5432/app?sslmode=disable
- REDIS_ADDR=redis:6379
- CACHE_TTL_SECONDS=3600
- REQUEST_TIMEOUT_MS=300
- PG_MAX_CONNS=24
- Tuning (optional):
  - PG_HEALTHCHECK_PERIOD_SEC, PG_MAX_CONN_IDLE_TIME_SEC, PG_MAX_CONN_LIFETIME_SEC
  - REDIS_POOL_SIZE, REDIS_MIN_IDLE_CONNS, REDIS_POOL_TIMEOUT_MS, REDIS_MAX_RETRIES, REDIS_CONN_MAX_IDLE_SEC

## Tech stack
Go, Gin, pgx/v5, go-redis/v9, zap, envconfig, singleflight, Docker/Compose

## Load testing
Target: ≥ 4000 req/min (≈67 rps) is easily achieved.
```bash
wrk -t4 -c128 -d60s --latency 'http://127.0.0.1:8080/v1/users?user_id=someid'
# or
hey -z 60s -c 128 'http://127.0.0.1:8080/v1/users?user_id=someid'
```

## Build & publish a container (Docker Hub)
```bash
docker buildx create --use --name builder || true
docker buildx build --platform linux/amd64,linux/arm64 \
  -t YOUR_DH_USER/hotcold-user-api:0.1.0 -t YOUR_DH_USER/hotcold-user-api:latest \
  --push .
```

## Notes
- Data model (users):
```sql
CREATE TABLE IF NOT EXISTS users (
  user_id       VARCHAR(20) PRIMARY KEY,
  deeplink      TEXT NOT NULL DEFAULT '',
  promo_message TEXT NOT NULL DEFAULT ''
);
```
- For local debugging without compose, run Postgres/Redis containers and start the API with appropriate env vars.

# Go Read-Through Cache User Lookup Service

## Endpoint
GET /v1/users?user_id=ID

- 200: `{code:0, message:"ok", user_id, deeplink, promo_message, cache_hit, metadata}`
- 404: `{code:4, message:"not found", cache_hit:false, metadata}`
- 400: `{code:1, message:"invalid user_id", cache_hit:false, metadata}`

## Run locally
```bash
cp .env.example .env   # if using docker-compose
make docker-up         # starts Postgres, Redis, API
# or
make deps && make run  # runs API only; ensure local PG/Redis envs are set
```

## Migration
Apply `migrations/001_users.sql` to your database. With psql:
```bash
psql "$DATABASE_URL" -f migrations/001_users.sql
```

## Load test
Target: >= 4000 req/min (~67 rps). Example using wrk:
```bash
wrk -t4 -c128 -d60s --latency 'http://127.0.0.1:8080/v1/users?user_id=someid'
```
Alternative using hey:
```bash
hey -z 60s -c 128 'http://127.0.0.1:8080/v1/users?user_id=someid'
```

## Env
See `.env.example` for defaults. Important:
- `DATABASE_URL` (Postgres)
- `REDIS_ADDR` (Redis)
- `CACHE_TTL_SECONDS` (0 disables TTL)
- `PG_MAX_CONNS` (pool size)


