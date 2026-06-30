package health

type Status string

const (
	StatusHealthy   Status = "HEALTHY"
	StatusDegraded  Status = "DEGRADED"
	StatusUnhealthy Status = "UNHEALTHY"
)
