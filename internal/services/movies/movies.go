package movies

import (
	"errors"
	"greenlight/proj/internal/domain/fields"
	"greenlight/proj/internal/domain/models"
	"greenlight/proj/internal/storage"
	"log/slog"
)

type MoviesStorage interface {
	Get(id int) (*models.Movie, error)
	Insert(title string, year int32, runtime fields.MovieRuntime, genres []string) (*models.Movie, error)
	List(limit int) ([]models.Movie, error)
	Update(*models.Movie) (*models.Movie, error)
	Delete(id int) error
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
	movie, err := s.storage.Get(id)
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
	movie, err := s.storage.Insert(title, year, runtime, genres)
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

func (s *MovieService) List(limit int) ([]models.Movie, error) {
	const op = "movies.MovieService.List"
	log := s.log.With("op", op)
	movies, err := s.storage.List(limit)
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
	updatedMovie, err := s.storage.Update(movie)
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
	err := s.storage.Delete(id)
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