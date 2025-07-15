package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPolkaWebhook(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "Couldn't decode parameters", err)
		return
	}

	if params.Event != "user.upgraded" {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	id, err := uuid.Parse(params.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "can't parse uuid from jwt:", err)
		return
	}
	_, err = cfg.db.UpgradeUserToRed(context.Background(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Error trying to upgrade user to Red", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
