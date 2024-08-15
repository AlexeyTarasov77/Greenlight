package postgres

import (
	"context"
	"errors"
	"greenlight/proj/internal/domain/models"
	"greenlight/proj/internal/storage"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDB struct {
	Conn *pgxpool.Pool
}

func New(storagePath string, maxConns int, maxConnIdleTime time.Duration) (*PostgresDB, error) {
	pool, err := pgxpool.New(context.Background(), storagePath)
	pool.Config().MaxConns = int32(maxConns)
	pool.Config().MaxConnIdleTime = maxConnIdleTime
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}
	return &PostgresDB{Conn: pool}, nil
}

func (db *PostgresDB) Get(id int) (*models.Movie, error) {
	rows, err := db.Conn.Query(context.Background(), "SELECT * FROM movies WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	movie, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.Movie])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return &movie, nil
}

func (db *PostgresDB) Insert(title string, year int32, runtime int32, genres []string) (*models.Movie, error) {
	rows, _ := db.Conn.Query(
		context.Background(),
		"INSERT INTO movies (title, year, runtime, genres) VALUES ($1, $2, $3, $4) RETURNING *",
		title,
		year,
		runtime,
		genres,
	)
	movie, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.Movie])
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) && pgxErr.Code == "23505" {
			return nil, storage.ErrConflict
		}
		return nil, err
	}
	return &movie, nil
}

func (db *PostgresDB) List(limit int) ([]models.Movie, error) {
	var rows pgx.Rows
	if limit == storage.EmptyIntValue {
		rows, _ = db.Conn.Query(context.Background(), "SELECT * FROM movies")
	} else {
		rows, _ = db.Conn.Query(context.Background(), "SELECT * FROM movies LIMIT $1", limit)
	}
	movies, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Movie])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return movies, nil
}
