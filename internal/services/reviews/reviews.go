package reviews

import "log/slog"

type ReviewStorage interface {
	// Insert(rating int32, comment string, movieID int64) (int64, error)
}

type ReviewService struct {
	log *slog.Logger
	storage ReviewStorage
}

func New(log *slog.Logger, storage ReviewStorage) *ReviewService {
	return &ReviewService{
		log:     log,
		storage: storage,
	}
}