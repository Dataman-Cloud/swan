package janitor

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Prometheus struct {
	RequestCounter  *prometheus.CounterVec
	RequestDuration prometheus.Summary
	RequestSize     prometheus.Summary
	ResponseSize    prometheus.Summary
	BackendDuration prometheus.Summary

	MetricsPath string
}

func (p *Prometheus) registerMetrics(subsystem string) {
	p.RequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "How many HTTP requests processed, partitioned by status code and HTTP method.",
		},
		[]string{"source", "code", "method", "path", "reason", "taskId"},
	)
	prometheus.MustRegister(p.RequestCounter)

	p.RequestDuration = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "The HTTP request latencies in seconds.",
		},
	)
	prometheus.MustRegister(p.RequestDuration)

	p.RequestSize = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Subsystem: subsystem,
			Name:      "request_size_bytes",
			Help:      "The HTTP request sizes in bytes.",
		},
	)
	prometheus.MustRegister(p.RequestSize)

	p.ResponseSize = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Subsystem: subsystem,
			Name:      "response_size_bytes",
			Help:      "The HTTP response sizes in bytes.",
		},
	)
	prometheus.MustRegister(p.ResponseSize)

	p.BackendDuration = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Subsystem: subsystem,
			Name:      "backend_request_duration_seconds",
			Help:      "The duration of backend handle the requst",
		},
	)
	prometheus.MustRegister(p.BackendDuration)
}
