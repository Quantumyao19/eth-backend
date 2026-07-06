package handler

import (
	"context"
	"math/big"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"eth-backend/internal/logger"
	"eth-backend/internal/middleware"
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

	Logs      []LogRaw      `json:"logs"`
	Transfers []TransferLog `json:"transfers"`
	Approvals []ApprovalLog `json:"approvals"`
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

type ApprovalLog struct {
	Token   string `json:"token"`
	Owner   string `json:"owner"`
	Spender string `json:"spender"`
	Value   string `json:"value"`
	Symbol  string `json:"symbol"`
}

const erc20ABI = `[
  {
    "anonymous": false,
    "inputs": [
      {"indexed": true, "name": "from", "type": "address"},
      {"indexed": true, "name": "to", "type": "address"},
      {"indexed": false, "name": "value", "type": "uint256"}
    ],
    "name": "Transfer",
    "type": "event"
  },
  {
    "anonymous": false,
    "inputs": [
      {"indexed": true, "name": "owner", "type": "address"},
      {"indexed": true, "name": "spender", "type": "address"},
      {"indexed": false, "name": "value", "type": "uint256"}
    ],
    "name": "Approval",
    "type": "event"
  }
]`

func (h *Handler) TxDetail(c *gin.Context) {
	requestID, _ := c.Request.Context().Value(middleware.RequestIDKey).(string)

	hash := c.Query("hash")
	if hash == "" {
		writeError(c, "missing hash", http.StatusBadRequest)
		return
	}

	if !common.IsHexHash(hash) {
		writeError(c, "invalid hash format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), defaultTimeout)
	defer cancel()

	// obtain transaction
	tx, isPending, err := h.service.GetTransaction(ctx, common.HexToHash(hash))
	if err != nil {
		handleError(c, err)
		return
	}

	if tx == nil {
		writeError(c, "transaction not found", http.StatusNotFound)
		return
	}

	// obtain receipt
	receipt, err := h.service.GetTransactionReceipt(ctx, common.HexToHash(hash))
	if err != nil {
		handleError(c, err)
		return
	}

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		logger.Log.Error("parsed abi", zap.Error(err), zap.String("request_id", requestID))
		handleError(c, err)
		return
	}

	var logList []LogRaw
	var transfers []TransferLog
	var approvals []ApprovalLog

	tokenMetaCache := make(map[string]struct {
		Symbol   string
		Decimals uint8
	})

	for _, l := range receipt.Logs {
		var topics []string
		for _, t := range l.Topics {
			topics = append(topics, t.Hex())
		}

		logList = append(logList, LogRaw{
			Address: l.Address.Hex(),
			Topics:  topics,
			Data:    "0x" + common.Bytes2Hex(l.Data),
		})

		if len(l.Topics) == 0 {
			continue
		}

		event, err := parsedABI.EventByID(l.Topics[0])
		if err != nil {
			continue
		}

		meta, ok := tokenMetaCache[l.Address.Hex()]
		if !ok {
			symbol := "UNKNOWN"
			decimals := uint8(18)

			s, d, err := h.service.GetTokenMeta(ctx, l.Address)
			if err == nil {
				symbol = s
				decimals = d
			}

			meta = struct {
				Symbol   string
				Decimals uint8
			}{
				symbol, decimals,
			}

			tokenMetaCache[l.Address.Hex()] = meta
		}

		switch event.Name {
		case "Transfer":
			if len(l.Topics) < 3 {
				continue
			}

			var data struct {
				Value *big.Int
			}
			if err := parsedABI.UnpackIntoInterface(&data, "Transfer", l.Data); err != nil {
				continue
			}

			from := common.HexToAddress(l.Topics[1].Hex())
			to := common.HexToAddress(l.Topics[2].Hex())

			transfers = append(transfers, TransferLog{
				Token:  l.Address.Hex(),
				From:   from.Hex(),
				To:     to.Hex(),
				Value:  utils.FormatTokenAmount(data.Value, meta.Decimals),
				Symbol: meta.Symbol,
			})

		case "Approval":
			if len(l.Topics) < 3 {
				continue
			}

			var data struct {
				Value *big.Int
			}

			if err := parsedABI.UnpackIntoInterface(&data, "Approval", l.Data); err != nil {
				continue
			}

			owner := common.HexToAddress(l.Topics[1].Hex())
			spender := common.HexToAddress(l.Topics[2].Hex())

			approvals = append(approvals, ApprovalLog{
				Token:   l.Address.Hex(),
				Owner:   owner.Hex(),
				Spender: spender.Hex(),
				Value:   utils.FormatTokenAmount(data.Value, meta.Decimals),
				Symbol:  meta.Symbol,
			})
		}
	}

	gasUsed := receipt.GasUsed
	status := "success"
	if receipt.Status != 1 {
		status = "failed"
	}

	from, _ := h.service.GetTransactionSender(ctx, tx)

	gasPrice := tx.GasPrice()
	if receipt.EffectiveGasPrice != nil {
		gasPrice = receipt.EffectiveGasPrice
	}

	feeWei := new(big.Int).Mul(new(big.Int).SetUint64(gasUsed), gasPrice)

	resp := TxDetailResponse{
		Hash:        tx.Hash().Hex(),
		From:        from.Hex(),
		To:          tx.To().Hex(),
		ValueEth:    utils.WeiToETH(tx.Value()),
		GasLimit:    tx.Gas(),
		GasUsed:     gasUsed,
		GasPrice:    gasPrice.String(),
		FeeEth:      utils.WeiToETH(feeWei),
		Status:      status,
		IsPending:   isPending,
		BlockNumber: receipt.BlockNumber.String(),
		Logs:        logList,
		Transfers:   transfers,
		Approvals:   approvals,
	}

	writeJSON(c, resp)

}
