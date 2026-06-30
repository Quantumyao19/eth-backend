package health

import (
	"context"
	"sync"
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

	results := make(chan checkResult, 3)
	var wg sync.WaitGroup
	run := func(name string, fn func(context.Context) error) {
		defer wg.Done()

		err := fn(ctx)
		results <- checkResult{
			name: name,
			err:  err,
		}
	}
	wg.Add(3)

	go run("db", c.checkDB)
	go run("redis", c.checkRedis)
	go run("rpc", c.checkRPC)

	wg.Wait()
	close(results)

	errorsMap := make(map[string]error)

	for r := range results {
		if r.err != nil {
			errorsMap[r.name] = r.err
		}
	}

	status := StatusHealthy

	if _, ok := errorsMap["db"]; ok {
		status = StatusUnhealthy
	} else if len(errorsMap) > 0 {
		status = StatusDegraded
	}

	return ReadinessResult{
		Status: status,
		Errors: errorsMap,
	}

}
