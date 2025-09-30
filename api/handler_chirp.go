package apiConfig

import (
	"net/http"
	"encoding/json"
	"fmt"
	"strings"
	"github.com/samuelhamann/chirpy/internal/database"
	"github.com/google/uuid"
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

func sanitize(input string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax", "Kerfuffle", "Sharbert", "Fornax"} // Replace with your actual words
	for _, word := range badWords {
		input = strings.ReplaceAll(input, word, "****")
	}
	return input
}