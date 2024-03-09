// Package database chibe the db :D
package database

import (
	"cmp"
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"slices"
	"sync"
)

const ChibeFile = "./database.json"

type DB struct {
	path string
	mux  sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

type Chirp struct {
	Uid  int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	Uid   int    `json:"id"`
	Email string `json:"email"`
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
	var newChirp Chirp
	if chirps, getChirpsErr := chibe.GetChirps(); getChirpsErr != nil {
		return newChirp, getChirpsErr
	} else {
		newChirpId := 1
		if len(chirps) > 0 {
			highestUid := chirps[len(chirps)-1].Uid
			newChirpId = highestUid + 1
		}
		if dbStruct, loadErr := chibe.loadDB(); loadErr != nil {
			return newChirp, loadErr
		} else {
			newChirp = Chirp{
				Uid:  newChirpId,
				Body: body,
			}
			dbStruct.Chirps[newChirp.Uid] = newChirp
			if writeErr := chibe.writeDB(dbStruct); writeErr != nil {
				return newChirp, writeErr
			}
		}
	}
	return newChirp, nil
}

// loadDB Read chibe into memory
func (chibe *DB) loadDB() (DBStructure, error) {
	chibeTheDb := DBStructure{}
	if jsonFile, openErr := os.OpenFile(chibe.path, os.O_RDWR, 0666); openErr != nil {
		log.Fatalf("Error while trying to open file %v", chibe.path)
		return chibeTheDb, openErr
	} else {
		defer closeDbFile(jsonFile)
		jsonDecoder := json.NewDecoder(jsonFile)
		if decodeErr := jsonDecoder.Decode(&chibeTheDb); decodeErr != nil {
			// Case when chibe is empty
			if errors.Is(decodeErr, io.EOF) {
				return DBStructure{map[int]Chirp{}}, nil
			} else {
				log.Fatalf("Error while decoding Chibe the DB :(")
				return chibeTheDb, decodeErr
			}
		}
	}
	return chibeTheDb, nil
}

// writeDB writes the database file to disk
func (chibe *DB) writeDB(dbStructure DBStructure) error {
	chibe.mux.Lock()
	defer chibe.mux.Unlock()
	if rawData, encodeErr := json.Marshal(dbStructure); encodeErr != nil {
		return encodeErr
	} else {
		if writeErr := os.WriteFile(chibe.path, rawData, 0666); writeErr != nil {
			return writeErr
		}
	}
	return nil
}

// ensureDB creates a new database file if it doesn't exist
func (chibe *DB) ensureDB() error {
	if _, readError := os.ReadFile(chibe.path); readError != nil {
		if errors.Is(readError, os.ErrNotExist) {
			if chibeFile, writeErr := os.OpenFile(chibe.path, os.O_CREATE|os.O_EXCL, 0666); writeErr != nil {
				return writeErr
			} else {
				defer closeDbFile(chibeFile)
			}
		} else {
			return readError
		}
	}
	return nil
}

// GetChirps returns all chirps in the database
func (chibe *DB) GetChirps() ([]Chirp, error) {
	var chirps []Chirp
	if dbStruct, loadErr := chibe.loadDB(); loadErr != nil {
		return nil, loadErr
	} else if len(dbStruct.Chirps) > 0 {
		chibe.mux.RLock()
		defer chibe.mux.RUnlock()
		for _, chirp := range dbStruct.Chirps {
			chirps = append(chirps, chirp)
		}
		slices.SortFunc(chirps, func(a, b Chirp) int { return cmp.Compare(a.Uid, b.Uid) })
	}
	return chirps, nil
}

func (chibe *DB) GetUsers() ([]User, error) {

}

func closeDbFile(file io.ReadCloser) {
	if closeErr := file.Close(); closeErr != nil {
		log.Fatalf("Error with closing file!")
	}
}
