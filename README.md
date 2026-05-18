# exchange-api

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Gin](https://img.shields.io/badge/Gin-Framework-00BFFF?logo=go)](https://gin-gonic.com)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](https://www.docker.com)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Neon-4169E1?logo=postgresql&logoColor=white)](https://neon.tech)
[![Redis](https://img.shields.io/badge/Redis-Upstash-DC382D?logo=redis&logoColor=white)](https://upstash.com)

A currency exchange rate REST API built with Go. Fetches live rates from [ExchangeRate-API](https://www.exchangerate-api.com/), caches results in Redis, persists history in PostgreSQL, enforces per-IP rate limiting, and fires webhooks when rates cross user-defined thresholds.

> **Live demo →** `http://localhost:8080` after running locally

---

## Dashboard

![Dashboard preview](web/assets/dashboard-preview.gif)

---

## Features

- **Live rates** — `GET /rates?from=USD&to=BRL` fetches from ExchangeRate-API
- **Redis cache** — configurable TTL, avoids redundant API calls; response includes `"cached": true/false`
- **PostgreSQL history** — every live fetch is persisted; query past data with `GET /rates/history`
- **Webhooks** — register a URL + threshold; a background worker fires an HTTP POST when the rate crosses it
- **Rate limiting** — per-IP token-bucket limiter (configurable RPS + burst)
- **Swagger UI** — interactive docs at `/swagger/index.html`
- **Dashboard** — single-page HTML with Chart.js, served at `/`

---

## Tech stack

| Layer | Technology |
|---|---|
| Language | Go 1.26 |
| HTTP framework | [Gin](https://gin-gonic.com) |
| Cache | Redis via [Upstash](https://upstash.com) |
| Database | PostgreSQL via [Neon](https://neon.tech) |
| Rate source | [ExchangeRate-API](https://www.exchangerate-api.com/) (free tier) |
| Docs | [swaggo/swag](https://github.com/swaggo/swag) |

---

## Getting started

### Prerequisites

- Go 1.22+
- A free [ExchangeRate-API](https://www.exchangerate-api.com/) key
- A PostgreSQL connection string (e.g. [Neon](https://neon.tech) free tier)
- A Redis connection string (e.g. [Upstash](https://upstash.com) free tier)

### Setup

```bash
git clone https://github.com/alvimrafael/exchange-api
cd exchange-api
cp .env.example .env   # fill in your credentials
```

Run the database migrations:

```bash
psql $DATABASE_URL -f migrations/001_create_rates.sql
psql $DATABASE_URL -f migrations/002_create_webhooks.sql
```

Start the server:

```bash
make run
```

### Run with Docker

No manual migration needed - Postgres runs them automatically on firts start.

```bash
cp .env.example .env.docker   # fill in your credentials
make docker-up                # builds and starts api + postgres + redis
make docker-down              # stop containers
make docker-logs              # follow api logs
```

---

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `DATABASE_URL` | — | PostgreSQL connection string |
| `REDIS_URL` | — | Redis connection string |
| `EXCHANGE_API_KEY` | — | ExchangeRate-API key **(required)** |
| `CACHE_TTL_SECONDS` | `300` | Redis TTL for cached rates |
| `RATE_LIMIT_RPS` | `5` | Max requests per second per IP |
| `RATE_LIMIT_BURST` | `10` | Max burst per IP |
| `WEBHOOK_INTERVAL_SECONDS` | `600` | How often the webhook worker checks thresholds |

---

## API reference

Full interactive docs: `http://localhost:8080/swagger/index.html`

### `GET /health`
```bash
curl http://localhost:8080/health
# {"status":"ok","postgres":"up","redis":"up"}
```

### `GET /rates`
```bash
curl "http://localhost:8080/rates?from=USD&to=BRL"
# {"from":"USD","to":"BRL","rate":5.7423,"cached":false}

# Second request hits Redis cache
curl "http://localhost:8080/rates?from=USD&to=BRL"
# {"from":"USD","to":"BRL","rate":5.7423,"cached":true}
```

### `GET /rates/history`
```bash
curl "http://localhost:8080/rates/history?from=USD&to=BRL&days=7"
# [{"id":1,"from":"USD","to":"BRL","rate":5.7423,"cached":false,"queried_at":"2026-05-06T15:33:26Z"}]
```

### `POST /webhooks`
```bash
curl -X POST http://localhost:8080/webhooks \
  -H "Content-Type: application/json" \
  -d '{"url":"https://webhook.site/your-id","from":"USD","to":"BRL","threshold":5.80,"direction":"above"}'
# {"id":1,"url":"https://webhook.site/your-id","from":"USD","to":"BRL","threshold":5.8,"direction":"above",...}
```

### `GET /webhooks`
```bash
curl http://localhost:8080/webhooks
```

### `DELETE /webhooks/:id`
```bash
curl -X DELETE http://localhost:8080/webhooks/1
# 204 No Content
```

---

## Project structure

```
exchange-api/
├── cmd/api/main.go              # entry point — wires all dependencies
├── internal/
│   ├── handler/                 # HTTP handlers (Gin)
│   │   ├── health.go
│   │   ├── rates.go
│   │   ├── webhooks.go
│   │   └── response.go
│   ├── service/rates.go         # business logic, cache-aside pattern
│   ├── repository/              # PostgreSQL queries
│   │   ├── rates.go
│   │   └── webhooks.go
│   ├── cache/redis.go           # Redis cache abstraction
│   ├── provider/exchangerate.go # ExchangeRate-API client + interface
│   ├── middleware/ratelimit.go  # per-IP token-bucket limiter
│   └── webhook/worker.go       # background threshold checker
├── pkg/config/config.go         # env var loading
├── migrations/
│   ├── 001_create_rates.sql
│   └── 002_create_webhooks.sql
├── web/index.html               # Chart.js dashboard
├── docs/                        # generated by swag
├── Makefile
└── .env.example
```

---

## Makefile

```bash
make run      # start the server
make build    # compile → bin/api (also regenerates swagger docs)
make swagger  # regenerate docs only
make test     # run tests
make tidy     # go mod tidy
```
