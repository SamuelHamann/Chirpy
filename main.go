package main

import (
	"net/http"
	"sync/atomic"
	api "github.com/samuelhamann/chirpy/api"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"os"
	"database/sql"
	"github.com/samuelhamann/chirpy/internal/database"
	"fmt"
	"strings"
)

func main() {

	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	platform := os.Getenv("PLATFORM")
	fmt.Println("Starting Chirpy on platform:", platform)

	dbQueries := database.New(db)
	cfg := api.ApiConfig{
		FileserverHits: atomic.Int32{},
		Database: dbQueries,
		Platform: strings.ToUpper(platform),
	}
	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /admin/metrics", cfg.HandlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.HandlerReset)
	mux.HandleFunc("POST /api/validate_chirp", cfg.ValidateChirp)
	mux.HandleFunc("GET /api/healthz", handlerFunc)
	mux.HandleFunc("POST /api/users", cfg.CreateUser)
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	server.ListenAndServe()
}

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}