package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kashyab12/chippy/chandler"
	"github.com/kashyab12/chippy/internal/database"
	"io/fs"
	"log"
	"net/http"
	"os"
)

func main() {
	if envVarLoadErr := godotenv.Load(); envVarLoadErr != nil {
		return
	}
	dbg := flag.Bool("debug", false, "Enable debug mode")
	if flag.Parse(); *dbg {
		if removeErr := os.Remove(database.ChibeFile); removeErr != nil && !errors.Is(removeErr, fs.ErrNotExist) {
			log.Fatalf("Could not remove %v\n", database.ChibeFile)
		} else {
			log.Printf("New %v will be instantiated\n", database.ChibeFile)
		}
	}
	const port = 8080
	config := chandler.ApiConfig{FsHits: 0, JwtSecret: os.Getenv("JWT_SECRET"), PolkaKey: os.Getenv("POLKA_KEY")}
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
