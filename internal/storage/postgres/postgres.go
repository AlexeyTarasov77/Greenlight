package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	Conn *pgxpool.Pool
}

const ErrConflictCode = "23505"

func New(ctx context.Context, storagePath string, maxConns int, maxConnIdleTime time.Duration) (*Storage, error) {
	pool, err := pgxpool.New(ctx, storagePath)
	pool.Config().MaxConns = int32(maxConns)
	pool.Config().MaxConnIdleTime = maxConnIdleTime
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &Storage{Conn: pool}, nil
}
