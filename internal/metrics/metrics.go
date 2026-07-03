package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	PingCounter prometheus.Counter
}

func NewMetrics() *Metrics {
	return &Metrics{
		PingCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "ping_count_total",
			Help: "Total Number of ping requests",
		}),
	}
}
