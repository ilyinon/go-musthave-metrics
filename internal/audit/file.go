package audit

import (
	"encoding/json"
	"os"
	"sync"
)

type FileSink struct {
	path string
	mu   sync.Mutex
}

func NewFileSink(path string) *FileSink {
	return &FileSink{path: path}
}

func (f *FileSink) Send(e Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	b, err := json.Marshal(e)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(f.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(append(b, '\n'))
	return err
}
