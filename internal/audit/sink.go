package audit

type Sink interface {
	Send(Event) error
}
