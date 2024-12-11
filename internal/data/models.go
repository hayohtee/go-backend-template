package data

import (
	"database/sql"
	"errors"
)

var (
	// ErrRecordNotFound is a custom error that is returned when
	// looking for a record in the database that doesn't exist.
	ErrRecordNotFound = errors.New("record not found")

	// ErrEditConflict is a custom error that is returned when
	// there is an edit conflict in the database operation.
	ErrEditConflict = errors.New("edit conflict")
)

// Models is a struct that wraps around all database models
// for the entire application.
type Models struct {
}

// NewModels return a Models containing initialized models for
// the entire application.
func NewModels(db *sql.DB) Models {
	return Models{}
}
