package env_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/brpaz/lib-go/env"
)

func TestGetString(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		setEnv       bool
		defaultValue string
		want         string
	}{
		{
			name:         "var set returns value",
			envKey:       "TEST_STRENV_SET",
			envValue:     "hello",
			setEnv:       true,
			defaultValue: "default",
			want:         "hello",
		},
		{
			name:         "var not set returns default",
			envKey:       "TEST_STRENV_UNSET",
			setEnv:       false,
			defaultValue: "default",
			want:         "default",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setEnv {
				t.Setenv(tc.envKey, tc.envValue)
			}
			got := env.GetString(tc.envKey, tc.defaultValue)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGetBool(t *testing.T) {
	const key = "TEST_BOOLENV"

	tests := []struct {
		name         string
		envValue     string
		setEnv       bool
		defaultValue bool
		want         bool
	}{
		{name: `"true" is truthy`, envValue: "true", setEnv: true, defaultValue: false, want: true},
		{name: `"1" is truthy`, envValue: "1", setEnv: true, defaultValue: false, want: true},
		{
			name:         `"false" returns false`,
			envValue:     "false",
			setEnv:       true,
			defaultValue: false,
			want:         false,
		},
		{name: `"0" returns false`, envValue: "0", setEnv: true, defaultValue: false, want: false},
		{
			name:         `"yes" is not truthy, returns defaultValue false`,
			envValue:     "yes",
			setEnv:       true,
			defaultValue: false,
			want:         false,
		},
		{
			name:         `"yes" is not truthy, returns defaultValue true`,
			envValue:     "yes",
			setEnv:       true,
			defaultValue: true,
			want:         false,
		},
		{name: "not set returns default false", setEnv: false, defaultValue: false, want: false},
		{name: "not set returns default true", setEnv: false, defaultValue: true, want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setEnv {
				t.Setenv(key, tc.envValue)
			}
			got := env.GetBool(key, tc.defaultValue)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGetInt(t *testing.T) {
	const key = "TEST_INTENV"

	tests := []struct {
		name         string
		envValue     string
		setEnv       bool
		defaultValue int
		want         int
	}{
		{name: "valid int", envValue: "42", setEnv: true, defaultValue: 0, want: 42},
		{
			name:         "invalid string returns default",
			envValue:     "abc",
			setEnv:       true,
			defaultValue: 99,
			want:         99,
		},
		{name: "not set returns default", setEnv: false, defaultValue: 10, want: 10},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setEnv {
				t.Setenv(key, tc.envValue)
			}
			got := env.GetInt(key, tc.defaultValue)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGetFloat(t *testing.T) {
	const key = "TEST_FLOATENV"

	tests := []struct {
		name         string
		envValue     string
		setEnv       bool
		defaultValue float64
		want         float64
	}{
		{name: "valid float", envValue: "3.14", setEnv: true, defaultValue: 0, want: 3.14},
		{name: "valid integer-looking value", envValue: "42", setEnv: true, defaultValue: 0, want: 42},
		{
			name:         "invalid string returns default",
			envValue:     "abc",
			setEnv:       true,
			defaultValue: 1.5,
			want:         1.5,
		},
		{name: "not set returns default", setEnv: false, defaultValue: 2.5, want: 2.5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setEnv {
				t.Setenv(key, tc.envValue)
			}
			got := env.GetFloat(key, tc.defaultValue)
			assert.InDelta(t, tc.want, got, 0.0001)
		})
	}
}

func TestGetDuration(t *testing.T) {
	const key = "TEST_DURENV"

	tests := []struct {
		name         string
		envValue     string
		setEnv       bool
		defaultValue time.Duration
		want         time.Duration
	}{
		{
			name:         `valid "30m" parsed correctly`,
			envValue:     "30m",
			setEnv:       true,
			defaultValue: 0,
			want:         30 * time.Minute,
		},
		{
			name:         "invalid string returns default",
			envValue:     "notaduration",
			setEnv:       true,
			defaultValue: 5 * time.Second,
			want:         5 * time.Second,
		},
		{name: "not set returns default", setEnv: false, defaultValue: time.Hour, want: time.Hour},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setEnv {
				t.Setenv(key, tc.envValue)
			}
			got := env.GetDuration(key, tc.defaultValue)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGetStringSlice(t *testing.T) {
	const key = "TEST_SLICEENV"

	tests := []struct {
		name         string
		envValue     string
		setEnv       bool
		defaultValue []string
		want         []string
	}{
		{
			name:         "comma-separated values are trimmed",
			envValue:     "a, b , c",
			setEnv:       true,
			defaultValue: nil,
			want:         []string{"a", "b", "c"},
		},
		{
			name:         "single value returns single-element slice",
			envValue:     "only",
			setEnv:       true,
			defaultValue: nil,
			want:         []string{"only"},
		},
		{
			name:         "empty entries between commas are dropped",
			envValue:     "a,,b, ,c",
			setEnv:       true,
			defaultValue: nil,
			want:         []string{"a", "b", "c"},
		},
		{
			name:         "not set returns default",
			setEnv:       false,
			defaultValue: []string{"x", "y"},
			want:         []string{"x", "y"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setEnv {
				t.Setenv(key, tc.envValue)
			}
			got := env.GetStringSlice(key, tc.defaultValue)
			require.Len(t, got, len(tc.want))
			assert.ElementsMatch(t, tc.want, got)
		})
	}
}

func TestMustGetString(t *testing.T) {
	const key = "TEST_MUST_STRENV"

	t.Run("var set returns value", func(t *testing.T) {
		t.Setenv(key, "hello")
		assert.Equal(t, "hello", env.MustGetString(key))
	})

	t.Run("var not set panics", func(t *testing.T) {
		assert.Panics(t, func() { env.MustGetString("TEST_MUST_STRENV_UNSET") })
	})

	t.Run("var empty panics", func(t *testing.T) {
		t.Setenv(key, "")
		assert.Panics(t, func() { env.MustGetString(key) })
	})
}

func TestMustGetBool(t *testing.T) {
	const key = "TEST_MUST_BOOLENV"

	tests := []struct {
		name      string
		envValue  string
		wantPanic bool
		want      bool
	}{
		{name: `"true" returns true`, envValue: "true", want: true},
		{name: `"1" returns true`, envValue: "1", want: true},
		{name: `"false" returns false`, envValue: "false", want: false},
		{name: `"0" returns false`, envValue: "0", want: false},
		{name: `"yes" panics`, envValue: "yes", wantPanic: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(key, tc.envValue)

			if tc.wantPanic {
				assert.Panics(t, func() { env.MustGetBool(key) })
				return
			}

			assert.Equal(t, tc.want, env.MustGetBool(key))
		})
	}

	t.Run("var not set panics", func(t *testing.T) {
		assert.Panics(t, func() { env.MustGetBool("TEST_MUST_BOOLENV_UNSET") })
	})
}

func TestMustGetInt(t *testing.T) {
	const key = "TEST_MUST_INTENV"

	t.Run("valid int", func(t *testing.T) {
		t.Setenv(key, "42")
		assert.Equal(t, 42, env.MustGetInt(key))
	})

	t.Run("invalid string panics", func(t *testing.T) {
		t.Setenv(key, "abc")
		assert.Panics(t, func() { env.MustGetInt(key) })
	})

	t.Run("var not set panics", func(t *testing.T) {
		assert.Panics(t, func() { env.MustGetInt("TEST_MUST_INTENV_UNSET") })
	})
}

func TestMustGetFloat(t *testing.T) {
	const key = "TEST_MUST_FLOATENV"

	t.Run("valid float", func(t *testing.T) {
		t.Setenv(key, "3.14")
		assert.InDelta(t, 3.14, env.MustGetFloat(key), 0.0001)
	})

	t.Run("invalid string panics", func(t *testing.T) {
		t.Setenv(key, "abc")
		assert.Panics(t, func() { env.MustGetFloat(key) })
	})

	t.Run("var not set panics", func(t *testing.T) {
		assert.Panics(t, func() { env.MustGetFloat("TEST_MUST_FLOATENV_UNSET") })
	})
}

func TestMustGetDuration(t *testing.T) {
	const key = "TEST_MUST_DURENV"

	t.Run(`valid "30m" parsed correctly`, func(t *testing.T) {
		t.Setenv(key, "30m")
		assert.Equal(t, 30*time.Minute, env.MustGetDuration(key))
	})

	t.Run("invalid string panics", func(t *testing.T) {
		t.Setenv(key, "notaduration")
		assert.Panics(t, func() { env.MustGetDuration(key) })
	})

	t.Run("var not set panics", func(t *testing.T) {
		assert.Panics(t, func() { env.MustGetDuration("TEST_MUST_DURENV_UNSET") })
	})
}

func TestMustGetStringSlice(t *testing.T) {
	const key = "TEST_MUST_SLICEENV"

	t.Run("comma-separated values are trimmed", func(t *testing.T) {
		t.Setenv(key, "a, b , c")
		got := env.MustGetStringSlice(key)
		require.Len(t, got, 3)
		assert.ElementsMatch(t, []string{"a", "b", "c"}, got)
	})

	t.Run("var not set panics", func(t *testing.T) {
		assert.Panics(t, func() { env.MustGetStringSlice("TEST_MUST_SLICEENV_UNSET") })
	})

	t.Run("var empty panics", func(t *testing.T) {
		t.Setenv(key, "")
		assert.Panics(t, func() { env.MustGetStringSlice(key) })
	})
}
