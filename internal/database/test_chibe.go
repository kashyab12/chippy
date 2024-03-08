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

func TestCreateChirp() {
	chibe, _ := TestNewDb()
	chirps := []string{"chibeeeee the deeeebeeee :DD::D::D::D::D::DD:", "cool", "awesome"}
	for _, chirp := range chirps {
		newChirp, err := chibe.CreateChirp(chirp)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("the new chirp: %v\n", newChirp)
	}
}
