package chandler

import (
	"encoding/json"
	"log"
	"net/http"
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

//func unprofaneChip(chip string) string {
//	badWords := []string{"kerfuffle", "sharbert", "fornax"}
//	splitString := strings.Split(chip, " ")
//	for idx, word := range splitString {
//		if slices.Contains(badWords, strings.ToLower(word)) {
//			splitString[idx] = "****"
//		}
//	}
//	return strings.Join(splitString, " ")
//}

func PostChirp(w http.ResponseWriter, r *http.Request) {
	log.Println("Validating Chippy!")
	const MaxChippyLen = 140
	if jsonBody, decodeErr := DecodeRequestBody(r, &BodyJson{}); decodeErr != nil {
		invalidChippyRequestStruct(w)
	} else if len(jsonBody.Body) > MaxChippyLen {
		chippyTooLong(w)
	} else {
		// TODO: Assign UID (save within DB) and return response
	}
}
