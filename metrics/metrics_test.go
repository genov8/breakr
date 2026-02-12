package metrics

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func newTestMetrics(t *testing.T) *Metrics {
	t.Helper()

	reg := prometheus.NewRegistry()

	m := &Metrics{
		requestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "requests_total",
			},
			[]string{"status", "state"},
		),
		duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "execution_duration_seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"status"},
		),
		stateGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "state",
			},
			[]string{"state"},
		),
		transitions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "state_transitions_total",
			},
			[]string{"from", "to"},
		),
	}

	reg.MustRegister(
		m.requestsTotal,
		m.duration,
		m.stateGauge,
		m.transitions,
	)

	return m
}

func TestObserveSuccess(t *testing.T) {
	m := newTestMetrics(t)

	m.ObserveSuccess("Closed", 10*time.Millisecond)

	if v := testutil.ToFloat64(
		m.requestsTotal.WithLabelValues("success", "Closed"),
	); v != 1 {
		t.Fatalf("expected success counter = 1, got %v", v)
	}
}

func TestObserveError(t *testing.T) {
	m := newTestMetrics(t)

	m.ObserveError("Closed", 5*time.Millisecond)

	if v := testutil.ToFloat64(
		m.requestsTotal.WithLabelValues("error", "Closed"),
	); v != 1 {
		t.Fatalf("expected error counter = 1, got %v", v)
	}
}

func TestObserveBlocked(t *testing.T) {
	m := newTestMetrics(t)

	m.ObserveBlocked("Open")

	if v := testutil.ToFloat64(
		m.requestsTotal.WithLabelValues("blocked", "Open"),
	); v != 1 {
		t.Fatalf("expected blocked counter = 1, got %v", v)
	}
}

func TestObserveIgnored(t *testing.T) {
	m := newTestMetrics(t)

	m.ObserveIgnored("Closed", time.Millisecond)

	if v := testutil.ToFloat64(
		m.requestsTotal.WithLabelValues("ignored_error", "Closed"),
	); v != 1 {
		t.Fatalf("expected ignored_error counter = 1, got %v", v)
	}
}

func TestSetState(t *testing.T) {
	m := newTestMetrics(t)

	m.SetState("Open")

	if v := testutil.ToFloat64(
		m.stateGauge.WithLabelValues("Open"),
	); v != 1 {
		t.Fatalf("expected state gauge = 1, got %v", v)
	}
}

func TestTransition(t *testing.T) {
	m := newTestMetrics(t)

	m.Transition("Closed", "Open")

	if v := testutil.ToFloat64(
		m.transitions.WithLabelValues("Closed", "Open"),
	); v != 1 {
		t.Fatalf("expected transition counter = 1, got %v", v)
	}
}

func TestTransitionSameStateIgnored(t *testing.T) {
	m := newTestMetrics(t)

	m.Transition("Closed", "Closed")

	if v := testutil.ToFloat64(
		m.transitions.WithLabelValues("Closed", "Closed"),
	); v != 0 {
		t.Fatalf("expected transition counter = 0, got %v", v)
	}
}
