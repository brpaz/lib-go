package uow

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	libuow "github.com/brpaz/lib-go/uow"
)

// GormManager is the GORM implementation of [libuow.Manager].
// Build is called once per [Begin] with the transaction-scoped *gorm.DB;
// it should construct all repository instances for that transaction.
type GormManager[T any] struct {
	db    *gorm.DB
	build func(*gorm.DB) T
}

// NewManager creates a GormManager. build receives a transaction *gorm.DB on
// each Begin call and must return the module's repository bundle.
func NewManager[T any](db *gorm.DB, build func(*gorm.DB) T) (*GormManager[T], error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}

	return &GormManager[T]{db: db, build: build}, nil
}

func (m *GormManager[T]) Begin(ctx context.Context) (libuow.UnitOfWork[T], error) {
	tx := m.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("begin transaction: %w", tx.Error)
	}

	return &gormUoW[T]{tx: tx, repos: m.build(tx)}, nil
}

type gormUoW[T any] struct {
	tx    *gorm.DB
	repos T
}

func (u *gormUoW[T]) Repositories() T {
	return u.repos
}

func (u *gormUoW[T]) Commit(ctx context.Context) error {
	return u.tx.WithContext(ctx).Commit().Error
}

func (u *gormUoW[T]) Rollback(ctx context.Context) error {
	u.tx.WithContext(ctx).Rollback()
	return nil
}
