package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
)

var forbiddenWords = []string{"kerfuffle", "sharbert", "fornax"}

func cleanProfane(text string) string {
	words := strings.Split(text, " ")
	for i, word := range words {
		if slices.Contains(forbiddenWords, strings.ToLower(word)) {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

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
