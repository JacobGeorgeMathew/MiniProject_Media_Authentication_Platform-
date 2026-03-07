package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JacobGeorgeMathew/MiniProject_Media_Authentication_Platform-/Backend/config"
)

// Connect creates a pgxpool connection pool.
// pgxpool is required by the repository layer (pgvector uses pgx native types).
func Connect(cfg *config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// Verify the connection is live before returning.
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}