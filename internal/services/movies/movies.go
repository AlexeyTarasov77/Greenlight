package movies

import (
	"context"
	"errors"
	"greenlight/proj/internal/domain/fields"
	"greenlight/proj/internal/domain/filters"
	"greenlight/proj/internal/domain/models"
	"greenlight/proj/internal/storage"
	"log/slog"
	"reflect"
	"time"
)

type MoviesStorage interface {
	Get(ctx context.Context, id int) (*models.Movie, error)
	Insert(ctx context.Context, title string, year int32, runtime fields.MovieRuntime, genres []string) (*models.Movie, error)
	List(ctx context.Context, limit int, title string, genres []string, filters filters.Filters) ([]models.Movie, error)
	Update(ctx context.Context, movie *models.Movie) (*models.Movie, error)
	Delete(ctx context.Context, id int) error
}

type MovieService struct {
	log     *slog.Logger
	storage MoviesStorage
}

func New(log *slog.Logger, storage MoviesStorage) *MovieService {
	return &MovieService{
		log:     log,
		storage: storage,
	}
}

func (s *MovieService) Get(id int) (*models.Movie, error) {
	const op = "movies.MovieService.Get"
	log := s.log.With("op", op, "id", id)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	movie, err := s.storage.Get(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Info("movie not found")
			return nil, ErrMovieNotFound
		}
		log.Error(err.Error())
		return nil, err
	}
	return movie, nil
}

func (s *MovieService) Create(title string, year int32, runtime fields.MovieRuntime, genres []string) (*models.Movie, error) {
	const op = "movies.MovieService.Create"
	log := s.log.With("op", op, "title", title, "year", year, "runtime", runtime, "genres", genres)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	movie, err := s.storage.Insert(ctx, title, year, runtime, genres)
	if err != nil {
		if errors.Is(err, storage.ErrConflict) {
			log.Info("movie already exists")
			return nil, ErrMovieAlreadyExists
		}
		log.Error(err.Error())
		return nil, err
	}
	return movie, nil
}

func (s *MovieService) List(limit int, title string, genres []string, page int, pageSize int, sort string) ([]models.Movie, error) {
	const op = "movies.MovieService.List"
	log := s.log.With("op", op)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	movieFieldsNum := reflect.TypeOf(models.Movie{}).NumField()
	movieFields := make([]string, 0, movieFieldsNum)
	for i := 0; i < movieFieldsNum; i++ {
		movieFields = append(movieFields, reflect.TypeOf(models.Movie{}).Field(i).Name)
	}
	filters := filters.Filters{
		Page: page,
		PageSize: pageSize,
		Sort: sort,
		SortSafelist: movieFields,
	}
	movies, err := s.storage.List(ctx, limit, title, genres, filters)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Info("movies not found")
			return nil, ErrMovieNotFound
		}
		log.Error(err.Error())
		return nil, err
	}
	return movies, nil
}

func (s *MovieService) Update(id int, title *string, year *int32, runtime *fields.MovieRuntime, genres []string) (*models.Movie, error) {
	const op = "movies.MovieService.Update"
	log := s.log.With("op", op, "id", id, "title", title, "year", year, "runtime", runtime, "genres", genres)
	movie, err := s.Get(id)
	if err != nil {
		if errors.Is(err, ErrMovieNotFound) {
			log.Info("movie not found")
			return nil, ErrMovieNotFound
		}
		log.Error("Error getting movie: " + err.Error())
		return nil, err
	}
	var argsChanged int
	if title != nil {
		movie.Title = *title
		argsChanged++
	}
	if year != nil {
		movie.Year = *year
		argsChanged++
	}
	if runtime != nil {
		movie.Runtime = *runtime
		argsChanged++
	}
	if genres != nil {
		movie.Genres = genres
		argsChanged++
	}
	if argsChanged == 0 {
		log.Info("no arguments changed")
		return nil, ErrNoArgumentsChanged
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	updatedMovie, err := s.storage.Update(ctx, movie)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrConflict):
			log.Info("movie already exists")
			return nil, ErrMovieAlreadyExists
		case errors.Is(err, storage.ErrNotFound):
			log.Warn("Update conflict, because of concurrent update")
			return nil, ErrEditConflict
		default:
			log.Error("Error updating movie: " + err.Error())
			return nil, err
		}
	}
	return updatedMovie, nil
}

func (s *MovieService) Delete(id int) error {
	const op = "movies.MovieService.Delete"
	log := s.log.With("op", op, "id", id)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	err := s.storage.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			log.Info("movie not found")
			return ErrMovieNotFound
		}
		log.Error(err.Error())
		return err
	}
	return nil
}