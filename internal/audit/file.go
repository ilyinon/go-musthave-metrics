package audit

import (
	"encoding/json"
	"os"
	"sync"
)

type FileSink struct {
	file *os.File
	mu   sync.Mutex
}

func NewFileSink(path string) *FileSink {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	return &FileSink{file: f}
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
