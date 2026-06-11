package handler

import (
	"context"
	"log"
	"math/big"
	"net/http"

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
}

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
	}

	writeJSON(w, resp)

}
