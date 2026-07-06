package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestsDuration *prometheus.HistogramVec
	HTTPRequestsErrors   *prometheus.CounterVec
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
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "route", "status"},
		),
		HTTPRequestsErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_error_total",
				Help: "Total number of failed HTTP request",
			},
			[]string{"method", "route", "status"},
		),
	}
	prometheus.MustRegister(m.HTTPRequestsTotal)
	prometheus.MustRegister(m.HTTPRequestsDuration)
	prometheus.MustRegister(m.HTTPRequestsErrors)

	return m
}
