package main

import (
	"context"
	"log"
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Can only reset environment in dev platform.", nil)
		return
	}

	cfg.fileserverHits.Store(0)
	rowsAffected, err := cfg.db.Reset(context.Background())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Coundn't delete all users", err)
		return
	}

	log.Printf("Deleted Users: %d\n", rowsAffected)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}
