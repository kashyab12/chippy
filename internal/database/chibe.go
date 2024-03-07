// Package database chibe the db :D
package database

import (
	"errors"
	"os"
	"sync"
)

type DB struct {
	path string
	mux  sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

type Chirp struct {
	Uid  int    `json:"id"`
	Body string `json:"body"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	_, readError := os.ReadFile(path)
	if readError != nil {
		if errors.Is(readError, os.ErrNotExist) {
			if writeErr := os.WriteFile(path, []byte(""), 666); writeErr != nil {
				return nil, writeErr
			}
		} else {
			return nil, readError
		}
	}
	return &DB{
		path: path,
	}, nil
}
