package events

import (
	"encoding/json"
	"io"
	"sync"
	"time"
)

type Writer struct {
	mu  sync.Mutex
	enc *json.Encoder
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{enc: json.NewEncoder(w)}
}

func (w *Writer) Write(ev Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if ev.Time.IsZero() {
		ev.Time = time.Now().UTC()
	}
	return w.enc.Encode(ev)
}
