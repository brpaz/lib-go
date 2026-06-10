// Package db provides driver-agnostic database connectivity helpers built on GORM.
//
// # Connection
//
// [NewConnection] opens a GORM connection using a caller-provided
// [gorm.io/gorm.Dialector] (e.g. [gorm.io/driver/postgres.Open]), keeping this
// package independent of any specific database driver. Use functional options
// to configure the connection pool ([WithMaxIdleConns], [WithMaxOpenConns],
// [WithConnMaxLifetime]), enable structured query logging ([WithQueryLog]) via
// a [github.com/brpaz/lib-go/logging.Logger] ([WithLogger]), Prometheus metrics
// ([WithMetrics]), and OpenTelemetry tracing ([WithTracing]).
//
// # Subpackages
//
//   - migrator: schema migration runner backed by goose.
//   - uow: GORM-backed unit-of-work and transaction manager.
//   - gormlog: GORM logger adapters (logging.Logger, nop).
//   - dbtest: test helpers for spinning up ephemeral PostgreSQL instances.
package db
