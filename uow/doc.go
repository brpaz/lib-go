// Package uow defines the Unit of Work interfaces for managing cross-repository
// transactions without exposing database implementation details to the service layer.
//
// The concrete implementation lives in internal/infra/db and is wired at the
// application layer. Service code depends only on these interfaces.
//
// # Interfaces
//
// [Manager] creates a new unit of work. [UnitOfWork] holds an active transaction
// and exposes the repository bundle T scoped to it.
//
// T is a plain struct defined by each module grouping the repositories it needs
// inside a single transaction:
//
//	type TxRepos struct {
//	    Users    UsersRepository
//	    Profiles ProfilesRepository
//	}
//
// # Usage
//
//	func (s *Service) CreateWithProfile(ctx context.Context, input Input) error {
//	    tx, err := s.txManager.Begin(ctx)
//	    if err != nil {
//	        return apperror.Internal("begin failed").WithCause(err)
//	    }
//	    defer tx.Rollback(ctx) // no-op after a successful Commit
//
//	    if err := tx.Repositories().Users.Create(ctx, user); err != nil {
//	        return err
//	    }
//	    if err := tx.Repositories().Profiles.Create(ctx, profile); err != nil {
//	        return err
//	    }
//
//	    return tx.Commit(ctx)
//	}
//
// See the inter-repository-transactions guide in docs/guides/ for a full
// wiring example including the app layer and test fakes.
package uow
