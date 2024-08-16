package models

import (
	"greenlight/proj/internal/domain/fields"
	"time"
)


type Movie struct {
	ID        int      `json:"id"`         // Unique integer ID for the movie
	Title     string   `json:"title"`      // Movie title
	Year      int32    `json:"year,omitempty"`       // Movie release year
	Runtime   fields.MovieRuntime   `json:"runtime,omitempty"`    // Movie runtime (in minutes)
	Genres    []string `json:"genres,omitempty"`	   // Movie genres (i.e. Comedy, drama, scifi)
	Version   uint     `json:"version"` // The version number starts at 1 and will be incremented each // time the movie information is updated
	CreatedAt time.Time   `json:"-"` // Timestamp for when the movie is added to our database
}
