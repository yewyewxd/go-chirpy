package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) polkaWebhooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	var body struct {
		Event string `json:"event"`
		Data  struct {
			UserId string `json:"user_id"`
		} `json:"data"`
	}
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(&body); err != nil {
		respondWithError(w, 500, "Something went wrong")
		return
	}

	// Parse user id
	userID, err := uuid.Parse(body.Data.UserId)
	if err != nil {
		respondWithError(w, 400, "Invalid chirp id")
		return
	}

	// Event switch
	switch body.Event {
	case "user.upgraded":
		if err := cfg.db.UpgradeUserToChirpyRed(r.Context(), userID); err != nil {
			respondWithJSON(w, 404, "User not found")
			return
		}
		respondWithJSON(w, 204, nil)
	default:
		respondWithJSON(w, 204, nil)
	}
}
