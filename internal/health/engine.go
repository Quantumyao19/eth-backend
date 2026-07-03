package health

import (
	"context"
	"time"
)

type DependencyResult struct {
	Name     string
	Status   Status
	Duration time.Duration
	Error    error
}

type ReadinessResult struct {
	Status Status
	Score  int

	Dependencies []DependencyResult
}

type Engine struct {
	deps []Dependency
}

func NewEngine(deps []Dependency) *Engine {
	return &Engine{deps: deps}
}

func (e *Engine) CheckReadiness(ctx context.Context) ReadinessResult {
	score := 100
	dependencies := make([]DependencyResult, 0)

	for _, dep := range e.deps {
		start := time.Now()
		err := dep.Check(ctx)
		dependency := DependencyResult{
			Name:     dep.Name(),
			Status:   StatusHealthy,
			Duration: time.Since(start),
			Error:    err,
		}
		if err != nil {
			dependency.Status = StatusUnhealthy

			if dep.Critical() {
				score = 0
			} else {
				score -= dep.Weight()
			}
		}
		dependencies = append(dependencies, dependency)
	}

	if score < 0 {
		score = 0
	}

	status := e.calculateStatus(score)
	return ReadinessResult{
		Status:       status,
		Score:        score,
		Dependencies: dependencies,
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
