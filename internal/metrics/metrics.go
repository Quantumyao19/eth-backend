package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	HTTPRequestsTotal      *prometheus.CounterVec
	HTTPRequestsDuration   *prometheus.HistogramVec
	HTTPRequestsErrors     *prometheus.CounterVec
	HTTPRequestsInProgress prometheus.Gauge

	RPCRequestsTotal   *prometheus.CounterVec
	RPCRequestDuration *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	m := &Metrics{
		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "route", "status"},
		),
		HTTPRequestsDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latency distributions",
				Buckets: []float64{0.01, 0.05, 0.1, 0.3, 0.5, 1, 2, 5},
			},
			[]string{"method", "route"},
		),
		HTTPRequestsErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_error_total",
				Help: "Total number of failed HTTP request",
			},
			[]string{"method", "route", "status_class"},
		),
		HTTPRequestsInProgress: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "go_goroutines",
				Help: "Current number of HTTP requests being processed",
			},
		),
	}
	prometheus.MustRegister(
		m.HTTPRequestsTotal,
		m.HTTPRequestsDuration,
		m.HTTPRequestsErrors,
		m.HTTPRequestsInProgress,
	)

	return m
}
