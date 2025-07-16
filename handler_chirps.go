package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/17xande/bd-chirpy/internal/auth"
	"github.com/17xande/bd-chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get token", err)
		return
	}

	ID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get user ID from token", err)
		return
	}

	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	type response struct {
		Chirp
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "Cound't decode parameters", err)
		return
	}

	params.UserID = ID

	ok, cleanChirp, err := cfg.validateChirp(params.Body)
	if !ok {
		respondWithError(w, http.StatusUnprocessableEntity, "Can't valicate chirp", err)
		return
	}

	chirpParams := database.CreateChirpsParams{
		Body:   cleanChirp,
		UserID: params.UserID,
	}

	chirp, err := cfg.db.CreateChirps(context.Background(), chirpParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating chirp", err)
		return
	}

	res := response{Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}}

	respondWithJSON(w, http.StatusCreated, res)
}

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Chirps []Chirp
	}

	queryParams := r.URL.Query()
	author_id := queryParams.Get("author_id")
	sortParam := queryParams.Get("sort")
	if sortParam == "" {
		sortParam = "asc"
	}
	var chirps []database.Chirp
	if author_id == "" {
		var err error
		chirps, err = cfg.db.GetChirps(context.Background())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Can't get all chirps", err)
			return
		}
	} else {
		uid, err := uuid.Parse(author_id)
		if err != nil {
			respondWithError(w, http.StatusUnprocessableEntity, "Invalid author_id", err)
			return
		}
		chirps, err = cfg.db.GetChirpsByAuthor(context.Background(), uid)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Can't get chirps for this user", err)
			return
		}
	}

	if sortParam == "desc" {
		sort.Slice(chirps, func(i, j int) bool { return chirps[i].CreatedAt.After(chirps[j].CreatedAt) })
	}

	resChirps := []Chirp{}

	for _, chirp := range chirps {
		c := Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		resChirps = append(resChirps, c)
	}

	respondWithJSON(w, http.StatusOK, resChirps)
}

func (cfg *apiConfig) handlerChirpGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	uid, err := uuid.Parse(id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't parse ID: "+id, err)
		return
	}

	chirp, err := cfg.db.GetChirp(context.Background(), uid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Can't get chirp with this ID", err)
		return
	}

	res := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, http.StatusOK, res)
}

func (cfg *apiConfig) validateChirp(chirp string) (bool, string, error) {
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

func (cfg *apiConfig) handlerChirpDelete(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("id")

	uid, err := uuid.Parse(chirpId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't parse ID: "+chirpId, err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get token", err)
		return
	}

	user_id, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get user ID from token", err)
		return
	}

	_, err = cfg.db.GetChirp(context.Background(), uid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Can't get chirp with this ID", err)
		return
	}

	args := database.DeleteChirpParams{
		ID:     uid,
		UserID: user_id,
	}

	rows, err := cfg.db.DeleteChirp(context.Background(), args)
	if err != nil || rows != 1 {
		respondWithError(w, http.StatusForbidden, "Couldn't delete this chirp", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
