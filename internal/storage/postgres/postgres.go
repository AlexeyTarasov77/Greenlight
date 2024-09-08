package postgres

import (
	"context"
	"errors"
	"fmt"
	"greenlight/proj/internal/domain/fields"
	"greenlight/proj/internal/domain/filters"
	"greenlight/proj/internal/domain/models"
	"greenlight/proj/internal/services/auth"
	"greenlight/proj/internal/storage"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresDB struct {
	Conn *pgxpool.Pool
}

type pgxTxWrapper struct {
    Tx pgx.Tx
}

func (tx *pgxTxWrapper) Commit(ctx context.Context) error {
    return tx.Tx.Commit(ctx)
}

func (tx *pgxTxWrapper) Rollback(ctx context.Context) error {
    return tx.Tx.Rollback(ctx)
}

func (db *PostgresDB) Begin(ctx context.Context) (auth.Transaction, error) {
	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &pgxTxWrapper{tx}, nil
}

const ErrConflictCode = "23505"

func New(ctx context.Context, storagePath string, maxConns int, maxConnIdleTime time.Duration) (*PostgresDB, error) {
	pool, err := pgxpool.New(ctx, storagePath)
	pool.Config().MaxConns = int32(maxConns)
	pool.Config().MaxConnIdleTime = maxConnIdleTime
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}
	return &PostgresDB{Conn: pool}, nil
}

func (db *PostgresDB) Get(ctx context.Context, id int) (*models.Movie, error) {
	rows, err := db.Conn.Query(
		ctx,
		`SELECT id, title, year, runtime, genres, version, created_at FROM movies WHERE id = $1`,
		id,
	)
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

func (db *PostgresDB) Insert(ctx context.Context, title string, year int32, runtime fields.MovieRuntime, genres []string) (*models.Movie, error) {
	rows, _ := db.Conn.Query(
		ctx,
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

func (db *PostgresDB) List(ctx context.Context, title string, genres []string, filters filters.Filters) ([]models.Movie, int, error) {
	var rows pgx.Rows
	query := fmt.Sprintf(`
	SELECT count(*) OVER(), id, title, year, runtime, genres, version, created_at FROM movies
	WHERE (to_tsvector('english', title) @@ plainto_tsquery('english', $1) OR $1 = '') 
	AND (genres @> $2 OR $2 = '{}')
	ORDER BY %s %s, id ASC
	LIMIT $3 OFFSET $4
	`, filters.SortColumn(), filters.SortDirection())
	args := []any{title, genres, filters.Limit(), filters.Offset()}
	rows, _ = db.Conn.Query(ctx, query, args...)
	type row struct {
		Count int
		models.Movie
	}
	outputRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[row])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, storage.ErrNotFound
		}
		return nil, 0, err
	}
	movies := make([]models.Movie, 0, len(outputRows))
	for _, row := range outputRows {
		movies = append(movies, row.Movie)
	}
	totalRecords := outputRows[0].Count
	return movies, totalRecords, nil
}

func (db *PostgresDB) Update(ctx context.Context, movie *models.Movie) (*models.Movie, error) {
	rows, _ := db.Conn.Query(
		ctx,
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

func (db *PostgresDB) Delete(ctx context.Context, id int) error {
	status, err := db.Conn.Exec(ctx, "DELETE FROM movies WHERE id = $1", id)
	if status.RowsAffected() == 0 {
		return storage.ErrNotFound
	}
	if err != nil {
		return err
	}
	return nil
}