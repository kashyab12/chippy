package chandler

import (
	"encoding/json"
	"github.com/kashyab12/chippy/internal/database"
	"log"
	"net/http"
	"slices"
	"strconv"
)

func DecodeRequestBody[J *BodyJson | *UserJson](r *http.Request, bodyStructure J) (J, error) {
	decoder := json.NewDecoder(r.Body)
	defer CloseIoReadCloserStream(r.Body)
	err := decoder.Decode(bodyStructure)
	return bodyStructure, err
}

func EncodeErrorResponse(responseBody ErrorJson) ([]byte, error) {
	encodedBody, encodingErr := json.Marshal(responseBody)
	return encodedBody, encodingErr
}

func invalidChippyRequestStruct(w http.ResponseWriter) {
	log.Printf("Error decoding body JSON params!")
	if encodedErrJson, encodingErr := EncodeErrorResponse(ErrorJson{ErrMsg: "Something went wrong"}); encodingErr != nil {
		log.Println("Inception wtf!")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, writeErr := w.Write(encodedErrJson)
		if writeErr != nil {
			log.Println("Stopping this right now lol")
		}
	}
}

func chippyTooLong(w http.ResponseWriter) {
	log.Printf("Chippy too damn long!")
	if encodedErrJson, encodingErr := json.Marshal(ErrorJson{ErrMsg: "Chirp is too long"}); encodingErr != nil {
		log.Println("Inception wtf!")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, writeErr := w.Write(encodedErrJson)
		if writeErr != nil {
			log.Println("Stopping this right now lol")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func PostChirp(w http.ResponseWriter, r *http.Request) {
	log.Println("Validating Chippy!")
	const MaxChippyLen = 140
	if jsonBody, decodeErr := DecodeRequestBody(r, &BodyJson{}); decodeErr != nil {
		invalidChippyRequestStruct(w)
	} else if len(jsonBody.Body) > MaxChippyLen {
		chippyTooLong(w)
	} else {
		if chibeDb, newDbErr := database.NewDB(database.ChibeFile); newDbErr != nil {
			log.Printf("Error while creating the database: %v\n", newDbErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if chirp, createErr := chibeDb.CreateChirp(jsonBody.Body); createErr != nil {
			log.Printf("Error while creating the chirp: %v\n", createErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if rawJson, encodeErr := json.Marshal(chirp); encodeErr != nil {
			log.Printf("Error while encoding the chirp to raw json %v: %v\n", rawJson, encodeErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			log.Println("Successfully encoded chirp and saved within CHIBE")
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_, err := w.Write(rawJson)
			if err != nil {
				return
			}
		}
	}
}

func GetChirp(w http.ResponseWriter, r *http.Request) {
	if chibeDb, newDbErr := database.NewDB(database.ChibeFile); newDbErr != nil {
		log.Printf("Error while creating the database: %v\n", newDbErr)
		w.WriteHeader(http.StatusInternalServerError)
	} else if chirps, getChirpsErr := chibeDb.GetChirps(); getChirpsErr != nil {
		log.Printf("Error while obtaining chibe entries: %v\n", getChirpsErr)
		w.WriteHeader(http.StatusInternalServerError)
	} else if rawJsonList, encodingErr := json.Marshal(chirps); encodingErr != nil {
		log.Printf("Error while encoding chibe entries: %v\n", encodingErr)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, err := w.Write(rawJsonList)
		if err != nil {
			return
		}
	}
}

func GetSingleChirp(w http.ResponseWriter, r *http.Request) {
	if targetChirpIdStr := r.PathValue("chirpID"); len(targetChirpIdStr) < 1 {
		log.Printf("Unable to match to chirp id based on provided path\n")
		w.WriteHeader(http.StatusInternalServerError)
	} else if targetChirpId, convErr := strconv.Atoi(targetChirpIdStr); convErr != nil {
		log.Printf("Conversion error of chirp id from string to integer\n")
		w.WriteHeader(http.StatusInternalServerError)
	} else if chibeDb, newDbErr := database.NewDB(database.ChibeFile); newDbErr != nil {
		log.Printf("Error while creating the database: %v\n", newDbErr)
		w.WriteHeader(http.StatusInternalServerError)
	} else if chirps, getChirpsErr := chibeDb.GetChirps(); getChirpsErr != nil {
		log.Printf("Error while obtaining chibe entries: %v\n", getChirpsErr)
		w.WriteHeader(http.StatusInternalServerError)
	} else if targetIdx := slices.IndexFunc(chirps, func(ch database.Chirp) bool {
		return ch.Uid == targetChirpId
	}); targetIdx == -1 {
		log.Printf("Chirp ID not present within chibe the db\n")
		w.WriteHeader(http.StatusNotFound)
	} else if rawData, encodingErr := json.Marshal(chirps[targetIdx]); encodingErr != nil {
		log.Printf("Error while encoding target chibe\n")
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, err := w.Write(rawData)
		if err != nil {
			return
		}
	}
}

func postUsers(w http.ResponseWriter, r *http.Request) {
	if jsonBody, decodeErr := DecodeRequestBody(r, &UserJson{}); decodeErr != nil {
		invalidChippyRequestStruct(w)
	} else {
		if chibeDb, newDbErr := database.NewDB(database.ChibeFile); newDbErr != nil {
			log.Printf("Error while creating the database: %v\n", newDbErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if chirp, createErr := chibeDb.CreateChirp(jsonBody.Body); createErr != nil {
			log.Printf("Error while creating the chirp: %v\n", createErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else if rawJson, encodeErr := json.Marshal(chirp); encodeErr != nil {
			log.Printf("Error while encoding the chirp to raw json %v: %v\n", rawJson, encodeErr)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			log.Println("Successfully encoded chirp and saved within CHIBE")
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_, err := w.Write(rawJson)
			if err != nil {
				return
			}
		}
}
