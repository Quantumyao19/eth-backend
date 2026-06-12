package handler

import (
	"context"
	"log"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"eth-backend/utils"

	"github.com/ethereum/go-ethereum/common"
)

type TxDetailResponse struct {
	Hash     string `json:"hash"`
	From     string `json:"from"`
	To       string `json:"to"`
	ValueEth string `json:"value_eth"`

	GasLimit uint64 `json:"gas_limit"`
	GasUsed  uint64 `json:"gas_used"`
	GasPrice string `json:"gas_price"`
	FeeEth   string `json:"fee_eth"`

	Status    string `json:"status"`
	IsPending bool   `json:"is_pending"`

	BlockNumber string `json:"block_number,omitempty"`

	Logs      []LogRaw
	Transfers []TransferLog
}

type LogRaw struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
	Data    string   `json:"data"`
}

type TransferLog struct {
	Token  string `json:"token"`
	From   string `json:"from"`
	To     string `json:"to"`
	Value  string `json:"value"`
	Symbol string `json:"symbol"`
}

const erc20ABI = `[{
	"anonymous": false,
	"inputs": [
		{"indexed": true, "name": "from", "type": "address"},
		{"indexed": true, "name": "to", "type": "address"},
		{"indexed": false, "name": "value", "type": "uint256"}
	],
	"name": "Transfer",
	"type": "event"
}]`

func (h *Handler) TxDetail(w http.ResponseWriter, r *http.Request) {
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		writeError(w, "missing hash", http.StatusBadRequest)
		return
	}

	if !common.IsHexHash(hash) {
		writeError(w, "invalid hash format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), defaultTimeout)
	defer cancel()

	// obtain transaction
	tx, isPending, err := h.service.GetTransaction(ctx, common.HexToHash(hash))
	if err != nil {
		log.Println("GetTransaction error:", err)
		handleError(w, err)
		return
	}

	if tx == nil {
		writeError(w, "transaction not found", http.StatusNotFound)
		return
	}

	// obtain receipt
	receipt, err := h.service.GetTransactionReceipt(ctx, common.HexToHash(hash))
	if err != nil {
		log.Println("GetReceipt error:", err)
		handleError(w, err)
		return
	}

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		log.Println("Parsed ABI error:", err)
		handleError(w, err)
		return
	}

	var logList []LogRaw
	var transfers []TransferLog

	for _, l := range receipt.Logs {
		if len(l.Topics) == 0 {
			continue
		}

		event, err := parsedABI.EventByID(l.Topics[0])
		if err != nil {
			continue
		}

		if event.Name != "Transfer" || len(l.Topics) < 3 {
			continue
		}

		var data struct {
			Value *big.Int
		}

		if err := parsedABI.UnpackIntoInterface(&data, "Transfer", l.Data); err != nil {
			log.Println("decode transfer error:", err)
			continue
		}

		from := common.HexToAddress(l.Topics[1].Hex())
		to := common.HexToAddress(l.Topics[2].Hex())

		symbol := "UNKNOWN"
		decimals := uint8(18)

		s, d, err := h.service.GetTokenMeta(ctx, l.Address)
		if err == nil {
			symbol = s
			decimals = d
		}

		transfers = append(transfers, TransferLog{
			Token:  l.Address.Hex(),
			From:   from.Hex(),
			To:     to.Hex(),
			Value:  utils.FormatTokenAmount(data.Value, decimals),
			Symbol: symbol,
		})

		var topics []string
		for _, t := range l.Topics {
			topics = append(topics, t.Hex())
		}

		logList = append(logList, LogRaw{
			Address: l.Address.Hex(),
			Topics:  topics,
			Data:    "0x" + common.Bytes2Hex(l.Data),
		})
	}

	gasUsed := uint64(0)
	status := "pending"
	blockNumber := ""
	if receipt != nil {
		gasUsed = receipt.GasUsed
		blockNumber = receipt.BlockNumber.String()
		if receipt.Status == 1 {
			status = "success"
		} else {
			status = "failed"
		}
	}

	to := ""
	if tx.To() != nil {
		to = tx.To().Hex()
	}

	from, err := h.service.GetTransactionSender(ctx, tx)
	if err != nil {
		log.Println("GetTransactionSender error:", err)
		handleError(w, err)
		return
	}

	gasPrice := tx.GasPrice()
	if receipt != nil && receipt.EffectiveGasPrice != nil {
		gasPrice = receipt.EffectiveGasPrice
	}

	feeEth := "0"
	if gasUsed > 0 {
		feeWei := new(big.Int).Mul(new(big.Int).SetUint64(gasUsed), gasPrice)
		feeEth = utils.WeiToETH(feeWei)
	}

	resp := TxDetailResponse{
		Hash:        tx.Hash().Hex(),
		From:        from.Hex(),
		To:          to,
		ValueEth:    utils.WeiToETH(tx.Value()),
		GasLimit:    tx.Gas(),
		GasUsed:     gasUsed,
		GasPrice:    gasPrice.String(),
		FeeEth:      feeEth,
		Status:      status,
		IsPending:   isPending,
		BlockNumber: blockNumber,
		Logs:        logList,
		Transfers:   transfers,
	}

	writeJSON(w, resp)

}
