package postgres

import (
	"context"
	"covalence/src/db/postgres/sqlc"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store handles database operations with a simplified interface
type DB struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
	Mu      sync.Mutex
}

// New creates a database store with connection pooling
func New(ctx context.Context, connString string) (*DB, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	return &DB{
		Pool:    pool,
		Queries: sqlc.New(pool),
	}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	db.Pool.Close()
}
