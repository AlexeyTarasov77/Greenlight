package models

import (
	"context"
	"errors"
	"greenlight/proj/internal/domain/models"
	"greenlight/proj/internal/storage"
	"greenlight/proj/internal/storage/postgres"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReviewModel struct {
	DB *pgxpool.Pool
}

func (m *ReviewModel) Insert(ctx context.Context, rating int32, comment string, movieID int64, userID int64) (*models.Review, error) {
	rows, _ := m.DB.Query(ctx, "INSERT INTO reviews (rating, comment, movie_id, user_id) VALUES ($1, $2, $3, $4) RETURNING *", rating, comment, movieID, userID)
	review, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[models.Review])
	if err != nil {
		var pgxErr *pgconn.PgError
		if errors.As(err, &pgxErr) && pgxErr.Code == postgres.ErrConflictCode {
			return nil, storage.ErrConflict
		}
		return nil, err
	}
	return &review, nil
}

func (m *ReviewModel) GetForMovie(ctx context.Context, movieID int64) ([]models.Review, error) {
	rows, _ := m.DB.Query(ctx, "SELECT * FROM reviews WHERE movie_id = $1", movieID)
	reviews, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Review])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	return reviews, nil
}