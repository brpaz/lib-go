// Package log provides a structured logging interface for Go applications.
//
// It is designed to be simple to use and easy to configure. The package provides
// a generic Logger interface that can be implemented by various logging adapters.
// Built-in adapters include Zap (high-performance production logging), InMemory
// (log capturing for testing), and Noop (no-operation logging).
//
// Key Features:
//
// - Adapters:
//   - Zap: A high-performance, production-ready logging adapter.
//   - InMemory: Suitable for testing or in-memory log capturing.
//   - **Noop**: A no-operation adapter, ideal for disabling logging.
//
// - Profiles:
//   - ProfileDevelopment: Configures the logger for development environments, with
//     human-readable log formats and more verbose output.
//   - ProfileProduction: Optimized for production, using structured and minimal logs.
//
// - Formats:
//   - json: Logs are output in structured JSON format for log aggregators.
//   - fmt: Logs are output in a human-readable format for development.
//
// Example Usage:
//
//	package main
//
//	import (
//		"context"
//		"errors"
//		"github.com/brpaz/lib-go/log"
//	)
//
//	func main() {
//	    logger, err := log.New(
//	        log.WithAdapter("zap"),
//	        log.WithLevel("debug"),
//	        log.WithProfile(log.ProfileProduction),
//	        log.WithFormat(log.FormatJSON),
//	    )
//	    if err != nil {
//	        panic(errors.Join(err, errors.New("failed to create logger")))
//	    }
//	    defer logger.Sync()
//
//	    // Use the logger
//	    logger.Info(context.Background(), "Application started", log.String("env", "production"))
//	}
package log
