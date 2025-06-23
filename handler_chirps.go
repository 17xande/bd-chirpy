package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

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

func (cfg *apiConfig) handlerChirpCreate(w http.ResponseWriter, r *http.Request) {
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

	ok, cleanChirp, err := cfg.chirpsValidate(params.Body)
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
