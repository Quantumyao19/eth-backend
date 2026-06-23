# eth-backend

Go backend for reading Ethereum data and indexing ERC-20 transfer logs.

It connects to an Ethereum JSON-RPC node, exposes simple HTTP APIs, stores parsed `Transfer` events in PostgreSQL, and caches transfer list queries in Redis.

## Features

- Query ETH balance, latest block, transaction, and receipt
- Parse transaction logs for ERC-20 `Transfer` and `Approval` events
- Poll chain logs for `Transfer(address,address,uint256)`
- Store token transfer records in PostgreSQL
- Cache transfer list responses in Redis
- Run SQL migrations from the app binary

## Stack

- Go 1.26.2
- go-ethereum
- PostgreSQL
- Redis
- goqu / pgx
- Docker Compose

## Project Structure

```text
cmd/                         app entry
config/                      env config
internal/bootstrap/          startup and migrations
internal/bootstrap/migrations SQL migrations
internal/db/                 postgres and redis clients
internal/eth/                Ethereum RPC service
internal/handler/            HTTP handlers
internal/listener/           ERC-20 transfer listener
internal/repository/         database access
internal/server/             routes and HTTP server
utils/                       shared helpers
```

## Config

Create `.env` from the example:

```powershell
cp .env.example .env
```

Required:

```env
RPC_URL=your_rpc_url_here
DB_URL=postgres://user:password@postgres:5432/db?sslmode=disable
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_DB=db
REDIS_PASSWORD=password
```

Optional:

```env
PORT=8080
CHAIN_ID=11155111
REDIS_ADDR=redis:6379
REDIS_DB=0
```

When running the app locally but PostgreSQL/Redis in Docker, use:

```env
DB_URL=postgres://user:password@localhost:5432/db?sslmode=disable
REDIS_ADDR=localhost:6379
```

## Run

Start the full stack:

```powershell
docker compose up --build
```

Run locally:

```powershell
go run ./cmd
```

## Migrations

```powershell
go run ./cmd migrate up
go run ./cmd migrate down
go run ./cmd migrate version
```

Migration files are in:

```text
internal/bootstrap/migrations
```

## APIs

### Balance

```http
GET /balance?address=<ETH_ADDRESS>
```

### Latest Block

```http
GET /block
```

### Transaction

```http
GET /tx?hash=<TX_HASH>
```

### Receipt

```http
GET /receipt?hash=<TX_HASH>
```

### Transaction Detail

Returns transaction info, raw logs, parsed ERC-20 transfers, and approvals.

```http
GET /tx/detail?hash=<TX_HASH>
```

### Transfers

Lists indexed ERC-20 transfer records by address.

```http
GET /transfers?address=<ETH_ADDRESS>&page=1&page_size=20
```

Current query behavior matches records where the address is in `from_address`, `to_address`, or `token_address`.

Response:

```json
{
  "page": 1,
  "page_size": 20,
  "total": 0,
  "data": []
}
```

## Transfer Indexing

The listener polls every 5 seconds, fetches ERC-20 `Transfer` logs, and stores:

- `tx_hash`
- `log_index`
- `block_number`
- `token_address`
- `from_address`
- `to_address`
- `value`

Duplicate records are ignored by the unique index on `(tx_hash, log_index)`.
