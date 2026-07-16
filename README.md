# eth-backend

A Go backend project focused on Ethereum data access, event parsing, and transfer indexing.

This project was built to explore how to connect a backend service to blockchain data, process on-chain events, and expose useful APIs for querying account and transaction information. It combines backend development, database design, caching, and background processing in a single full-stack-style service.

## Project highlights

- Built a REST API service in Go to query Ethereum account, block, transaction, and receipt data
- Parsed ERC-20 transfer and approval events from transaction logs
- Implemented a background listener to continuously scan new blocks and index transfer activity
- Stored indexed transfer records in PostgreSQL for structured retrieval and analysis
- Used Redis to cache transfer query results and improve read performance
- Added health checks, logging, and metrics to improve service observability

## What I learned / demonstrated

This project demonstrates practical experience in:

- Backend service development with Go and Gin
- Integrating with external blockchain RPC APIs using go-ethereum
- Designing data flow between application, database, and cache
- Handling asynchronous background processing for real-time data ingestion
- Building services with basic production-minded concerns such as monitoring and reliability

## Tech stack

- Go
- Gin
- go-ethereum
- PostgreSQL
- Redis
- Docker Compose
- Prometheus / health monitoring

## Quick start

```bash
docker compose up --build
```

Example endpoints:

```bash
GET /balance?address=<ETH_ADDRESS>
GET /tx?hash=<TX_HASH>
GET /transfers?address=<ETH_ADDRESS>
```

## Run locally

```bash
go run ./cmd
```
