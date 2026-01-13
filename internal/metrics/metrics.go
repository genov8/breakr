package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	requestsTotal *prometheus.CounterVec
	duration      *prometheus.HistogramVec
	stateGauge    *prometheus.GaugeVec
	transitions   *prometheus.CounterVec
}

func NewMetrics(subsystem string) *Metrics {
	m := &Metrics{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      "requests_total",
				Help:      "Total number of requests through the circuit breaker",
			},
			[]string{"status", "state"},
		),

		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: subsystem,
				Name:      "execution_duration_seconds",
				Help:      "Execution duration of requests",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"status"},
		),

		stateGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: subsystem,
				Name:      "state",
				Help:      "Current state of the circuit breaker",
			},
			[]string{"state"},
		),

		transitions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Subsystem: subsystem,
				Name:      "state_transitions_total",
				Help:      "Total number of circuit breaker state transitions",
			},
			[]string{"from", "to"},
		),
	}

	prometheus.MustRegister(
		m.requestsTotal,
		m.duration,
		m.stateGauge,
		m.transitions,
	)

	return m
}
