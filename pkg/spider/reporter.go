package spider

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type MatchResult struct {
	Path      string `json:"path"`
	Reason    string `json:"reason"`
	Hash      string `json:"sha256,omitempty"`
	Size      int64  `json:"size,omitempty"`
	Timestamp string `json:"timestamp"`
	Host      string `json:"host,omitempty"`
	Share     string `json:"share,omitempty"`
}

type Reporter interface {
	Report(MatchResult)
	Close()
}

type JSONReporter struct {
	file *os.File
	enc  *json.Encoder
	mu   sync.Mutex
}

func NewJSONReporter(path string) (*JSONReporter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	return &JSONReporter{
		file: f,
		enc:  json.NewEncoder(f),
	}, nil
}

func (r *JSONReporter) Report(m MatchResult) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if m.Timestamp == "" {
		m.Timestamp = time.Now().Format(time.RFC3339)
	}
	r.enc.Encode(m)
}

func (r *JSONReporter) Close() {
	if r.file != nil {
		r.file.Close()
	}
}

// ConsoleReporter is a dummy reporter that does nothing (logs are handled by utils.Logger)
type ConsoleReporter struct{}

func (c *ConsoleReporter) Report(m MatchResult) {}
func (c *ConsoleReporter) Close()               {}
