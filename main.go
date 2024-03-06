package main

import (
	"fmt"
	chi2 "github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

type apiConfig struct {
	fsHits int
}

func (cfg *apiConfig) metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cfg.fsHits += 1
		next.ServeHTTP(writer, request)
	})
}

func (cfg *apiConfig) fsHitsHandler(writer http.ResponseWriter, _ *http.Request) {
	log.Println("fsHitsHandler ep hit!")
	writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	hitsString := fmt.Sprintf("Hits: %d", cfg.fsHits)
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

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		writer.Header().Set("Access-Control-Allow-Headers", "*")
		if request.Method == "OPTIONS" {
			writer.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func main() {
	const port = 8080
	router := chi2.NewRouter()
	apiCfg := apiConfig{fsHits: 0}
	fsHandler := apiCfg.metricsMiddleware(http.StripPrefix("/app", http.FileServer(http.Dir("./"))))
	router.Handle("/app/*", fsHandler)
	router.Handle("/app", fsHandler)
	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		log.Println("Readiness endpoint toggled.")
		_, err := w.Write([]byte("OK"))
		if err != nil {
			return
		}
	})
	router.Get("/metrics/", apiCfg.fsHitsHandler)
	router.HandleFunc("/reset/", apiCfg.resetFsHitsHandler)
	corsMux := corsMiddleware(router)
	server := http.Server{
		Handler: corsMux,
		Addr:    fmt.Sprintf(":%v", port),
	}
	err := server.ListenAndServe()
	if err != nil {
		return
	}
}
