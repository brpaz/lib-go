package gormlog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	gormlogger "gorm.io/gorm/logger"

	"github.com/brpaz/lib-go/db/gormlog"
	"github.com/brpaz/lib-go/logging"
)

func newGormLogger(
	t *testing.T,
	buf *bytes.Buffer,
	level slog.Level,
	opts ...gormlog.Option,
) *gormlog.Logger {
	t.Helper()

	l, err := logging.NewLogger(
		logging.WithOutput(buf),
		logging.WithFormat(logging.FormatJSON),
		logging.WithLevel(level.String()),
	)
	require.NoError(t, err)

	return gormlog.New(l, opts...)
}

func parseLogEntry(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var entry map[string]any
	require.NotEmpty(t, buf.Bytes(), "expected a log entry but buffer is empty")
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry))
	return entry
}

func TestNewLogger(t *testing.T) {
	t.Parallel()

	t.Run("slow query logged at Warn by default", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		l := newGormLogger(t, &buf, slog.LevelDebug)

		begin := time.Now().Add(-300 * time.Millisecond)
		l.Trace(context.Background(), begin, func() (string, int64) { return "SELECT 1", 1 }, nil)

		assert.Equal(t, "WARN", parseLogEntry(t, &buf)["level"])
	})

	t.Run("normal query logged at Info when ignoreTrace=false", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		l := newGormLogger(t, &buf, slog.LevelDebug)

		begin := time.Now().Add(-10 * time.Millisecond)
		l.Trace(context.Background(), begin, func() (string, int64) { return "SELECT 1", 1 }, nil)

		assert.Equal(t, "INFO", parseLogEntry(t, &buf)["level"])
	})
}

func TestLogger_LogMode(t *testing.T) {
	t.Parallel()

	t.Run("returns a different value", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		original := newGormLogger(t, &buf, slog.LevelDebug)
		cloned := original.LogMode(gormlogger.Info)
		assert.NotSame(t, original, cloned)
	})

	t.Run("Silent clone suppresses output", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		base := newGormLogger(t, &buf, slog.LevelDebug)
		base.LogMode(gormlogger.Silent).Info(context.Background(), "should not appear")
		assert.Empty(t, buf.Bytes())
	})

	t.Run("original unchanged after LogMode", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		base := newGormLogger(t, &buf, slog.LevelDebug)
		_ = base.LogMode(gormlogger.Silent)
		base.Info(context.Background(), "hello")
		assert.NotEmpty(t, buf.Bytes())
	})
}

func TestLogger_Info(t *testing.T) {
	t.Parallel()

	t.Run("logged at Info level", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		l := newGormLogger(t, &buf, slog.LevelInfo)
		l.Info(context.Background(), "hello %s", "world")

		entry := parseLogEntry(t, &buf)
		assert.Equal(t, "INFO", entry["level"])
		assert.Equal(t, "hello world", entry["message"])
	})

	t.Run("suppressed when level is Warn", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		newGormLogger(t, &buf, slog.LevelWarn).Info(context.Background(), "no")
		assert.Empty(t, buf.Bytes())
	})
}

func TestLogger_Warn(t *testing.T) {
	t.Parallel()

	t.Run("logged at Warn level", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		l := newGormLogger(t, &buf, slog.LevelWarn)
		l.Warn(context.Background(), "watch out %s", "!")

		entry := parseLogEntry(t, &buf)
		assert.Equal(t, "WARN", entry["level"])
		assert.Equal(t, "watch out !", entry["message"])
	})

	t.Run("suppressed when level is Error", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		newGormLogger(t, &buf, slog.LevelError).Warn(context.Background(), "no")
		assert.Empty(t, buf.Bytes())
	})
}

func TestLogger_Error(t *testing.T) {
	t.Parallel()

	t.Run("logged at Error level", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		l := newGormLogger(t, &buf, slog.LevelError)
		l.Error(context.Background(), "boom %d", 1)

		entry := parseLogEntry(t, &buf)
		assert.Equal(t, "ERROR", entry["level"])
		assert.Equal(t, "boom 1", entry["message"])
	})

	t.Run("suppressed when Silent", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		base := newGormLogger(t, &buf, slog.LevelDebug)
		base.LogMode(gormlogger.Silent).Error(context.Background(), "no")
		assert.Empty(t, buf.Bytes())
	})
}

func TestLogger_Trace(t *testing.T) {
	t.Parallel()

	const testSQL = "SELECT * FROM users"

	cases := []struct {
		name          string
		level         slog.Level
		silent        bool
		ignoreTrace   bool
		slowThreshold time.Duration
		elapsed       time.Duration
		err           error
		wantEmpty     bool
		wantLevel     string
		wantFields    map[string]bool
	}{
		{
			name:      "Silent – nothing logged",
			silent:    true,
			level:     slog.LevelDebug,
			elapsed:   10 * time.Millisecond,
			wantEmpty: true,
		},
		{
			name:       "query error logged at Error with error field",
			level:      slog.LevelError,
			elapsed:    10 * time.Millisecond,
			err:        errors.New("connection refused"),
			wantLevel:  "ERROR",
			wantFields: map[string]bool{"elapsed": true, "rows": true, "sql": true, "error": true},
		},
		{
			name:        "ErrRecordNotFound + ignoreTrace – nothing logged",
			level:       slog.LevelError,
			elapsed:     10 * time.Millisecond,
			err:         gorm.ErrRecordNotFound,
			ignoreTrace: true,
			wantEmpty:   true,
		},
		{
			name:          "ErrRecordNotFound treated as normal query – logged at Info",
			level:         slog.LevelInfo,
			slowThreshold: 200 * time.Millisecond,
			elapsed:       10 * time.Millisecond,
			err:           gorm.ErrRecordNotFound,
			wantLevel:     "INFO",
			wantFields:    map[string]bool{"elapsed": true, "rows": true, "sql": true},
		},
		{
			name:          "slow query logged at Warn with threshold field",
			level:         slog.LevelWarn,
			slowThreshold: 50 * time.Millisecond,
			elapsed:       100 * time.Millisecond,
			wantLevel:     "WARN",
			wantFields: map[string]bool{
				"elapsed":   true,
				"rows":      true,
				"sql":       true,
				"threshold": true,
			},
		},
		{
			name:          "slowThreshold=0 disables slow-query check, falls through to Info",
			level:         slog.LevelInfo,
			slowThreshold: 0,
			elapsed:       500 * time.Millisecond,
			wantLevel:     "INFO",
			wantFields:    map[string]bool{"elapsed": true, "rows": true, "sql": true},
		},
		{
			name:          "normal query ignoreTrace=false – logged at Info",
			level:         slog.LevelInfo,
			slowThreshold: 200 * time.Millisecond,
			elapsed:       10 * time.Millisecond,
			wantLevel:     "INFO",
			wantFields:    map[string]bool{"elapsed": true, "rows": true, "sql": true},
		},
		{
			name:          "normal query ignoreTrace=true – nothing logged",
			level:         slog.LevelInfo,
			slowThreshold: 200 * time.Millisecond,
			elapsed:       10 * time.Millisecond,
			ignoreTrace:   true,
			wantEmpty:     true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			var opts []gormlog.Option
			if tc.slowThreshold != 200*time.Millisecond {
				opts = append(opts, gormlog.WithSlowThreshold(tc.slowThreshold))
			}
			if tc.ignoreTrace {
				opts = append(opts, gormlog.WithIgnoreTrace())
			}

			l := newGormLogger(t, &buf, tc.level, opts...)
			var iface gormlogger.Interface = l
			if tc.silent {
				iface = l.LogMode(gormlogger.Silent)
			}

			iface.Trace(context.Background(), time.Now().Add(-tc.elapsed),
				func() (string, int64) { return testSQL, 5 }, tc.err)

			if tc.wantEmpty {
				assert.Empty(t, buf.Bytes())
				return
			}

			entry := parseLogEntry(t, &buf)
			assert.Equal(t, tc.wantLevel, entry["level"])
			assert.Equal(t, testSQL, entry["sql"])
			for field := range tc.wantFields {
				assert.Contains(t, entry, field)
			}
			if tc.err != nil && !errors.Is(tc.err, gorm.ErrRecordNotFound) {
				assert.Equal(t, tc.err.Error(), entry["error"])
			}
		})
	}
}
