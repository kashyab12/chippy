package database

import "log"

func TestNewDb() (*DB, error) {
	chibe, err := NewDB("./database.json")
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Got the man chibee at %v\n", chibe.path)
	return chibe, err
}
