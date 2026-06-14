package repository

import (
	"context"
	"eth-backend/internal/model"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
)

const (
	tblTokenTransfers = "token_transfers"
)

type TransferRepository struct {
	gdb *goqu.Database
}

func NewTransferRepository(gdb *goqu.Database) *TransferRepository {
	return &TransferRepository{gdb: gdb}
}

func (r *TransferRepository) Insert(ctx context.Context, t *model.Transfer) error {
	if t == nil {
		return nil
	}
	_, err := r.InsertMany(ctx, []*model.Transfer{t})
	return err
}

func (r *TransferRepository) InsertMany(ctx context.Context, ts []*model.Transfer) (int64, error) {
	if len(ts) == 0 {
		return 0, nil
	}

	records := make([]goqu.Record, 0, len(ts))
	for _, t := range ts {
		if t == nil {
			continue
		}
		records = append(records, goqu.Record{
			"id":            uuid.NewString(),
			"tx_hash":       t.TxHash,
			"log_index":     t.LogIndex,
			"block_number":  t.BlockNumber,
			"token_address": t.TokenAddress,
			"from_address":  t.From,
			"to_address":    t.To,
			"value":         t.Value.String(),
		})
	}

	if len(records) == 0 {
		return 0, nil
	}

	if r.gdb == nil {
		return 0, nil
	}

	res, err := r.gdb.Insert(tblTokenTransfers).Rows(records).OnConflict(goqu.DoNothing()).Executor().ExecContext(ctx)
	if err != nil {
		return 0, err
	}
	if res == nil {
		return 0, nil
	}
	ra, _ := res.RowsAffected()
	return ra, nil
}
