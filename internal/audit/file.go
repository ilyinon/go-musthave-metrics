package audit

import (
	"encoding/json"
	"os"
	"sync"
)

// FileSink writes audit events to a file.
type FileSink struct {
	file *os.File
	mu   sync.Mutex
}

// NewFileSink creates a new FileSink.
func NewFileSink(path string) (*FileSink, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &FileSink{file: f}, nil
}

func (f *FileSink) Send(e Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	b, err := json.Marshal(e)
	if err != nil {
		return err
	}

	_, err = f.file.Write(append(b, '\n'))
	return err
}
