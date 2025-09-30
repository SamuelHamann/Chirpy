package apiConfig

import (
	"net/http"
	"encoding/json"
	"fmt"
	"strings"
	"github.com/samuelhamann/chirpy/internal/database"
	"github.com/google/uuid"
	"database/sql"
	"errors"
)


func (cfg *ApiConfig) ValidateChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`"error": "Something went wrong"`))
		return
	}
	var c struct {
		Body string `json:"body"`
	}
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

func (cfg *ApiConfig) CreateChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`"error": "Something went wrong"`))
		return
	}
	var c struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&c)
	if err != nil || len(c.Body) == 0 || len(c.Body) > 140 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`"error": "Something went wrong -- body"`))
		return
	}
	fmt.Sprintf("Creating chirp for user ID: %s\n", c.UserID)
	uuidUser, err := uuid.Parse(c.UserID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`"error": "Something went wrong -- user_id"`))
		return
	}
	chirp, err := cfg.Database.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   c.Body,
		UserID: uuidUser,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "Something went wrong -- Database"` + err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(chirp)
	if err != nil {
		w.Write([]byte(`{"error": "Something went wrong -- Encode"}`))
		return
	}
}

func (cfg *ApiConfig) GetChirps(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`"error": "Something went wrong"`))
		return
	}

	chirps, err := cfg.Database.GetChirps(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "Something went wrong -- Database"` + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(chirps)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "Something went wrong -- Encode"`))
		return
	}
}

func (cfg *ApiConfig) GetChirpByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`"error": "Something went wrong"`))
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/api/chirps/")
	uuidId, err := uuid.Parse(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`"error": "Something went wrong -- chirpID"`))
		return
	}

	chirp, err := cfg.Database.GetChirpById(r.Context(), uuidId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`"error": "Chirp not found"`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "Something went wrong -- Database"` + err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(chirp)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "Something went wrong -- Encode"`))
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