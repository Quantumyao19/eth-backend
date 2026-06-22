# eth-backend

A Go backend service for querying Ethereum data and indexing ERC-20 transfer events.

The service connects to an Ethereum JSON-RPC endpoint, exposes HTTP APIs for account and transaction data, listens for ERC-20 `Transfer` logs, stores indexed transfer records in PostgreSQL, and uses Redis to cache transfer history queries.

## Features

- Query ETH balance, latest block number, transaction data, and transaction receipts
- Parse transaction details, including gas usage, logs, ERC-20 `Transfer`, and `Approval` events
- Poll ERC-20 `Transfer(address,address,uint256)` logs and persist them to PostgreSQL
- List transfer history by address with pagination
- Cache transfer list responses in Redis with a short lock to reduce duplicate database reads
- Run database migrations with built-in `up`, `down`, and `version` commands
- Request ID, structured logging, panic recovery, and graceful shutdown

## Tech Stack

- Go 1.26.2
- go-ethereum
- PostgreSQL
- Redis
- pgx
- goqu
- golang-migrate
- zap
- Docker Compose

## Project Structure

```text
cmd/                         Application entry point
config/                      Environment configuration
internal/bootstrap/          Dependency wiring, startup flow, and migrations
internal/bootstrap/migrations/ SQL migration files
internal/db/                 PostgreSQL and Redis clients
internal/eth/                Ethereum RPC client and service layer
internal/handler/            HTTP handlers
internal/listener/           ERC-20 transfer log listener
internal/middleware/         Request ID, logging, and recovery middleware
internal/repository/         Transfer data access layer
internal/server/             HTTP route registration
utils/                       Shared helper functions
```

## Configuration

Create a `.env` file from the example:

```powershell
cp .env.example .env
```

Required variables:

| Variable | Description |
| --- | --- |
| `RPC_URL` | Ethereum JSON-RPC URL |
| `DB_URL` | PostgreSQL connection string |
| `POSTGRES_USER` | PostgreSQL user used by Docker Compose |
| `POSTGRES_PASSWORD` | PostgreSQL password used by Docker Compose |
| `POSTGRES_DB` | PostgreSQL database used by Docker Compose |

Optional variables:

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `8080` | HTTP server port |
| `CHAIN_ID` | `11155111` | Ethereum chain ID, defaulting to Sepolia |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `REDIS_PASSWORD` | Empty | Redis password |
| `REDIS_DB` | `0` | Redis database number |

For Docker Compose, use service names:

```env
DB_URL=postgres://<user>:<password>@postgres:5432/<database>?sslmode=disable
REDIS_ADDR=redis:6379
REDIS_PASSWORD=<redis_password>
POSTGRES_USER=<user>
POSTGRES_PASSWORD=<password>
POSTGRES_DB=<database>
```

For running the app locally while PostgreSQL and Redis run in Docker, use localhost:

```env
DB_URL=postgres://<user>:<password>@localhost:5432/<database>?sslmode=disable
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=<redis_password>
POSTGRES_USER=<user>
POSTGRES_PASSWORD=<password>
POSTGRES_DB=<database>
```

## Database Migrations

Migrations live in:

```text
internal/bootstrap/migrations
```

Migration files must use the `golang-migrate` naming format:

```text
001_init.up.sql
001_init.down.sql
002_add_status_column.up.sql
002_add_status_column.down.sql
```

Run all pending migrations:

```powershell
go run ./cmd migrate up
```

Rollback the latest migration:

```powershell
go run ./cmd migrate down
```

Rollback multiple migrations:

```powershell
go run ./cmd migrate down 2
```

Check the current migration version:

```powershell
go run ./cmd migrate version
```

When using Docker Compose, the `migrate` service runs `./main migrate up` automatically before the app starts.

## Run with Docker Compose

After creating `.env`, start the full stack:

```powershell
docker compose up --build
```

Docker Compose starts PostgreSQL and Redis, runs database migrations, and then starts the app.

To run only migrations through Docker Compose:

```powershell
docker compose run migrate ./main migrate up
docker compose run migrate ./main migrate down
docker compose run migrate ./main migrate version
```

## API Endpoints

### `GET /balance`

Query ETH balance by address.

```http
GET /balance?address=<ETH_ADDRESS>
```

### `GET /block`

Get the latest block number.

```http
GET /block
```

### `GET /tx`

Get basic transaction data by transaction hash.

```http
GET /tx?hash=<TX_HASH>
```

### `GET /receipt`

Get transaction receipt data.

```http
GET /receipt?hash=<TX_HASH>
```

### `GET /tx/detail`

Get transaction details, raw logs, and parsed ERC-20 `Transfer` / `Approval` events.

```http
GET /tx/detail?hash=<TX_HASH>
```

### `GET /transfers`

List ERC-20 transfer records where the address appears as sender or recipient.

```http
GET /transfers?address=<ETH_ADDRESS>&page=1&page_size=20
```

Response shape:

```json
{
  "page": 1,
  "page_size": 20,
  "total": 0,
  "data": []
}
```

## Transfer Indexing

The listener checks the latest block every 5 seconds and fetches ERC-20 `Transfer` logs by block range. Parsed records are inserted into `token_transfers` with a unique index on `(tx_hash, log_index)` to avoid duplicate writes.

Transfer list queries are cached in Redis using:

```text
transfer:list:<address>:<page>:<page_size>
```

On cache miss, the handler uses a short Redis lock so only one request refreshes the same query result while other requests briefly retry the cache.
