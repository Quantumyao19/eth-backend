package health

import (
	"context"
)

type ReadinessResult struct {
	Status Status
	Errors map[string]error
}

type Engine struct {
	deps []Dependency
}

func NewEngine(deps []Dependency) *Engine {
	return &Engine{deps: deps}
}

func (e *Engine) CheckReadiness(ctx context.Context) ReadinessResult {
	errorsMap := make(map[string]error)

	status := StatusHealthy
	for _, dep := range e.deps {
		if err := dep.Check(ctx); err != nil {
			errorsMap[dep.Name()] = err

			if dep.Critical() {
				status = StatusUnhealthy
			} else {
				if status != StatusUnhealthy {
					status = StatusDegraded
				}
			}
		}
	}

	return ReadinessResult{
		Status: status,
		Errors: errorsMap,
	}

}
