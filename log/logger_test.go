package log_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/log"
)

func TestLoggerOpts_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    log.LoggerOpts
		wantErr bool
	}{
		{
			name: "Valid Development Profile, JSON Format, Zap Adapter",
			opts: log.LoggerOpts{
				Profile: log.ProfileDevelopment,
				Format:  log.FormatJSON,
				Adapter: log.AdapterZap,
			},
			wantErr: false,
		},
		{
			name: "Valid Production Profile, LogFmt Format, Nop Adapter",
			opts: log.LoggerOpts{
				Profile: log.ProfileProduction,
				Format:  log.FormatLogFmt,
				Adapter: log.AdapterNop,
			},
			wantErr: false,
		},
		{
			name: "Invalid Profile",
			opts: log.LoggerOpts{
				Profile: "invalid", // Invalid profile
				Format:  log.FormatJSON,
				Adapter: log.AdapterZap,
			},
			wantErr: true,
		},
		{
			name: "Invalid Format",
			opts: log.LoggerOpts{
				Profile: log.ProfileDevelopment,
				Format:  "invalid", // Invalid format
				Adapter: log.AdapterZap,
			},
			wantErr: true,
		},
		{
			name: "Invalid Adapter",
			opts: log.LoggerOpts{
				Profile: log.ProfileDevelopment,
				Format:  log.FormatJSON,
				Adapter: "invalid", // Invalid adapter
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.opts.Validate()
			if tt.wantErr {
				assert.Error(t, err, "Expected error but got none")
			} else {
				assert.NoError(t, err, "Expected no error but got one")
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	t.Parallel()

	t.Run("NoopLogger", func(t *testing.T) {
		t.Parallel()
		logger, err := log.New(
			log.WithAdapter(log.AdapterNop),
			log.WithFormat(log.FormatJSON),
			log.WithProfile(log.ProfileDevelopment),
		)
		require.NoError(t, err)
		assert.IsType(t, &log.NopLogger{}, logger)
	})

	t.Run("InMemoryLogger", func(t *testing.T) {
		t.Parallel()
		logger, err := log.New(
			log.WithAdapter(log.AdapterInMemory),
			log.WithLevel(log.LevelWarn),
		)
		require.NoError(t, err)
		assert.IsType(t, &log.InMemoryLogger{}, logger)
		assert.Equal(t, log.LevelWarn, logger.GetLevel())
	})

	t.Run("ZapLogger", func(t *testing.T) {
		t.Parallel()
		logger, err := log.New(
			log.WithAdapter(log.AdapterZap),
			log.WithLevel(log.LevelWarn),
		)
		require.NoError(t, err)
		assert.IsType(t, &log.ZapAdapter{}, logger)
		assert.Equal(t, log.LevelWarn, logger.GetLevel())
	})

	t.Run("InvalidAdapter", func(t *testing.T) {
		t.Parallel()
		_, err := log.New(
			log.WithAdapter("invalid"),
			log.WithLevel(log.LevelWarn),
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create logger: invalid log adapter")
	})
}
