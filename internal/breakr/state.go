package breakr

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

func (s State) String() string {
	switch s {
	case Closed:
		return "Closed"
	case Open:
		return "Open"
	case HalfOpen:
		return "Half-Open"
	default:
		return "Unknown"
	}
}

func (b *Breaker) setState(to State) {
	from := b.state
	if from == to {
		return
	}

	b.state = to

	if b.config.Metrics != nil {
		b.config.Metrics.Transition(from.String(), to.String())
		b.config.Metrics.SetState(to.String())
	}
}
