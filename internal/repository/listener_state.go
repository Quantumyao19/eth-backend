package repository

import (
	"context"
	"eth-backend/internal/logger"

	"github.com/doug-martin/goqu/v9"
	"go.uber.org/zap"
)

const (
	tblListenerState = "listener_state"
	transferListener = "transfer_listener"
)

type ListenerStateRepository struct {
	gdb *goqu.Database
}

func NewListenerStateRepository(gdb *goqu.Database) *ListenerStateRepository {
	return &ListenerStateRepository{
		gdb: gdb,
	}
}

func (r *ListenerStateRepository) EnsureState(ctx context.Context, name string) error {
	_, err := r.gdb.Insert(tblListenerState).Rows(goqu.Record{
		"name": name,
	}).OnConflict(goqu.DoNothing()).Executor().ExecContext(ctx)

	return err
}

func (r *ListenerStateRepository) GetLastProcessedBlock(ctx context.Context, name string) (int64, error) {
	var block int64

	found, err := r.gdb.From(tblListenerState).Select("last_processed_block").Where(goqu.C("name").Eq(name)).Executor().ScanValContext(ctx, &block)

	if err != nil {
		logger.Log.Error("failed to query listener state", zap.Error(err), zap.String("name", name))
		return 0, err
	}

	if !found {
		logger.Log.Warn("listener state not found", zap.String("name", name))
		return 0, nil
	}

	return block, nil
}

func (r *ListenerStateRepository) UpdateLastProcessedBlockTx(ctx context.Context, tx *goqu.TxDatabase, name string, block int64) error {
	_, err := r.gdb.Update(tblListenerState).Set(
		goqu.Record{"last_processed_block": block,
			"updated_at": goqu.L("NOW()"),
		}).Where(goqu.C("name").Eq(name)).Executor().ExecContext(ctx)

	if err != nil {
		logger.Log.Error("failed to update listener state", zap.Error(err), zap.String("name", name), zap.Int64("block", block))
		return err
	}

	return nil
}
