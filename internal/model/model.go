package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"eth-backend/internal/logger"

	"go.uber.org/zap"
)

type BigInt struct {
	*big.Int
}

func (z *BigInt) Scan(src interface{}) error {
	if z == nil {
		err := fmt.Errorf("BigInt: Scan on nil pointer")
		logger.Log.Error("BigInt scan error", zap.Error(err))
		return err
	}
	if z.Int == nil {
		z.Int = new(big.Int)
	}

	var s string
	switch v := src.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	case int64:
		z.SetInt64(v)
		return nil
	case nil:
		z.Int = nil
		return nil
	default:
		err := fmt.Errorf("BigInt: cannot scan type %T", src)
		logger.Log.Error("BigInt scan error", zap.Error(err), zap.String("src_type", fmt.Sprintf("%T", src)))
		return err
	}

	s = strings.TrimSpace(s)
	if s == "" {
		z.Int = nil
		return nil
	}
	if strings.Contains(s, ".") {
		if strings.HasSuffix(s, ".0") {
			s = strings.TrimSuffix(s, ".0")
			if s == "" {
				s = "0"
			}
		} else {
			err := fmt.Errorf("BigInt: cannot parse decimal %q", s)
			logger.Log.Error("BigInt scan error", zap.Error(err), zap.String("value", s))
			return err
		}
	}
	if _, ok := z.SetString(s, 10); !ok {
		err := fmt.Errorf("BigInt: invalid integer %q", s)
		logger.Log.Error("BigInt scan error", zap.Error(err), zap.String("value", s))
		return err
	}
	return nil
}

func (z BigInt) Value() (driver.Value, error) {
	if z.Int == nil {
		return nil, nil
	}
	return z.String(), nil
}

func (z BigInt) MarshalJSON() ([]byte, error) {
	if z.Int == nil {
		return []byte("null"), nil
	}
	return json.Marshal(z.String())
}

func (z *BigInt) UnmarshalJSON(data []byte) error {
	if z == nil {
		err := fmt.Errorf("BigInt: UnmarshalJSON on nil pointer")
		logger.Log.Error("BigInt JSON error", zap.Error(err))
		return err
	}
	if z.Int == nil {
		z.Int = new(big.Int)
	}
	if string(data) == "null" {
		z.Int = nil
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		return z.Scan(s)
	}

	var n big.Int
	if err := json.Unmarshal(data, &n); err == nil {
		z.Int = &n
		return nil
	}

	err := fmt.Errorf("BigInt: invalid JSON %s", string(data))
	logger.Log.Error("BigInt JSON error", zap.Error(err), zap.String("data", string(data)))
	return err
}

type Transfer struct {
	ID           string    `db:"id"`
	TxHash       string    `db:"tx_hash"`
	LogIndex     uint      `db:"log_index"`
	BlockNumber  uint64    `db:"block_number"`
	TokenAddress string    `db:"token_address"`
	From         string    `db:"from_address"`
	To           string    `db:"to_address"`
	Value        BigInt    `db:"value"`
	CreatedAt    time.Time `db:"created_at"`
}
