package main

import (
	"encoding/json"
	"fmt"
	chi2 "github.com/go-chi/chi/v5"
	"github.com/kashyab12/chippy/chippleware"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
)

type apiConfig struct {
	fsHits int
}

type errorJson struct {
	ErrMsg string `json:"error"`
}

type bodyJson struct {
	Body string `json:"body"`
}

func (cfg *apiConfig) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cfg.fsHits += 1
		next.ServeHTTP(writer, request)
	})
}

func (cfg *apiConfig) fsHitsHandler(writer http.ResponseWriter, _ *http.Request) {
	log.Println("fsHitsHandler ep hit!")
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	hitsString := fmt.Sprintf("<html>\n\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p>\n</body>\n\n</html>\n", cfg.fsHits)
	_, err := writer.Write([]byte(hitsString))
	if err != nil {
		log.Fatal(err)
		return
	}
}

func (cfg *apiConfig) resetFsHitsHandler(writer http.ResponseWriter, _ *http.Request) {
	log.Println("resetFsHandler ep hit!")
	cfg.fsHits = 0
	writer.WriteHeader(http.StatusOK)
}

func readinessEndpoint(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	log.Println("Readiness endpoint toggled.")
	_, err := w.Write([]byte("OK"))
	if err != nil {
		return
	}
}

func validateChippy(w http.ResponseWriter, r *http.Request) {
	log.Println("Validating Chippy!")
	const MaxChippyLen = 140
	decoder := json.NewDecoder(r.Body)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(r.Body)
	jsonBody := bodyJson{}
	if err := decoder.Decode(&jsonBody); err != nil {
		log.Printf("Error decoding body JSON params!")
		if encodedErrJson, encodingErr := json.Marshal(errorJson{ErrMsg: "Something went wrong"}); encodingErr != nil {
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
	} else if len(jsonBody.Body) > MaxChippyLen {
		log.Printf("Chippy too damn long!")
		if encodedErrJson, encodingErr := json.Marshal(errorJson{ErrMsg: "Chirp is too long"}); encodingErr != nil {
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

func postChirp(w http.ResponseWriter, r *http.Request) {
	jsonDecoder := json.NewDecoder(r.Body)
	defer closeIoReadCloserStream(r.Body)
	postBody := bodyJson{}
	decodeErr := jsonDecoder.Decode(&postBody)
	if decodeErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Something went wrong trying to decode the postChirp JSON")
	} else {

	}

}

func closeIoReadCloserStream(stream io.ReadCloser) {
	err := stream.Close()
	if err != nil {
		return
	}
}

func main() {
	const port = 8080
	apiCfg := apiConfig{fsHits: 0}
	fsHandler := apiCfg.metricsMiddleware(http.StripPrefix("/app", http.FileServer(http.Dir("./"))))

	appRouter := chi2.NewRouter()
	appRouter.Handle("/app/*", fsHandler)
	appRouter.Handle("/app", fsHandler)

	apiRouter := chi2.NewRouter()
	apiRouter.Get("/healthz", readinessEndpoint)
	apiRouter.HandleFunc("/reset", apiCfg.resetFsHitsHandler)
	apiRouter.Post("/chirps", postChirp)
	appRouter.Mount("/api", apiRouter)

	adminRouter := chi2.NewRouter()
	adminRouter.Get("/metrics", apiCfg.fsHitsHandler)
	appRouter.Mount("/admin", adminRouter)

	corsMux := chippleware.CorsMiddleware(appRouter)
	server := http.Server{
		Handler: corsMux,
		Addr:    fmt.Sprintf(":%v", port),
	}
	err := server.ListenAndServe()
	if err != nil {
		return
	}
}
