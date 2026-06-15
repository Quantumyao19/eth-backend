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

func (r *TransferRepository) ListByAddress(ctx context.Context, address string, page int, pageSize int) ([]model.Transfer, uint64, error) {
	offset := (page - 1) * pageSize

	var result []model.Transfer

	whereCondition := goqu.Or(
		goqu.Ex{"from_address": address},
		goqu.Ex{"to_address": address},
		goqu.Ex{"token_address": address},
	)

	err := r.gdb.From(tblTokenTransfers).Where(whereCondition).
		Order(goqu.I("block_number").Desc(), goqu.I("log_index").Desc()).
		Limit(uint(pageSize)).
		Offset(uint(offset)).Executor().ScanStructsContext(ctx, &result)
	if err != nil {
		return nil, 0, err
	}

	var total uint64
	_, err = r.gdb.Select(goqu.COUNT("*")).From(tblTokenTransfers).Where(whereCondition).Executor().ScanValContext(ctx, &total)
	if err != nil {
		return nil, 0, err
	}

	return result, total, nil
}
