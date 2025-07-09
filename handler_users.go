package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/17xande/bd-chirpy/internal/auth"
	"github.com/17xande/bd-chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "Coundn't decode parameters", err)
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error hashing password", nil)
		return
	}

	userParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hash,
	}

	u, err := cfg.db.CreateUser(context.Background(), userParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating user", err)
		return
	}

	res := response{User{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
	}}

	respondWithJSON(w, http.StatusCreated, res)
}

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "Couldn't decode parameters", err)
		return
	}

	expires := time.Hour

	if params.ExpiresInSeconds > 0 && params.ExpiresInSeconds < 60*60 {
		expires = time.Duration(params.ExpiresInSeconds) * time.Second
	}

	user, err := cfg.db.GetUserByEmail(context.Background(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error getting user by email", err)
		return
	}

	if err := auth.CheckPasswordHash(params.Password, user.HashedPassword); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, expires)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating JWT", err)
	}

	res := response{User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	}}

	respondWithJSON(w, http.StatusOK, res)
}
