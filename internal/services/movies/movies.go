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

func (s *MovieService) Update(id int, title string, year int32, runtime fields.MovieRuntime, genres []string) (*models.Movie, error) {
	const op = "movies.MovieService.Update"
	log := s.log.With("op", op, "id", id, "title", title, "year", year, "runtime", runtime, "genres", genres)
	movie, err := s.Get(id)
	if err != nil {
		if errors.Is(err, ErrMovieNotFound) {
			return nil, ErrMovieNotFound
		}
		log.Error("Error getting movie: " + err.Error())
		return nil, err
	}
	if title != "" {
		movie.Title = title
	}
	if year != 0 {
		movie.Year = year
	}
	if runtime != 0 {
		movie.Runtime = runtime
	}
	if genres != nil {
		movie.Genres = genres
	}
	updatedMovie, err := s.storage.Update(movie)
	if err != nil {
		if errors.Is(err, storage.ErrConflict) {
			log.Info("movie already exists")
			return nil, ErrMovieAlreadyExists
		} else if errors.Is(err, storage.ErrNotFound) {
			log.Info("movie not found")
			return nil, ErrMovieNotFound
		}
		log.Error("Error updating movie: " + err.Error())
		return nil, err
	}
	return updatedMovie, nil
}