package health

import (
	"context"
	"time"
)

const (
	timeoutForHealthCheck = 300 * time.Millisecond
)

type Checker struct {
	engine *Engine
}

func NewChecker(engine *Engine) *Checker {
	return &Checker{
		engine: engine,
	}
}

func (c *Checker) CheckReadiness(ctx context.Context) ReadinessResult {
	return c.engine.CheckReadiness(ctx)
}
