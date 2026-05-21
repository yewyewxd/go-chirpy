package main

import (
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
