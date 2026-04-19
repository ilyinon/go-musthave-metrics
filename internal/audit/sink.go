package audit

// Sink represents a destination for audit events.
type Sink interface {
	Send(Event) error
}
