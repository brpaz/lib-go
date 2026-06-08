// Package env provides helpers for reading typed values from environment variables.
package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// GetString returns the value of the environment variable named by key,
// or defaultValue if the variable is not set or empty.
func GetString(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value
}

// GetBool returns the boolean value of the environment variable named by key.
// Truthy values: "true", "1". Any other non-empty value returns defaultValue.
func GetBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return value == "true" || value == "1"
}

// GetInt returns the integer value of the environment variable named by key.
// Invalid values are silently ignored and defaultValue is returned instead.
func GetInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// GetFloat returns the float64 value of the environment variable named by key.
// Invalid values are silently ignored and defaultValue is returned instead.
func GetFloat(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}

	return floatValue
}

// GetDuration returns the time.Duration value of the environment variable named by key,
// or defaultValue if the variable is not set, empty, or unparseable.
func GetDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	d, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}

	return d
}

// GetStringSlice returns a slice of strings parsed from a comma-separated
// environment variable. Whitespace around each element is trimmed and empty
// elements are dropped. Returns defaultValue if the variable is not set or empty.
func GetStringSlice(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	return splitAndTrim(value)
}

// MustGetString returns the value of the environment variable named by key.
// It panics if the variable is not set or empty.
func MustGetString(key string) string {
	return mustLookup(key)
}

// MustGetBool returns the boolean value of the environment variable named by key.
// It panics if the variable is not set, empty, or not one of "true", "1", "false", "0".
func MustGetBool(key string) bool {
	value := mustLookup(key)

	switch value {
	case "true", "1":
		return true
	case "false", "0":
		return false
	default:
		panic(fmt.Sprintf("env: environment variable %q has invalid boolean value %q", key, value))
	}
}

// MustGetInt returns the integer value of the environment variable named by key.
// It panics if the variable is not set, empty, or not a valid integer.
func MustGetInt(key string) int {
	value := mustLookup(key)

	intValue, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Sprintf("env: environment variable %q has invalid integer value %q: %v", key, value, err))
	}

	return intValue
}

// MustGetFloat returns the float64 value of the environment variable named by key.
// It panics if the variable is not set, empty, or not a valid float.
func MustGetFloat(key string) float64 {
	value := mustLookup(key)

	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		panic(fmt.Sprintf("env: environment variable %q has invalid float value %q: %v", key, value, err))
	}

	return floatValue
}

// MustGetDuration returns the time.Duration value of the environment variable named by key.
// It panics if the variable is not set, empty, or not a valid duration.
func MustGetDuration(key string) time.Duration {
	value := mustLookup(key)

	d, err := time.ParseDuration(value)
	if err != nil {
		panic(fmt.Sprintf("env: environment variable %q has invalid duration value %q: %v", key, value, err))
	}

	return d
}

// MustGetStringSlice returns a slice of strings parsed from a comma-separated
// environment variable named by key. Whitespace around each element is trimmed and
// empty elements are dropped. It panics if the variable is not set or empty.
func MustGetStringSlice(key string) []string {
	return splitAndTrim(mustLookup(key))
}

// mustLookup returns the raw value of the environment variable named by key.
// It panics if the variable is not set or empty — used by every MustGet* helper
// to enforce "fail fast on missing required config" semantics.
func mustLookup(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("env: required environment variable %q is not set", key))
	}

	return value
}

// splitAndTrim splits a comma-separated string, trims whitespace around each
// element and drops empty elements.
func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
