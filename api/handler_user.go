package apiConfig

import (
	"net/http"
	"encoding/json"
)
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