// Package database chibe the db :D
package database

import (
	"cmp"
	"encoding/json"
	"errors"
	"golang.org/x/crypto/bcrypt"
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
	Uid      int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
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
				return DBStructure{map[int]Chirp{}, map[int]User{}}, nil
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

// GetUsers returns all users in the database
func (chibe *DB) GetUsers() ([]User, error) {
	var users []User
	if dbStruct, loadErr := chibe.loadDB(); loadErr != nil {
		return nil, loadErr
	} else if len(dbStruct.Users) > 0 {
		chibe.mux.RLock()
		defer chibe.mux.RUnlock()
		for _, user := range dbStruct.Users {
			users = append(users, user)
		}
		slices.SortFunc(users, func(a, b User) int { return cmp.Compare(a.Uid, b.Uid) })
	}
	return users, nil
}

// CreateUser creates a new user and saves it to disk
func (chibe *DB) CreateUser(email, password string) (User, error) {
	var newUser User
	if users, getUserErr := chibe.GetUsers(); getUserErr != nil {
		return newUser, getUserErr
	} else if presentIdx := slices.IndexFunc(users, func(us User) bool {
		return us.Email == email
	}); presentIdx != -1 {
		log.Printf("user %v already exists\n", email)
		return newUser, errors.New("the mentioned user already exists")
	} else {
		newUserId := 1
		if len(users) > 0 {
			highestUid := users[len(users)-1].Uid
			newUserId = highestUid + 1
		}
		if dbStruct, loadErr := chibe.loadDB(); loadErr != nil {
			return newUser, loadErr
		} else if encryptedPass, encryptErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost); encryptErr != nil {
			log.Printf("Error encrypting the password of %v\n", email)
			return newUser, encryptErr
		} else {
			newUser = User{
				Uid:      newUserId,
				Email:    email,
				Password: string(encryptedPass),
			}
			dbStruct.Users[newUser.Uid] = newUser
			if writeErr := chibe.writeDB(dbStruct); writeErr != nil {
				return newUser, writeErr
			}
		}
	}
	return newUser, nil
}

func (chibe *DB) AuthUser(email, password string) (User, error) {
	var targetUser User
	if users, getUserErr := chibe.GetUsers(); getUserErr != nil {
		return targetUser, getUserErr
	} else if presentIdx := slices.IndexFunc(users, func(us User) bool {
		return us.Email == email
	}); presentIdx == -1 {
		log.Printf("user %v does not exist within chibe, need to create account!", email)
		return targetUser, errors.New("user not present")
	} else {
		// Check whether password matches
		targetUser = users[presentIdx]
		if matchErr := bcrypt.CompareHashAndPassword([]byte(targetUser.Password), []byte(password)); matchErr != nil {
			return targetUser, errors.New("password does not match")
		}
	}
	return targetUser, nil
}

func (chibe *DB) UpdateUser(targetUserId int, updatedEmail, updatedPassword string) (User, error) {
	var updatedUser User
	if users, getUserErr := chibe.GetUsers(); getUserErr != nil {
		return updatedUser, getUserErr
	} else if presentIdx := slices.IndexFunc(users, func(us User) bool {
		return us.Uid == targetUserId
	}); presentIdx == -1 {
		log.Printf("user with id %v does not exist within chibe\n", targetUserId)
		return updatedUser, errors.New("user does not exist")
	} else {
		updatedUser = User{
			Uid:      targetUserId,
			Email:    updatedEmail,
			Password: updatedPassword,
		}
		if dbStruct, loadErr := chibe.loadDB(); loadErr != nil {
			return updatedUser, loadErr
		} else {
			dbStruct.Users[targetUserId] = updatedUser
			if writeErr := chibe.writeDB(dbStruct); writeErr != nil {
				return updatedUser, writeErr
			}
		}
	}
	return updatedUser, nil
}

func closeDbFile(file io.ReadCloser) {
	if closeErr := file.Close(); closeErr != nil {
		log.Fatalf("Error with closing file!")
	}
}
