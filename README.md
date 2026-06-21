# eth-backend

A Go backend service for querying Ethereum data and indexing ERC-20 transfer events.

The service connects to an Ethereum JSON-RPC endpoint, exposes HTTP APIs for account and transaction data, listens for ERC-20 `Transfer` logs, stores indexed transfer records in PostgreSQL, and uses Redis to cache transfer history queries.

## Features

- Query ETH balance, latest block number, transaction data, and transaction receipts
- Parse transaction details, including gas usage, logs, ERC-20 `Transfer`, and `Approval` events
- Poll ERC-20 `Transfer(address,address,uint256)` logs and persist them to PostgreSQL
- List transfer history by address with pagination
- Cache transfer list responses in Redis with a short lock to reduce duplicate database reads
- Request ID, structured logging, panic recovery, and graceful shutdown

## Tech Stack

- Go 1.26.2
- go-ethereum
- PostgreSQL
- Redis
- pgx
- goqu
- zap
- Docker Compose

## Project Structure

```text
cmd/                    Application entry point
config/                 Environment configuration
internal/bootstrap/     Dependency wiring and startup flow
internal/db/            PostgreSQL, Redis, and DDL
internal/eth/           Ethereum RPC client and service layer
internal/handler/       HTTP handlers
internal/listener/      ERC-20 transfer log listener
internal/middleware/    Request ID, logging, and recovery middleware
internal/repository/    Transfer data access layer
internal/server/        HTTP route registration
utils/                  Shared helper functions
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

Optional variables:

| Variable | Default | Description |
| --- | --- | --- |
| `PORT` | `8080` | HTTP server port |
| `CHAIN_ID` | `11155111` | Ethereum chain ID, defaulting to Sepolia |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `REDIS_PASSWORD` | Empty | Redis password |
| `REDIS_DB` | `0` | Redis database number |

For Docker Compose, use service names instead of `localhost`:

```env
DB_URL=postgres://<user>:<password>@postgres:5432/<database>?sslmode=disable
REDIS_ADDR=redis:6379
REDIS_PASSWORD=<redis_password>
```

For running the app locally while PostgreSQL and Redis run in Docker:

```env
DB_URL=postgres://<user>:<password>@localhost:5432/<database>?sslmode=disable
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=<redis_password>
```

## Database Setup

The app expects the `token_transfers` table to exist. Run the DDL before starting the service:

```powershell
docker compose up -d postgres
Get-Content internal/db/ddl/001_create_token_transfers_table.up.sql | docker exec -i postgres psql -U <user> -d <database>
```

DDL file:

```text
internal/db/ddl/001_create_token_transfers_table.up.sql
```

## Run Locally

Start PostgreSQL and Redis:

```powershell
docker compose up -d postgres redis
```

Start the Go service:

```powershell
go run ./cmd
```

The server listens on:

```text
http://localhost:8080
```

## Run with Docker Compose

After creating `.env` and initializing the database table:

```powershell
docker compose up --build
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
