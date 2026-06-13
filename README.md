## Ethereum API Service (Go)

A lightweight and extensible Ethereum backend API service built with Go.

It connects to the Ethereum blockchain via RPC endpoint and provides RESTful APIs for querying on-chain data such as account balances, block numbers, and detailed transaction information.

This project is designed as a learning-oriented backend system to understand Ethereum fundamentals and how blockchain data is accessed programmatically.

---

## Features

### Blockchain Integration 
* Connect to Ethereum network via RPC endpoint
* Query latest block number (`/block`)
* Query account balance by address (`/balance`)
* Convert WEI → ETH (human-readable format)

### Transaction Insights
* Query transaction details (`/tx`) - Get pending/confirmed transaction info
* Query transaction receipt (`/receipt`) - Get execution results (gas used, status)
* Combined transaction + execution result (`/tx/detail`) - Complete transaction analysis
* Distinguish transaction states: success / failed
* Understand GasLimit vs GasUsed

### ERC20 Token Events
* Parse `Transfer` events from transaction logs
* Parse `Approval` events from transaction logs
* Automatic token symbol and decimals lookup
* Human-readable token amount formatting

### Learning Focus
* Ethereum transaction lifecycle
* Gas mechanism and fee calculation
* Context propagation in Go
* API design with layered architecture (handler → service → client)
* Middleware patterns (logging, error recovery)

---

## API Endpoints

### `/balance?address=<ADDRESS>`
Query account balance on Ethereum.

**Parameters:**
- `address` (required): Ethereum address (with or without 0x prefix)

**Response:**
```json
{
  "address": "0x1234...",
  "balance_wei": "1000000000000000000",
  "balance_eth": "1.000000000000000000"
}
```

### `/block`
Get the latest block number on the Ethereum network.

**Response:**
```json
{
  "block": 21234567
}
```

### `/tx?hash=<TX_HASH>`
Query transaction details (pending or confirmed).

**Parameters:**
- `hash` (required): Transaction hash (hex format with 0x prefix)

**Response:**
```json
{
  "pending": false,
  "hash": "0xabcd...",
  "to": "0x5678...",
  "value": "1000000000000000000",
  "gas_limit": 21000,
  "nonce": 42,
  "input": "",
  "gas_price": "20000000000"
}
```

### `/receipt?hash=<TX_HASH>`
Get transaction receipt (execution results).

**Parameters:**
- `hash` (required): Transaction hash (hex format with 0x prefix)

**Response:**
```json
{
  "tx_hash": "0xabcd...",
  "status": 1,
  "gas_used": 21000
}
```

### `/tx/detail?hash=<TX_HASH>`
Get comprehensive transaction details including logs, transfers, and approvals.

**Parameters:**
- `hash` (required): Transaction hash (hex format with 0x prefix)

**Response:**
```json
{
  "hash": "0xabcd...",
  "from": "0x1111...",
  "to": "0x2222...",
  "value_eth": "1.500000000000000000",
  "gas_limit": 100000,
  "gas_used": 67890,
  "gas_price": "20000000000",
  "fee_eth": "0.0013578",
  "status": "success",
  "is_pending": false,
  "block_number": "21234567",
  "logs": [...],
  "transfers": [
    {
      "token": "0xtoken...",
      "from": "0xsender...",
      "to": "0xreceiver...",
      "value": "100.000000",
      "symbol": "USDC"
    }
  ],
  "approvals": [
    {
      "token": "0xtoken...",
      "owner": "0xowner...",
      "spender": "0xspender...",
      "value": "999999999.999999999999999999",
      "symbol": "USDC"
    }
  ]
}
```

---

## Project Structure

```
eth-backend/
├── cmd/
│   └── main.go              # Application entry point
├── config/
│   └── config.go            # Configuration management
├── internal/
│   ├── eth/
│   │   ├── client.go        # Ethereum RPC client wrapper
│   │   └── service.go       # Business logic for blockchain operations
│   ├── handler/
│   │   ├── balance.go       # Balance query handler
│   │   ├── block.go         # Block number handler
│   │   ├── tx.go            # Transaction query handler
│   │   ├── receipt.go       # Receipt query handler
│   │   ├── tx_details.go    # Detailed transaction + logs handler
│   │   └── utils.go         # Response writing utilities
│   ├── middleware/
│   │   ├── logging.go       # HTTP request/response logging
│   │   └── recover.go       # Panic recovery middleware
│   └── server/
│       └── server.go        # HTTP server setup and routing
└── utils/
    └── utils.go             # Utility functions (WEI conversion, token formatting)
```

## Usage Examples

### Query Account Balance
```bash
curl "http://localhost:8080/balance?address=0x1234567890123456789012345678901234567890"
```

### Get Latest Block
```bash
curl "http://localhost:8080/block"
```

### Query Transaction
```bash
curl "http://localhost:8080/tx?hash=0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
```

### Get Transaction Receipt
```bash
curl "http://localhost:8080/receipt?hash=0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
```

### Get Detailed Transaction with Logs
```bash
curl "http://localhost:8080/tx/detail?hash=0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
```

---

## Architecture

The application follows a **layered architecture**:

```
HTTP Handler Layer
        ↓
   Business Layer (Service)
        ↓
  Blockchain Layer (Client)
        ↓
  Ethereum RPC Node
```

### Components

**Handler Layer** (`internal/handler/`)
- Handles HTTP requests and responses
- Validates input parameters
- Manages context timeouts (5 seconds per request)
- Returns JSON responses

**Service Layer** (`internal/eth/service.go`)
- Contains business logic
- Coordinates multiple RPC calls
- Handles chain ID validation
- ERC20 token metadata retrieval

**Client Layer** (`internal/eth/client.go`)
- Wraps `go-ethereum` RPC client
- Direct interaction with Ethereum nodes
- Connection lifecycle management

**Middleware** (`internal/middleware/`)
- **Logging**: JSON-formatted request/response logs
- **Recovery**: Panic recovery to prevent server crashes

