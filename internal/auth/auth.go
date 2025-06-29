package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	strTok, err := token.SignedString(tokenSecret)
	if err != nil {
		return "", fmt.Errorf("can't sign token: %w", err)
	}

	return strTok, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (any, error) {
		return []byte("somestringquestion"), nil
	})

	if err != nil {
		return uuid.Nil, fmt.Errorf("can't parse jwt token: %w", err)
	}
	subject, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, fmt.Errorf("can't get subject out of jwt: %w", err)
	}
	id, err := uuid.Parse(subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("can't parse uuid from jwt: %w", err)
	}
	return id, nil
}
