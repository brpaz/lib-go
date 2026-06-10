package migrator

import (
	"context"
	"fmt"
	"os"

	"github.com/brpaz/lib-go/logging"
)

// Logger is the logging interface consumed by [Migrator].
// It mirrors the goose.Logger contract so any adapter that satisfies goose
// also satisfies this interface.
type Logger interface {
	Printf(format string, v ...any)
	Fatalf(format string, v ...any)
}

// NewLibLogger returns a Logger backed by a [logging.Logger].
// Printf maps to Info; Fatalf maps to Error followed by os.Exit(1).
func NewLibLogger(l *logging.Logger) Logger {
	return &libLogger{l: l}
}

// NewNopLogger returns a Logger that discards all output.
func NewNopLogger() Logger {
	return nopLogger{}
}

type libLogger struct{ l *logging.Logger }

func (s *libLogger) Printf(format string, v ...any) {
	s.l.Info(context.Background(), fmt.Sprintf(format, v...))
}

func (s *libLogger) Fatalf(format string, v ...any) {
	s.l.Error(context.Background(), fmt.Sprintf(format, v...))
	os.Exit(1)
}

type nopLogger struct{}

func (nopLogger) Printf(string, ...any) {}
func (nopLogger) Fatalf(string, ...any) {}
