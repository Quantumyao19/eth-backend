package model

import "math/big"

type Transfer struct {
	TxHash       string
	LogIndex     uint
	BlockNumber  uint64
	TokenAddress string
	From         string
	To           string
	Value        *big.Int
}
