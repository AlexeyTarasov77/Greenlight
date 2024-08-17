package postgres

import (
	"context"
	"errors"
	"greenlight/proj/internal/domain/fields"
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

const ErrConflictCode = "23505"

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

func (db *PostgresDB) Insert(title string, year int32, runtime fields.MovieRuntime, genres []string) (*models.Movie, error) {
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
		if errors.As(err, &pgxErr) && pgxErr.Code == ErrConflictCode {
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

func (db *PostgresDB) Update(movie *models.Movie) (*models.Movie, error) {
	rows, _ := db.Conn.Query(
		context.Background(),
		`UPDATE movies SET version = version + 1, title = $1, year = $2, runtime = $3, genres = $4
		WHERE id = $5 AND version = $6 RETURNING *`,
		movie.Title,
		movie.Year,
		movie.Runtime,
		movie.Genres,
		movie.ID,
		movie.Version,
	)
	m, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.Movie])
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) && pgxErr.Code == ErrConflictCode {
			return nil, storage.ErrConflict
		} else if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (db *PostgresDB) Delete(id int) error {
	status, err := db.Conn.Exec(context.Background(), "DELETE FROM movies WHERE id = $1", id)
	if status.RowsAffected() == 0 {
		return storage.ErrNotFound
	}
	if err != nil {
		return err
	}
	return nil
}