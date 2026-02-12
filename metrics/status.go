package metrics

type Status string

const (
	StatusSuccess Status = "success"
	StatusError   Status = "error"
	StatusTimeout Status = "timeout"
	StatusBlocked Status = "blocked"
	StatusIgnored Status = "ignored_error"
)
