package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "mysecretpassword"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(hashedPassword) == 0 {
		t.Fatalf("Expected hashed password to be non-empty")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "mysecretpassword"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = CheckPasswordHash(password, hashedPassword)
	if err != nil {
		t.Fatalf("Expected password to match hash, got %v", err)
	}

	wrongPassword := "wrongpassword"
	err = CheckPasswordHash(wrongPassword, hashedPassword)
	if err == nil {
		t.Fatalf("Expected password to not match hash, but it did")
	}
}
