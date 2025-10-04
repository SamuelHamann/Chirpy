package apiConfig

import (
	"net/http"
	"encoding/json"
	"github.com/google/uuid"
	"database/sql"
	"github.com/samuelhamann/chirpy/internal/auth"
)

type PolkaWebhookPayload struct {
	Event string `json:"event"`
	Data  struct {
		UserId string `json:"user_id"`
	} `json:"data"`
}

func (cfg *ApiConfig) HandlePolkaWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`"error": "Something went wrong"`))
		return
	}

	polkaKey, err := auth.GetApiKey(r.Header)
	if err != nil || polkaKey != cfg.PolkaKey {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`"error": "Invalid API key"`))
		return
	}

	var payload PolkaWebhookPayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`"error": "Invalid payload"`))
		return
	}

	if payload.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userID, err := uuid.Parse(payload.Data.UserId)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`"error": "Invalid user ID"`))
		return
	}
	
	_, err = cfg.Database.UpgradeUserToChirpyRed(r.Context(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`"error": "User not found"`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "Failed to update user"`))
		return
	}

	w.WriteHeader(http.StatusNoContent)
	return
}