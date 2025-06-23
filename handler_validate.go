package main

import (
	"fmt"
	"strings"
)

func (cfg *apiConfig) chirpsValidate(chirp string) (bool, string, error) {
	const maxChirpLength = 140
	if len(chirp) > maxChirpLength {
		return false, "", fmt.Errorf("chirp is too long")
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	cleaned := getCleanedBody(chirp, badWords)

	return true, cleaned, nil
}

func getCleanedBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}

	return strings.Join(words, " ")
}
