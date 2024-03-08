// Package database chibe the db :D
package database

import (
	"encoding/json"
	"errors"
	"github.com/kashyab12/chippy/chandler"
	"log"
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
	chibeTheDb := DB{path: path}
	if dbFileErr := chibeTheDb.ensureDB(); dbFileErr != nil {
		return nil, dbFileErr
	}
	return &chibeTheDb, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (chibe *DB) CreateChirp(body string) (Chirp, error) {
}

// loadDB Read chibe into memory
func (chibe *DB) loadDB() (DBStructure, error) {
	chibeTheDb := DBStructure{}
	if jsonFile, openErr := os.OpenFile(chibe.path, os.O_RDWR, 0666); openErr != nil {
		log.Fatalf("Error while trying to open file %v", chibe.path)
		return chibeTheDb, openErr
	} else {
		defer chandler.CloseIoReadCloserStream(jsonFile)
		jsonDecoder := json.NewDecoder(jsonFile)
		if decodeErr := jsonDecoder.Decode(&chibeTheDb); decodeErr != nil {
			log.Fatalf("Error while decoding Chibe the DB :(")
			return chibeTheDb, decodeErr
		}
	}
	return chibeTheDb, nil
}

// writeDB writes the database file to disk
func (chibe *DB) writeDB(dbStructure DBStructure) error {

}

// ensureDB creates a new database file if it doesn't exist
func (chibe *DB) ensureDB() error {
	if _, readError := os.ReadFile(chibe.path); readError != nil {
		if errors.Is(readError, os.ErrNotExist) {
			if chibeFile, writeErr := os.OpenFile(chibe.path, os.O_CREATE|os.O_EXCL, 0666); writeErr != nil {
				return writeErr
			} else {
				defer chandler.CloseIoReadCloserStream(chibeFile)
			}
		} else {
			return readError
		}
	}
	return nil
}

// GetChirps returns all chirps in the database
func (chibe *DB) GetChirps() ([]Chirp, error) {

}
