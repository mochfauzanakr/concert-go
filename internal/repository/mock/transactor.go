package mock

import (
	"context"
	"sync"

	"concert-go/internal/repository"
)

type mockTxKey struct{}

type mockTx struct {
	mu    sync.Mutex
	undos []func()
}

func (tx *mockTx) onRollback(fn func()) {
	tx.mu.Lock()
	defer tx.mu.Unlock()
	tx.undos = append(tx.undos, fn)
}

func (tx *mockTx) rollback() {
	tx.mu.Lock()
	undos := tx.undos
	tx.undos = nil
	tx.mu.Unlock()

	// Execute undo callbacks in reverse order outside tx.mu to avoid lock-order deadlocks.
	for i := len(undos) - 1; i >= 0; i-- {
		undos[i]()
	}
}

func registerRollback(ctx context.Context, fn func()) {
	if tx, ok := ctx.Value(mockTxKey{}).(*mockTx); ok && tx != nil {
		tx.onRollback(fn)
	}
}

type transactorMock struct{}

// NewTransactorMock returns an in-memory Transactor that rolls back mock mutations on error or panic.
func NewTransactorMock() repository.Transactor {
	return &transactorMock{}
}

func (t *transactorMock) WithinTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	if existing, ok := ctx.Value(mockTxKey{}).(*mockTx); ok && existing != nil {
		return fn(ctx)
	}

	tx := &mockTx{}
	txCtx := context.WithValue(ctx, mockTxKey{}, tx)
	committed := false
	defer func() {
		if !committed {
			tx.rollback()
		}
	}()

	if err := fn(txCtx); err != nil {
		return err
	}

	committed = true
	return nil
}
