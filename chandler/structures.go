package chandler

import (
	chi2 "github.com/go-chi/chi/v5"
	"io"
	"net/http"
)

type ApiConfig struct {
	FsHits    int
	JwtSecret string
}

type ErrorJson struct {
	ErrMsg string `json:"error"`
}

type BodyJson struct {
	Body string `json:"body"`
}

type UserJson struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserReturnJson struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Token        string `json:"token,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

type RequestBody interface {
	BodyJson
}

func GetAppRouter(config *ApiConfig) *chi2.Mux {
	fsHandler := config.MetricsMiddleware(http.StripPrefix("/app", http.FileServer(http.Dir("./"))))
	appRouter := chi2.NewRouter()
	appRouter.Handle("/app/*", fsHandler)
	appRouter.Handle("/app", fsHandler)
	return appRouter
}

func GetApiRouter(config *ApiConfig) *chi2.Mux {
	apiRouter := chi2.NewRouter()
	apiRouter.Get("/healthz", readinessEndpoint)
	apiRouter.HandleFunc("/reset", config.resetFsHitsHandler)
	apiRouter.Get("/chirps", GetChirp)
	apiRouter.Get("/chirps/{chirpID}", GetSingleChirp)
	apiRouter.Post("/chirps", PostChirp)
	apiRouter.Post("/users", PostUsers)
	apiRouter.Put("/users", config.PutUser)
	apiRouter.Post("/login", config.PostLogin)
	return apiRouter
}

func GetAdminRouter(config *ApiConfig) *chi2.Mux {
	adminRouter := chi2.NewRouter()
	adminRouter.Get("/metrics", config.fsHitsHandler)
	return adminRouter
}

func CloseIoReadCloserStream(stream io.ReadCloser) {
	err := stream.Close()
	if err != nil {
		return
	}
}
