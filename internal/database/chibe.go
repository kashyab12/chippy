// Package database chibe the db :D
package database

import (
	"errors"
	"github.com/kashyab12/chippy/chandler"
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
			if chibeFile, writeErr := os.OpenFile(path, os.O_CREATE|os.O_EXCL, 0666); writeErr != nil {
				return nil, writeErr
			} else {
				defer chandler.CloseIoReadCloserStream(chibeFile)
			}
		} else {
			return nil, readError
		}
	}
	return &DB{
		path: path,
	}, nil
}
