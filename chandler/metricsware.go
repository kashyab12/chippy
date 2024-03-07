package chandler

import (
	"net/http"
)

func (cfg *ApiConfig) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		cfg.FsHits += 1
		next.ServeHTTP(writer, request)
	})
}
