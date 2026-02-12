package metrics

import "time"

func (m *Metrics) ObserveSuccess(state string, d time.Duration) {
	m.requestsTotal.WithLabelValues(string(StatusSuccess), state).Inc()
	m.duration.WithLabelValues(string(StatusSuccess)).Observe(d.Seconds())
}

func (m *Metrics) ObserveError(state string, d time.Duration) {
	m.requestsTotal.WithLabelValues(string(StatusError), state).Inc()
	m.duration.WithLabelValues(string(StatusError)).Observe(d.Seconds())
}

func (m *Metrics) ObserveTimeout(state string, d time.Duration) {
	m.requestsTotal.WithLabelValues(string(StatusTimeout), state).Inc()
	m.duration.WithLabelValues(string(StatusTimeout)).Observe(d.Seconds())
}

func (m *Metrics) ObserveBlocked(state string) {
	m.requestsTotal.WithLabelValues(string(StatusBlocked), state).Inc()
}

func (m *Metrics) ObserveIgnored(state string, d time.Duration) {
	m.requestsTotal.WithLabelValues(string(StatusIgnored), state).Inc()
	m.duration.WithLabelValues(string(StatusIgnored)).Observe(d.Seconds())
}
