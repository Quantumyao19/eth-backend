# eth-backend

A small Go backend for Ethereum data.

It reads blockchain data from an Ethereum node, stores token transfer events, and exposes simple HTTP APIs.

## Quick Start

1. Configure your environment (RPC URL, database connection).
2. Run:

```powershell
go run cmd/main.go
```

3. Open API requests at `http://localhost:8080`.

## Endpoints

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
- `/transfers` returns transfer history for an address.

## Structure

- `cmd/main.go` - startup
- `internal/eth/` - Ethereum client and service logic
- `internal/handler/` - HTTP handlers
- `internal/repository/` - database access
- `internal/server/` - server routing
- `internal/middleware/` - logging and recovery
- `utils/` - helper functions
