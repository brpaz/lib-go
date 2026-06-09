// Package logtest provides test helpers for the log package.
//
// Use [NewRecorder] to capture structured log output and assert on individual
// fields without timestamp noise or JSON parsing boilerplate:
//
//	rec := logtest.NewRecorder(t)
//	logger, _ := log.NewLogger(log.WithOutput(rec), log.WithFormat(log.FormatJSON))
//
//	logger.Info(ctx, "user created", slog.String("user_id", "u-1"))
//
//	entry := rec.Last()
//	assert.Equal(t, "user created", entry["message"])
//	assert.Equal(t, "u-1", entry["user_id"])
package logtest

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"

	"github.com/brpaz/lib-go/logging"
)


// Recorder captures log output written by a [log.Logger] configured with
// [log.FormatJSON] and [log.WithOutput](rec). Each written record is parsed
// into a map with the "timestamp" key stripped for stable assertions.
type Recorder struct {
	t  *testing.T
	mu sync.Mutex
	rs []map[string]any
}

// NewRecorder returns a Recorder ready for use in t.
func NewRecorder(t *testing.T) *Recorder {
	t.Helper()
	return &Recorder{t: t}
}

// Write implements [io.Writer]. Each call is expected to contain exactly one
// complete JSON log record (the slog JSON handler guarantees this).
func (r *Recorder) Write(p []byte) (int, error) {
	line := strings.TrimSpace(string(p))
	if line == "" {
		return len(p), nil
	}

	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		r.t.Errorf("logtest: failed to parse log line as JSON: %v\nline: %s", err, line)
		return len(p), nil
	}
	delete(m, "timestamp")

	r.mu.Lock()
	r.rs = append(r.rs, m)
	r.mu.Unlock()

	return len(p), nil
}

// All returns a copy of every captured record in order.
func (r *Recorder) All() []map[string]any {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]map[string]any, len(r.rs))
	copy(out, r.rs)
	return out
}

// Last returns the most recent captured record.
// Calls t.Fatal if no records have been captured yet.
func (r *Recorder) Last() map[string]any {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.rs) == 0 {
		r.t.Fatal("logtest: Last called but no records were captured")
		return nil
	}
	return r.rs[len(r.rs)-1]
}

// Len returns the number of captured records.
func (r *Recorder) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.rs)
}

// Reset discards all captured records.
func (r *Recorder) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rs = r.rs[:0]
}

// New returns a [logging.Logger] and a [Recorder] pre-configured for testing.
// The logger writes JSON at DEBUG level so all records are captured.
func New(t *testing.T) (*logging.Logger, *Recorder) {
	t.Helper()
	rec := NewRecorder(t)
	l, err := logging.NewLogger(
		logging.WithOutput(rec),
		logging.WithFormat(logging.FormatJSON),
		logging.WithLevel("debug"),
	)
	if err != nil {
		t.Fatalf("logtest: failed to create logger: %v", err)
	}
	return l, rec
}
