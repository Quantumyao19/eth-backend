package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	// HTTP
	HTTPRequestsTotal      *prometheus.CounterVec
	HTTPRequestsDuration   *prometheus.HistogramVec
	HTTPRequestsErrors     *prometheus.CounterVec
	HTTPRequestsInProgress prometheus.Gauge

	// RPC
	RPCRequestsTotal   *prometheus.CounterVec
	RPCRequestDuration *prometheus.HistogramVec
	RPCRequestsErrors  *prometheus.CounterVec

	// Listener
	ListenerBlocksProcessedTotal prometheus.Counter
	ListenerEventsProcessedTotal prometheus.Counter
	ListenerCycleDurationSeconds prometheus.Histogram
	ListenerLastProcessedBlock   prometheus.Gauge
	ListenerBlockLag             prometheus.Gauge
	ListenerErrorsTotal          *prometheus.CounterVec
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
				Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30},
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

		ListenerBlocksProcessedTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "listener_blocks_processed_total",
				Help: "Total number of pocessed blocks",
			},
		),
		ListenerEventsProcessedTotal: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "listener_events_processed_total",
				Help: "Total number of processed events",
			},
		),
		ListenerCycleDurationSeconds: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "listener_cycle_duration_seconds",
				Help:    "Duration of a listener synchronization cycle in seconds",
				Buckets: []float64{0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10},
			},
		),
		ListenerLastProcessedBlock: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "listener_last_processed_block",
				Help: "Last processed Ethereum block number",
			},
		),
		ListenerBlockLag: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "listener_block_lag",
				Help: "Number of blocks listener is behind the chain head",
			},
		),
		ListenerErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "listener_errors_total",
				Help: "Total number of listener errors",
			},
			[]string{"stage"},
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

		m.ListenerBlocksProcessedTotal,
		m.ListenerEventsProcessedTotal,
		m.ListenerCycleDurationSeconds,
		m.ListenerLastProcessedBlock,
		m.ListenerBlockLag,
		m.ListenerErrorsTotal,
	)

	return m
}
