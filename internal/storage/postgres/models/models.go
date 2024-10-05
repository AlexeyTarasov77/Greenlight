package models

import "greenlight/proj/internal/storage/postgres"

type Models struct {
	Movie  *MovieModel
	Review *ReviewModel
}

func New(db *postgres.Storage) *Models {
	return &Models{
		Movie:  &MovieModel{db.Conn},
		Review: &ReviewModel{db.Conn},
	}
}
