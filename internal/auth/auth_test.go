package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestAuth(t *testing.T) {
	cases := []struct {
		input string
	}{
		{
			input: "password",
		},
		{
			input: "Charmander Bulbasaur",
		},
		{
			input: "A213%#xc@df#Zk",
		},
	}

	for _, c := range cases {
		hash, err := HashPassword(c.input)
		if err != nil {
			t.Errorf("Expected no error from HashPassword, got: %v", err)
		}

		if hash == c.input {
			t.Errorf("Expected hashed value, got: %v", err)
		}

		match, err := CheckPasswordHash(c.input, hash)
		if err != nil {
			t.Errorf("Expected no error from CheckPasswordHash, got: %v", err)
		}

		if !match {
			t.Errorf("Expected true from CheckPasswordHash, got: %v", match)
		}
	}
}

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "super-secret"
	expiresIn := time.Minute

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("Expected no error from MakeJWT, got: %v", err)
	}

	gotID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("Expected no error from ValidateJWT, got: %v", err)
	}
	if gotID != userID {
		t.Fatalf("Expected user ID %v, got: %v", userID, gotID)
	}
}

func TestValidateJWTWrongSecret(t *testing.T) {
	userID := uuid.New()
	secret := "super-secret"
	wrongSecret := "wrong-secret"

	token, err := MakeJWT(userID, secret, time.Minute)
	if err != nil {
		t.Fatalf("Expected no error from MakeJWT, got: %v", err)
	}

	_, err = ValidateJWT(token, wrongSecret)
	if err == nil {
		t.Fatalf("Expected error from ValidateJWT with wrong secret")
	}
}

func TestValidateJWTExpired(t *testing.T) {
	userID := uuid.New()
	secret := "super-secret"

	token, err := MakeJWT(userID, secret, -time.Minute)
	if err != nil {
		t.Fatalf("Expected no error from MakeJWT, got: %v", err)
	}

	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Fatalf("Expected error from ValidateJWT with expired token")
	}
}

func TestValidateJWTWrongSigningMethod(t *testing.T) {
	userID := uuid.New()
	secret := "super-secret"

	// new token with SigningMethodHS512
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		Subject:   userID.String(),
	})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Expected no error from SignedString, got: %v", err)
	}

	// verify token with SigningMethodHS256
	_, err = ValidateJWT(signed, secret)
	if err == nil {
		t.Fatalf("Expected error from ValidateJWT with wrong signing method")
	}
}
