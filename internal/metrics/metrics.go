package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type Metrics struct {
	HTTPRequestsTotal      *prometheus.CounterVec
	HTTPRequestsDuration   *prometheus.HistogramVec
	HTTPRequestsErrors     *prometheus.CounterVec
	HTTPRequestsInProgress prometheus.Gauge

	RPCRequestsTotal   *prometheus.CounterVec
	RPCRequestDuration *prometheus.HistogramVec
	RPCRequestsErrors  *prometheus.CounterVec
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
				Help:    "HTTP requests latency distributions",
				Buckets: []float64{0.01, 0.05, 0.1, 0.3, 0.5, 1, 2, 5},
			},
			[]string{"method", "route"},
		),
		HTTPRequestsErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_errors_total",
				Help: "Total number of failed HTTP requests",
			},
			[]string{"method", "route", "status_class"},
		),
		HTTPRequestsInProgress: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_progress",
				Help: "Current number of HTTP requests being processed",
			},
		),

		RPCRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rpc_requests_total",
				Help: "Total number of RPC requests",
			},
			[]string{"method"},
		),
		RPCRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "rpc_request_duration_seconds",
				Help:    "RPC requests latency distributions",
				Buckets: []float64{0.01, 0.05, 0.1, 0.3, 0.5, 1, 2, 5},
			},
			[]string{"method"},
		),
		RPCRequestsErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rpc_requests_errors_total",
				Help: "Total number of failed RPC requests",
			},
			[]string{"method"},
		),
	}
	prometheus.MustRegister(
		m.HTTPRequestsTotal,
		m.HTTPRequestsDuration,
		m.HTTPRequestsErrors,
		m.HTTPRequestsInProgress,

		m.RPCRequestsTotal,
		m.RPCRequestDuration,
		m.RPCRequestsErrors,

		collectors.NewGoCollector(),
	)

	return m
}
