package repository

import (
	"context"

	"github.com/doug-martin/goqu/v9"
)

type CleanupRepository struct {
	gdb *goqu.Database
}

func NewCleanupRepository(gdb *goqu.Database) *CleanupRepository {
	return &CleanupRepository{
		gdb: gdb,
	}
}

func (r *CleanupRepository) DeleteOldTransfers(ctx context.Context, retainRecords int) error {
	subQuery := r.gdb.From("token_transfers").Select("id").Order(goqu.I("id").Desc()).Offset(uint(retainRecords)).Limit(1)

	query := r.gdb.Delete("token_transfers").Where(goqu.I("id").Lt(subQuery))

	_, err := query.Executor().ExecContext(ctx)

	return err
}
