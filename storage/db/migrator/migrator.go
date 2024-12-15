package migrator

import (
	"context"
)

// Migrator is an interface for managing database migrations.
type Migrator interface {
	Up(ctx context.Context) error
	Down(ctx context.Context) error
	Reset(ctx context.Context) error
	Create(ctx context.Context, name string) error
}
