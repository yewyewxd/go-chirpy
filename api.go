package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/yewyewxd/go-chirpy/internal/auth"
	"github.com/yewyewxd/go-chirpy/internal/database"
)

// Response data types
type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

type User struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

// Helpers
func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	resBody := struct {
		Error string `json:"error"`
	}{Error: msg}
	bytes, err := json.Marshal(resBody)
	log.Printf("Error decoding parameters: %s", err)
	w.Write(bytes)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.WriteHeader(code)
	bytes, _ := json.Marshal(payload)
	w.Write(bytes)
}

// Routes
type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
	polkaKey       string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(200)
	msg := fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>
	`, cfg.fileserverHits.Load())
	w.Write([]byte(msg))
}

func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	if cfg.platform != "dev" {
		respondWithError(w, 403, "Dev only feature")
		return
	}

	if err := cfg.db.DeleteUsers(r.Context()); err != nil {
		log.Printf("Error deleting users: %s", err)
		return
	}

	w.WriteHeader(200)
	cfg.fileserverHits.Swap(0)
	w.Write([]byte("Hits: 0"))
}

func handlerHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, "Failed to get chirps")
		return
	}

	data := make([]Chirp, 0, len(chirps))

	for _, chirp := range chirps {
		data = append(data, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	respondWithJSON(w, 200, data)
}

func (cfg *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	chirpId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 400, "Invalid chirp id")
		return
	}

	chirp, err := cfg.db.GetChirpById(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, 404, "Chirp does not exist")
		return
	}

	respondWithJSON(w, 200, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	var body struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&body); err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	// Validate chirp
	if len(body.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	// Validate user
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "User not logged in")
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "Failed to validate user")
		return
	}

	// Create chirp for user
	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleanProfane(body.Body),
		UserID: userID,
	})
	if err != nil {
		respondWithError(w, 500, "Failed to create chirp")
		return
	}

	respondWithJSON(w, 201, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&body); err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	hashedPw, err := auth.HashPassword(body.Password)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          body.Email,
		HashedPassword: sql.NullString{String: hashedPw, Valid: true},
	})
	if err != nil {
		respondWithError(w, 400, "Failed to create user")
		return
	}

	respondWithJSON(w, 201, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&body); err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	// Get user hashed pw
	user, err := cfg.db.GetUser(r.Context(), body.Email)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	// Validate user pw
	match, err := auth.CheckPasswordHash(body.Password, user.HashedPassword.String)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}
	if !match {
		respondWithError(w, 401, "Incorrect email or password")
	}

	// Generate JWT
	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	// Create refresh token
	refreshToken, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     auth.MakeRefreshToken(),
		ExpiresAt: time.Now().Add(time.Hour * 24 * 60),
		UserID:    user.ID,
	})
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	respondWithJSON(w, 200, User{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken.Token,
		IsChirpyRed:  user.IsChirpyRed,
	})
}

func (cfg *apiConfig) refresh(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	// Validate token
	rt, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Token is required")
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), rt)
	if err != nil {
		respondWithError(w, 401, "Failed to retrieve token")
		return
	}

	if time.Now().Compare(refreshToken.ExpiresAt) > -1 || refreshToken.RevokedAt.Valid {
		respondWithError(w, 401, "Invalid refresh token")
		return
	}

	// Create new token for user
	token, err := auth.MakeJWT(refreshToken.UserID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	respondWithJSON(w, 200, struct {
		Token string `json:"token"`
	}{Token: token})
}

func (cfg *apiConfig) revoke(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	// Validate token
	rt, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Token is required")
		return
	}

	// Revoke token
	if err := cfg.db.RevokeRefreshToken(r.Context(), rt); err != nil {
		respondWithError(w, 401, "Invalid refresh token")
	}

	respondWithJSON(w, 204, nil)
}

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&body); err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	// Validate user
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "User not logged in")
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "Failed to validate user")
		return
	}

	// Hash password
	hashedPw, err := auth.HashPassword(body.Password)
	if err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	// Update user
	user, err := cfg.db.UpdateUserById(r.Context(), database.UpdateUserByIdParams{
		ID:             userID,
		Email:          body.Email,
		HashedPassword: sql.NullString{String: hashedPw, Valid: true},
	})

	respondWithJSON(w, 200, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}

func (cfg *apiConfig) deleteChirp(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	// Get chirp id
	chirpId, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 400, "Invalid chirp id")
		return
	}

	// Validate user
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "User not logged in")
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, 401, "Failed to validate user")
		return
	}

	// Get chirp
	chirp, err := cfg.db.GetChirpById(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, 404, "Chirp does not exist")
		return
	}

	// Check if chirp belongs to user
	if chirp.UserID != userID {
		respondWithError(w, 403, "Cannot delete chirp")
		return
	}

	// Delete chirp for user
	if err := cfg.db.DeleteChirpForUser(r.Context(), database.DeleteChirpForUserParams{
		ID:     chirpId,
		UserID: userID,
	}); err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	respondWithJSON(w, 204, nil)
}
