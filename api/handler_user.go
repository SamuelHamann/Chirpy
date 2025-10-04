package apiConfig

import (
	"net/http"
	"encoding/json"
	"github.com/samuelhamann/chirpy/internal/auth"
	"github.com/samuelhamann/chirpy/internal/database"
	"time"
	"github.com/google/uuid"
	"fmt"
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
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&u)

	if err != nil || len(u.Email) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`"error": "Something went wrong -- email"`))
		return
	}
	
	hashedPassword, err := auth.HashPassword(u.Password)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "Something went wrong -- Hash"` + err.Error()))
		return
	}

	createUserParams := database.CreateUserParams{
		Email: u.Email,
		HashedPassword: hashedPassword,
	}

	user, err := cfg.Database.CreateUser(r.Context(), createUserParams)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "Something went wrong -- Database"` + err.Error()))
		return
	}
	user.HashedPassword = "" // Clear hashed password before sending response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(user); err != nil {
        w.Write([]byte(`{"error": "Something went wrong -- Encode"}`))
        return
    }
}

func (cfg *ApiConfig) LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`"error": "Something went wrong"`))
		return
	}
	var u struct {
		Email string `json:"email"`
		Password string `json:"password"`
		ExpiresInSeconds int64 `json:"expires_in_seconds"`
	}
	err := json.NewDecoder(r.Body).Decode(&u)

	if err != nil || len(u.Email) == 0 || len(u.Password) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`"error": "Something went wrong -- email or password"`))
		return
	}

	user, err := cfg.Database.GetUserByEmail(r.Context(), u.Email)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "Something went wrong -- Database"` + err.Error()))
		return
	}

	isValid, err := auth.CheckPasswordHash(u.Password, user.HashedPassword); if err != nil || !isValid {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`Incorrect email or password`))
		return
	}

	user.HashedPassword = ""

	expiresIn := int64(3600) // 24 hours in seconds
    if u.ExpiresInSeconds > 0 {
        expiresIn = u.ExpiresInSeconds
    }
	tokenString, err := auth.MakeJWT(user.ID, cfg.JWTSecret, time.Duration(expiresIn))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "Something went wrong -- Token"` + err.Error()))
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "Something went wrong -- Refresh Token"` + err.Error()))
		return
	}

	_, err = cfg.Database.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		UserID: user.ID,
		Token: refreshToken,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour), // Refresh token valid for 7 days
	})

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "Something went wrong -- Store Refresh Token"` + err.Error()))
		return
	}

	type loginResponse struct {
		Id uuid.UUID       `json:"id"`
		Is_chirpy_red bool `json:"is_chirpy_red"`
		Token string        `json:"token"`
		RefreshToken string `json:"refresh_token"`
		Email string	   `json:"email"`
	}
	respUser := loginResponse{
		Id: user.ID,
		Is_chirpy_red: user.IsChirpyRed,
		Token: tokenString,
		RefreshToken: refreshToken,
		Email: user.Email,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(respUser); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Something went wrong -- Encode"}`))
		return
	}
}

func (cfg *ApiConfig) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`"error": "Something went wrong"`))
		return
	}
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`"error": "Something went wrong -- Token"` + err.Error()))
		return
	}
	if len(token) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`"error": "Something went wrong -- No Token"`))
		return
	}
	fmt.Sprintf("Refreshing token: %s\n", token)
	rt, err := cfg.Database.GetRefreshToken(r.Context(), token)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`"error": "Something went wrong -- Database"` + err.Error()))
		return
	}

	newToken, err := auth.MakeJWT(rt.UserID, cfg.JWTSecret, time.Duration(3600))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "Something went wrong -- New Token"` + err.Error()))
		return
	}

	type refreshResponse struct {
		Token string `json:"token"`
	}
	resp := refreshResponse{
		Token: newToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Something went wrong -- Encode"}`))
		return
	}
}

func (cfg *ApiConfig) RevokeToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`"error": "Something went wrong"`))
		return
	}
	token, err := auth.GetBearerToken(r.Header)
	
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`"error": "Something went wrong -- Token"` + err.Error()))
		return
	}
	if len(token) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`"error": "Something went wrong -- No Token"`))
		return
	}

	_, err = cfg.Database.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`"error": "Something went wrong -- Database"` + err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

func (cfg *ApiConfig) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
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

	var u struct {
		Email string `json:"email"`
		Password string `json:"password"`
	}
	err = json.NewDecoder(r.Body).Decode(&u)

	if err != nil || (len(u.Email) == 0 && len(u.Password) == 0) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`"error": "Something went wrong -- email or password"`))
		return
	}

	var hashedPassword string
	if len(u.Password) > 0 {
		hashedPassword, err = auth.HashPassword(u.Password)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`"error": "Something went wrong -- Hash"` + err.Error()))
			return
		}
	}
	
	updateUserParams := database.UpdateUserParams{
		ID: userId,
		Column2: u.Email,
		Column3: hashedPassword,
	}

	user, err := cfg.Database.UpdateUser(r.Context(), updateUserParams)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`"error": "Something went wrong -- Database"` + err.Error()))
		return
	}
	user.HashedPassword = "" // Clear hashed password before sending response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		w.Write([]byte(`{"error": "Something went wrong -- Encode"}`))
		return
	}
}