package audit

type Auditor struct {
	sinks []Sink
}

func New(sinks ...Sink) *Auditor {
	return &Auditor{sinks: sinks}
}

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
