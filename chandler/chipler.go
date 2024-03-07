package chandler

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
)

func DecodeRequestBody(r *http.Request, bodyStructure *BodyJson) (*BodyJson, error) {
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

func isChippyTooLong(chip string, w http.ResponseWriter) {
	const MaxChippyLen = 140
	if len(chip) > MaxChippyLen {
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

func ValidateChippy(w http.ResponseWriter, r *http.Request) {
	log.Println("Validating Chippy!")

	if jsonBody, decodeErr := DecodeRequestBody(r, &BodyJson{}); decodeErr != nil {
		invalidChippyRequestStruct(w)
	} else if
	} else {
		badWords := []string{"kerfuffle", "sharbert", "fornax"}
		splitString := strings.Split(jsonBody.Body, " ")
		for idx, word := range splitString {
			if slices.Contains(badWords, strings.ToLower(word)) {
				splitString[idx] = "****"
			}
		}
		validJson := struct {
			ProfanedString string `json:"cleaned_body"`
		}{ProfanedString: strings.Join(splitString, " ")}
		if encodedValidJson, encodingErr := json.Marshal(validJson); encodingErr != nil {
			log.Println("Inception wtf!")
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, writeErr := w.Write(encodedValidJson)
			if writeErr != nil {
				log.Println("Stopping this right now lol")
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	}
}

func PostChirp(w http.ResponseWriter, r *http.Request) {
	jsonDecoder := json.NewDecoder(r.Body)
	defer CloseIoReadCloserStream(r.Body)
	postBody := BodyJson{}
	decodeErr := jsonDecoder.Decode(&postBody)
	if decodeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Something went wrong trying to decode the postChirp JSON")
	} else {

	}

}
