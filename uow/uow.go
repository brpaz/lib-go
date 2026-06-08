package uow

import "context"

// UnitOfWork represents an atomic unit of database operations.
// T is the module-specific repository bundle available within the transaction.
type UnitOfWork[T any] interface {
	Repositories() T
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// Manager creates units of work. Inject this into services that need
// cross-repository transactions.
type Manager[T any] interface {
	Begin(ctx context.Context) (UnitOfWork[T], error)
}
