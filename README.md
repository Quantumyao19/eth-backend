 ## 🚀 Ethereum API Service (Go)

A lightweight and extensible Ethereum backend API service built with Go.

It connects to the Ethereum blockchain via Infura RPC and provides RESTful APIs for querying on-chain data such as account balances and block numbers.

This project is designed as a learning-oriented backend system to understand Ethereum fundamentals and how blockchain data is accessed programmatically.

---

## ✨ Features

 🌐 Blockchain Integration 
* Connect to Ethereum network via Infura RPC
* Query latest block number
* Query account balance by address
* Convert WEI → ETH (human-readable format)

🔍 Transaction Insights
* Query transaction details (`/tx`)
* Query transaction receipt (`/receipt`)
* Combined trasnaction + execution result (`/tx/detail`)
* Distinguish transaction states: pending / success / failed
* Understand GasLimit vs GasUsed

🧠 Learning Focus
* Ethereum transaction lifecycle
* Gas mechanism
* Context propagation in Go
* API design with layered architecture

