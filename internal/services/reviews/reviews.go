package reviews

import (
	"context"
	"errors"
	"greenlight/proj/internal/domain/models"
	"greenlight/proj/internal/storage"
	"log/slog"
	"time"
)

type ReviewStorage interface {
	Insert(ctx context.Context, rating int32, comment string, movieID int64, userID int64) (*models.Review, error)
}

type ReviewService struct {
	log     *slog.Logger
	storage ReviewStorage
}

func New(log *slog.Logger, storage ReviewStorage) *ReviewService {
	return &ReviewService{
		log:     log,
		storage: storage,
	}
}

func (s *ReviewService) Create(rating int32, comment string, movieID int64, userID int64) (*models.Review, error) {
	const op = "reviews.ReviewService.Create"
	log := s.log.With("op", op, "rating", rating, "comment", comment, "movieID", movieID, "userID", userID)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	review, err := s.storage.Insert(ctx, rating, comment, movieID, userID)
	if err != nil {
		if errors.Is(err, storage.ErrConflict) {
			log.Info("review already exists")
			return nil, ErrReviewAlreadyExists
		}
		log.Error(err.Error())
		return nil, err
	}
	return review, nil
}
