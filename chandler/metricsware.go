package chandler

import (
	"net/http"
)

func (config *ApiConfig) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		config.FsHits += 1
		next.ServeHTTP(writer, request)
	})
}
