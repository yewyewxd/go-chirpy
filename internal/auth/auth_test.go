package auth

import (
	"testing"
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
