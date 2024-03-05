package main

import (
	"fmt"
	"log"
	"net/http"
)

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
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir("./"))))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		log.Println("Readiness endpoint toggled.")
		_, err := w.Write([]byte("OK"))
		if err != nil {
			return
		}
	})
	corsMux := corsMiddleware(mux)
	server := http.Server{
		Handler: corsMux,
		Addr:    fmt.Sprintf(":%v", port),
	}
	err := server.ListenAndServe()
	if err != nil {
		return
	}
}
