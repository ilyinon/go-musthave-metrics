package audit

// Auditor distributes audit events to all configured sinks.
type Auditor struct {
	sinks []Sink
}

// New creates a new Auditor with the provided sinks.
func New(sinks ...Sink) *Auditor {
	return &Auditor{sinks: sinks}
}

// Notify sends an audit event to all sinks asynchronously.
func (a *Auditor) Notify(e Event) {
	for _, s := range a.sinks {
		go func(s Sink) {
			defer func() {
				recover()
			}()
			_ = s.Send(e)
		}(s)
	}
}
