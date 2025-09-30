package apiConfig

import (
	"net/http"
	"sync/atomic"
	"fmt"
	"encoding/json"
	"strings"
	"github.com/samuelhamann/chirpy/internal/database"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	Database *database.Queries
	Platform string
}

func (cfg *ApiConfig) HandlerReset(w http.ResponseWriter, r *http.Request) {
	cfg.FileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))

	if cfg.Platform != "DEV" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Forbidden"))
		return
	}

	_, err := cfg.Database.DeleterAllUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error resetting users"))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("All users deleted"))

}

func (cfg *ApiConfig) HandlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.FileserverHits.Load())))
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *ApiConfig) ValidateChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`"error": "Something went wrong"`))
		return
	}
	var c Chirp
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`"error": "Something went wrong"`))
		return
	}
	if len(c.Body) == 0 || len(c.Body) > 140 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`"error": "Chirp is too long"`))
		return
	}

	c.Body = sanitize(c.Body)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"cleaned_body":"%s"}`, c.Body)))
}

func (cfg *ApiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`"error": "Something went wrong"`))
		return
	}
	var u struct {
		Email string `json:"email"`
	}
	err := json.NewDecoder(r.Body).Decode(&u)

	if err != nil || len(u.Email) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`"error": "Something went wrong -- email"`))
		return
	}
	
	user, err := cfg.Database.CreateUser(r.Context(), u.Email)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "Something went wrong -- Database"` + err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(user); err != nil {
        w.Write([]byte(`{"error": "Something went wrong -- Encode"}`))
        return
    }
}


func sanitize(input string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax", "Kerfuffle", "Sharbert", "Fornax"} // Replace with your actual words
	for _, word := range badWords {
		input = strings.ReplaceAll(input, word, "****")
	}
	return input
}
