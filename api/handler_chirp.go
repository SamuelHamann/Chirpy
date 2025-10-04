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
	"github.com/samuelhamann/chirpy/internal/auth"
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

	tokenString, err := auth.GetBearerToken(r.Header)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        w.Write([]byte(`{"error": "Missing or invalid bearer token"}`))
        return
    }

    userId, err := auth.ParseJWT(tokenString, cfg.JWTSecret)
    if err != nil {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusUnauthorized)
        w.Write([]byte(fmt.Sprintf(`{"error": "Invalid token: %v"}`, err)))
        return
    }

	var c struct {
		Body   string `json:"body"`
	}
	err = json.NewDecoder(r.Body).Decode(&c)
	if err != nil || len(c.Body) == 0 || len(c.Body) > 140 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`"error": "Something went wrong -- body"`))
		return
	}
	fmt.Sprintf("Creating chirp for user ID: %s\n", userId)

	chirp, err := cfg.Database.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   c.Body,
		UserID: userId,
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

func (cfg *ApiConfig) DeleteChirp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`"error": "Something went wrong"`))
		return
	}

	tokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`"error": "Something went wrong -- Token"` + err.Error()))
		return
	}
	if len(tokenString) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`"error": "Something went wrong -- No Token"`))
		return
	}

	userId, err := auth.ParseJWT(tokenString, cfg.JWTSecret)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`"error": "Something went wrong -- Parse JWT"` + err.Error()))
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

	_, err = cfg.Database.GetChirpById(r.Context(), uuidId)
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

	_, err = cfg.Database.DeleteChirp(r.Context(), database.DeleteChirpParams{
		ID:     uuidId,
		UserID: userId,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`"error": "Chirp not found or not owned by user"`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "Something went wrong -- Database"` + err.Error()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}

func sanitize(input string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax", "Kerfuffle", "Sharbert", "Fornax"} // Replace with your actual words
	for _, word := range badWords {
		input = strings.ReplaceAll(input, word, "****")
	}
	return input
}