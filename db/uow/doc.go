// Package uow provides a GORM-backed implementation of [libuow.Manager]
// and [libuow.UnitOfWork].
//
// Use [NewManager] to create a transaction manager scoped to a set of
// repositories. Each call to [GormManager.Begin] opens a new database
// transaction and returns the repository bundle constructed by the supplied
// build function.
package uow
