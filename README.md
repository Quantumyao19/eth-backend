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
- Kubernetes
- Prometheus / health monitoring

## API overview

The service exposes several REST endpoints for blockchain data access and transfer history:

- GET /balance?address=<ETH_ADDRESS>
  - Returns the ETH balance for a given wallet address
- GET /block
  - Returns the latest block number
- GET /tx?hash=<TX_HASH>
  - Returns transaction information by hash
- GET /receipt?hash=<TX_HASH>
  - Returns transaction receipt details
- GET /tx/detail?hash=<TX_HASH>
  - Returns transaction details along with parsed logs and transfer information
- GET /transfers?address=<ETH_ADDRESS>&page=1&page_size=20
  - Returns indexed transfer records associated with the given address

## Deployment

The project is prepared for containerized deployment and includes Kubernetes manifests under the k8s folder for:

- application deployment
- PostgreSQL
- Redis
- Prometheus
- Grafana

## Quick start

The application can be run locally with Docker Compose, but it is also designed for deployment in a Kubernetes environment.

```bash
docker compose up --build
```
