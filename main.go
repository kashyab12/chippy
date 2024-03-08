package main

import (
	"fmt"
	"github.com/kashyab12/chippy/chandler"
	"github.com/kashyab12/chippy/internal/database"
	"net/http"
)

func main() {
	const port = 8080
	database.TestCreateChirp()
	config := chandler.ApiConfig{FsHits: 0}
	appRouter := chandler.GetAppRouter(&config)
	apiRouter := chandler.GetApiRouter(&config)
	adminRouter := chandler.GetAdminRouter(&config)

	appRouter.Mount("/api", apiRouter)
	appRouter.Mount("/admin", adminRouter)
	corsMux := chandler.CorsMiddleware(appRouter)
	server := http.Server{
		Handler: corsMux,
		Addr:    fmt.Sprintf(":%v", port),
	}
	err := server.ListenAndServe()
	if err != nil {
		return
	}
}
