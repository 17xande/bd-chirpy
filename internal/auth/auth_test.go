package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestCheckPasswordHash(t *testing.T) {
	password1 := "correctPassword123!"
	password2 := "anotherPassword456!"
	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "Correct password",
			password: password1,
			hash:     hash1,
			wantErr:  false,
		},
		{
			name:     "Incorrect password",
			password: "wrong",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Password doesn't match different hash",
			password: password1,
			hash:     hash2,
			wantErr:  true,
		},
		{
			name:     "Empty password",
			password: "",
			hash:     hash1,
			wantErr:  true,
		},
		{
			name:     "Invalid hash",
			password: password1,
			hash:     "invalidhash",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPasswordHash() err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestJWTCreation(t *testing.T) {
	id1, _ := uuid.Parse("b8b69eb9-4fc8-4dcb-9d15-4978cbde44aa")
	secret1 := "secret1"
	expires1, _ := time.ParseDuration("24h")

	tests := []struct {
		name    string
		id      uuid.UUID
		secret  string
		expires time.Duration
		wantErr bool
	}{
		{
			name:    "first",
			id:      id1,
			secret:  secret1,
			expires: expires1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tok, err := MakeJWT(tt.id, tt.secret, tt.expires)
			if (err != nil) != tt.wantErr {
				t.Errorf("MakeJWT() err = %v, wantErr %v", err, tt.wantErr)
			}

			id, err := ValidateJWT(tok, tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateJWT() err = %v, wantErr %v", err, tt.wantErr)
			}

			if id != tt.id {
				t.Errorf("Validated ID doesn't match")
			}
		})
	}
}
