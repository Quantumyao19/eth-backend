package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	m := &Metrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latency distributions",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "route", "status"},
		),
	}
	prometheus.MustRegister(m.HTTPRequestsTotal)
	prometheus.MustRegister(m.HTTPRequestDuration)
	return m
}
