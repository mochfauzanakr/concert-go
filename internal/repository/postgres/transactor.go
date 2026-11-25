package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"concert-go/internal/repository"
)

// DBTX abstracts query and execution operations shared by *sql.DB and *sql.Tx.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type txContextKey struct{}

func execer(ctx context.Context, db *sql.DB) DBTX {
	if tx, ok := ctx.Value(txContextKey{}).(*sql.Tx); ok && tx != nil {
		return tx
	}
	return db
}

type transactorImpl struct {
	db *sql.DB
}

// NewTransactor returns a PostgreSQL Transactor backed by *sql.DB.
func NewTransactor(db *sql.DB) repository.Transactor {
	return &transactorImpl{db: db}
}

func (t *transactorImpl) WithinTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	if existing, ok := ctx.Value(txContextKey{}).(*sql.Tx); ok && existing != nil {
		return fn(ctx)
	}

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	txCtx := context.WithValue(ctx, txContextKey{}, tx)
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	if err := fn(txCtx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	committed = true
	return nil
}
