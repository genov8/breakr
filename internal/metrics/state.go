package metrics

func (m *Metrics) SetState(state string) {
	if m == nil {
		return
	}

	m.stateGauge.Reset()
	m.stateGauge.
		WithLabelValues(state).
		Set(1)
}

func (m *Metrics) Transition(from, to string) {
	if m == nil || from == to {
		return
	}

	m.transitions.
		WithLabelValues(from, to).
		Inc()
}
