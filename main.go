package main

import (
	"fmt"
	"net/http"
)

type apiHandler struct{}
type corsHandler struct {
}

func (handler apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request)  {}
func (handler corsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/api", apiHandler{})
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/" {
			http.NotFound(writer, request)
			return
		}
		_, err := fmt.Fprintf(writer, "Welcome to the home page!!!")
		if err != nil {
			return
		}
	})
}
