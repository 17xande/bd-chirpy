package auth

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TokenType string

const TokenTypeAccess TokenType = "chirpy-access"

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	signinKey := []byte(tokenSecret)
	claims := jwt.RegisteredClaims{
		Issuer:    string(TokenTypeAccess),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(signinKey)
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("can't parse jwt token: %w", err)
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("can't get claim subject: %w", err)
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, fmt.Errorf("can't get claim issuer: %w", err)
	}

	if issuer != string(TokenTypeAccess) {
		return uuid.Nil, errors.New("invalid issuer")
	}

	id, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("can't parse uuid from jwt: %w", err)
	}
	return id, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	token := headers.Get("Authorization")
	strip := "Bearer "
	if token == "" {
		return "", errors.New("No authorization header")
	}
	if len(token) <= len(strip) || token[0:len(strip)] != strip {
		return "", fmt.Errorf("Invalid authorization header: %v", token)
	}
	return token[len(strip):], nil
}
