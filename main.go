package main

import (
	"net/http"
	"sync/atomic"
	api "github.com/samuelhamann/chirpy/api"
	_ "github.com/lib/pq"
)

func main() {
	cfg := api.ApiConfig{
		FileserverHits: atomic.Int32{},
	}
	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /admin/metrics", cfg.HandlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.HandlerReset)
	mux.HandleFunc("POST /api/validate_chirp", cfg.ValidateChirp)
	mux.HandleFunc("GET /api/healthz", handlerFunc)
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