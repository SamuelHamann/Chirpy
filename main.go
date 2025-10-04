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

	JWTSecret := os.Getenv("JWT_SECRET")
	if len(JWTSecret) == 0 {
		fmt.Println("JWT_SECRET is not set")
		os.Exit(1)
	}
	PolkaKey:= os.Getenv("POLKA_KEY")
	if len(PolkaKey) == 0 {
		fmt.Println("POLKA_KEY is not set")
		os.Exit(1)
	}
	platform := os.Getenv("PLATFORM")
	fmt.Println("Starting Chirpy on platform:", platform)

	dbQueries := database.New(db)
	cfg := api.ApiConfig{
		FileserverHits: atomic.Int32{},
		Database: dbQueries,
		Platform: strings.ToUpper(platform),
		JWTSecret: JWTSecret,
		PolkaKey: PolkaKey,
	}
	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /admin/metrics", cfg.HandlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.HandlerReset)
	mux.HandleFunc("POST /api/validate_chirp", cfg.ValidateChirp)
	mux.HandleFunc("GET /api/healthz", handlerFunc)
	mux.HandleFunc("POST /api/users", cfg.CreateUser)
	mux.HandleFunc("POST /api/login", cfg.LoginUser)
	mux.HandleFunc("POST /api/chirps", cfg.CreateChirp)
	mux.HandleFunc("GET /api/chirps", cfg.GetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.GetChirpByID)
	mux.HandleFunc("POST /api/refresh", cfg.RefreshToken)
	mux.HandleFunc("POST /api/revoke", cfg.RevokeToken)
	mux.HandleFunc("PUT /api/users", cfg.UpdateUser)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.DeleteChirp)
	mux.HandleFunc("POST /api/polka/webhooks", cfg.HandlePolkaWebhook)

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