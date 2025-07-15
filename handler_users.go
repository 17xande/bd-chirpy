package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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

func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
		RefreshToken string `json:"refresh_token"`
	}

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

	userParams := database.UpdateUserParams{
		ID:             ID,
		Email:          params.Email,
		HashedPassword: hash,
	}

	user, err := cfg.db.UpdateUser(context.Background(), userParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating user in database", err)
		return
	}

	res := response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
			Token:     token,
		},

		// RefreshToken: refreshToken,
	}

	respondWithJSON(w, http.StatusOK, res)

}

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type response struct {
		User
		RefreshToken string `json:"refresh_token"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(context.Background(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error getting user by email", err)
		return
	}

	if err := auth.CheckPasswordHash(params.Password, user.HashedPassword); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
	}

	expires := time.Hour

	token, err := auth.MakeJWT(user.ID, cfg.secret, expires)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating JWT", err)
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error generating refreshToken", err)
	}

	args := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().UTC().Add(time.Hour * 24 * 60),
	}

	_, err = cfg.db.CreateRefreshToken(context.Background(), args)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error storing refreshToken in db", err)
	}

	res := response{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
			Token:     token,
		},

		RefreshToken: refreshToken,
	}

	respondWithJSON(w, http.StatusOK, res)
}

func (cfg *apiConfig) handlerGetRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get token", err)
		return
	}

	token, err := cfg.db.GetUserFromRefreshToken(context.Background(), refreshToken)
	if errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusUnauthorized, "Token doesn't exist", err)
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error trying to get token from database", err)
		return
	}
	if token.ExpiresAt.Before(time.Now().UTC()) {
		respondWithError(w, http.StatusUnauthorized, "Token expired", nil)
		return
	}

	if token.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "Token revoked", nil)
		return
	}

	newToken, err := auth.MakeJWT(token.UserID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Can't create new access token from refresh token", err)
		return
	}

	type response struct {
		Token string `json:"token"`
	}
	respondWithJSON(w, http.StatusOK, response{Token: newToken})
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't get token", err)
		return
	}

	_, err = cfg.db.RevokeRefreshToken(context.Background(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error trying to revoke token in database", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
