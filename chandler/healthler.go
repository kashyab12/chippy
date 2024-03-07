package chandler

import (
	"log"
	"net/http"
)

func readinessEndpoint(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	log.Println("Readiness endpoint toggled.")
	_, err := w.Write([]byte("OK"))
	if err != nil {
		return
	}
}
