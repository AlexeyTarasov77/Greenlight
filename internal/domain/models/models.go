package models

import (
	"greenlight/proj/internal/domain/fields"
	"time"
)

type Movie struct {
	ID        int64               `json:"id"`                // Unique integer ID for the movie
	Title     string              `json:"title"`             // Movie title
	Year      int32               `json:"year,omitempty"`    // Movie release year
	Runtime   fields.MovieRuntime `json:"runtime,omitempty"` // Movie runtime (in minutes)
	Genres    []string            `json:"genres,omitempty"`  // Movie genres (i.e. Comedy, drama, scifi)
	Version   uint                `json:"version"`           // The version number starts at 1 and will be incremented each // time the movie information is updated
	CreatedAt time.Time           `json:"-"`                 // Timestamp for when the movie is added to our database
	Reviews   []Review            `json:"reviews" db:"-"`    // List of reviews
}

var AnonymousUser = &User{}

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash []byte    `json:"-"`
	Email        string    `json:"email"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

type Review struct {
	ID        int64     `json:"id"`
	MovieID   int64     `json:"movie_id"`
	UserID    int64     `json:"user_id"`
	Comment   string    `json:"comment"`
	Rating    int       `json:"rating"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type AuthTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
