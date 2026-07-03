package health

import (
	"context"
)

type ReadinessResult struct {
	Status Status
	Score  int
	Errors map[string]error
}

type Engine struct {
	deps []Dependency
}

func NewEngine(deps []Dependency) *Engine {
	return &Engine{deps: deps}
}

func (e *Engine) CheckReadiness(ctx context.Context) ReadinessResult {
	score := 100
	errorsMap := make(map[string]error)

	for _, dep := range e.deps {
		if err := dep.Check(ctx); err != nil {
			errorsMap[dep.Name()] = err

			if dep.Critical() {
				score = 0
			} else {
				score -= dep.Weight()
			}
		}
	}

	if score < 0 {
		score = 0
	}

	status := e.calculateStatus(score)
	return ReadinessResult{
		Status: status,
		Score:  score,
		Errors: errorsMap,
	}

}

func (e *Engine) calculateStatus(score int) Status {
	switch {
	case score >= 90:
		return StatusHealthy
	case score >= 50:
		return StatusDegraded
	default:
		return StatusUnhealthy
	}
}
