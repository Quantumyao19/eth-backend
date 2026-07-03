package deps

import (
	"context"
	"eth-backend/internal/health"
	"time"

	"github.com/doug-martin/goqu/v9"
)

type PostgresDependency struct {
	gdb *goqu.Database
}

const (
	timeoutForHealthCheck = 300 * time.Millisecond
)

func NewPostgresDependency(gdb *goqu.Database) health.Dependency {
	return &PostgresDependency{gdb: gdb}
}

func (p *PostgresDependency) Name() string {
	return "Postgres"
}

func (p *PostgresDependency) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutForHealthCheck)
	defer cancel()

	var one int
	_, err := p.gdb.Select(goqu.L("1")).ScanValContext(ctx, &one)
	return err
}

func (p *PostgresDependency) Weight() int {
	return 60
}

func (p *PostgresDependency) Critical() bool {
	return true
}
