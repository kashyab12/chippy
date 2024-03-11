package chandler

import (
	"fmt"
	"log"
	"net/http"
)

func (config *ApiConfig) fsHitsHandler(writer http.ResponseWriter, _ *http.Request) {
	log.Println("fsHitsHandler ep hit!")
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	hitsString := fmt.Sprintf("<html>\n\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p>\n</body>\n\n</html>\n", config.FsHits)
	_, err := writer.Write([]byte(hitsString))
	if err != nil {
		log.Fatal(err)
		return
	}
}

func (config *ApiConfig) resetFsHitsHandler(writer http.ResponseWriter, _ *http.Request) {
	log.Println("resetFsHandler ep hit!")
	config.FsHits = 0
	writer.WriteHeader(http.StatusOK)
}
