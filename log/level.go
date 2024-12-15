package log

import (
	"errors"
	"strings"
)

// Level defines all available log levels for log messages.
type Level int

// Log levels supported by the application.
const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// levelNames maps Level values to their string representations.
var levelNames = []string{
	"debug",
	"info",
	"warn",
	"error",
}

// levelLookup provides a case-insensitive lookup for log levels by name.
var levelLookup = func() map[string]Level {
	lookup := make(map[string]Level, len(levelNames))
	for i, name := range levelNames {
		lookup[strings.ToLower(name)] = Level(i)
	}
	return lookup
}()

// ErrInvalidLevel is returned when an invalid log level is provided.
var ErrInvalidLevel = errors.New("logger: invalid log level")

// String returns the string representation of a logging level.
// If the level is invalid, it returns "unknown".
func (p Level) String() string {
	if p < LevelDebug || p > LevelError {
		return "unknown"
	}
	return levelNames[p]
}

// LevelFromString returns the log level from a string representation.
// It performs a case-insensitive comparison and returns an error for invalid levels.
func LevelFromString(level string) (Level, error) {
	if lvl, ok := levelLookup[strings.ToLower(level)]; ok {
		return lvl, nil
	}
	return LevelInfo, ErrInvalidLevel
}
