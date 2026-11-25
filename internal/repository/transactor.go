package repository

import "context"

// Transactor executes a callback inside a single database transaction propagated via context.
// Repositories inspect ctx for an active transaction and fall back to the default DB handle otherwise.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(txCtx context.Context) error) error
}
