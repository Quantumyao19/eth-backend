# eth-backend

A small Go backend for Ethereum data.

It reads blockchain data from an Ethereum node, stores token transfer events in PostgreSQL, and exposes simple HTTP APIs with optional Redis caching for transfers.

## Features

- Query Ethereum account balance
- Get current block number
- Fetch transaction and receipt details
- List token transfer history by address
- Redis caching for `/transfers` responses
- Request ID, logging, and panic recovery middleware

### Notes

- `RPC_URL` is required.
- `DB_URL` is used for PostgreSQL.
- `REDIS_ADDR` defaults to `localhost:6379` if unset.
- `REDIS_DB` defaults to `0`.

## Quick Start

```powershell
cp .env.example .env
# edit .env with your settings

go run cmd/main.go
```

Then open API requests at `http://localhost:8080`.

## API Endpoints

- `GET /balance?address=<ADDRESS>`
- `GET /block`
- `GET /tx?hash=<TX_HASH>`
- `GET /receipt?hash=<TX_HASH>`
- `GET /tx/detail?hash=<TX_HASH>`
- `GET /transfers?address=<ADDRESS>&page=1&page_size=10`

## Notes

- `address` must be a valid Ethereum address.
- `page` starts at `1`.
- `page_size` controls page size.
- `/transfers` caches results in Redis for a short TTL.

## Project Structure

- `cmd/main.go` - application entry point
- `config/` - environment and configuration loading
- `internal/eth/` - Ethereum client and service layer
- `internal/db/` - database and Redis client initialization
- `internal/handler/` - HTTP handlers
- `internal/repository/` - database access logic
- `internal/server/` - HTTP server and route wiring
- `internal/middleware/` - request ID, logging, recovery middleware
- `internal/logger/` - structured logging setup
- `internal/model/` - domain models
- `utils/` - miscellaneous helper functions

## Redis Cache Behavior

- `/transfers` uses a Redis key in the form `transfer:list:<address>:<page>:<page_size>`.
- On cache miss, the handler queries PostgreSQL, then writes the response to Redis.
- The response is cached for a short duration before expiring.
