package postgres

import (
	"database/sql"
	"concert-go/internal/repository"
)

type helloRepositoryImpl struct {
	db *sql.DB
}

// NewHelloRepository implements HelloRepository using pure SQL
func NewHelloRepository(db *sql.DB) repository.HelloRepository {
	return &helloRepositoryImpl{
		db: db,
	}
}

func (r *helloRepositoryImpl) PingDatabase() error {
	// Example of pure SQL query
	err := r.db.Ping()
	return err
}
