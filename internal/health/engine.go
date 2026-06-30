package health

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

type ReadinessResult struct {
	Status Status
	Errors map[string]error
}

func (c *Checker) CheckReadiness(ctx context.Context) ReadinessResult {
	type checkResult struct {
		name string
		err  error
	}
	g, ctx := errgroup.WithContext(ctx)

	var mu sync.Mutex
	errorsMap := make(map[string]error)

	g.Go(func() error {
		if err := c.checkDB(ctx); err != nil {
			mu.Lock()
			errorsMap["db"] = err
			mu.Unlock()
		}
		return nil
	})

	g.Go(func() error {
		if err := c.checkRedis(ctx); err != nil {
			mu.Lock()
			errorsMap["redis"] = err
			mu.Unlock()
		}
		return nil
	})
	_ = g.Wait()

	status := StatusHealthy
	if len(errorsMap) > 0 {
		status = StatusDegraded
	}
	if _, ok := errorsMap["db"]; ok {
		status = StatusUnhealthy
	}

	return ReadinessResult{
		Status: status,
		Errors: errorsMap,
	}

}
