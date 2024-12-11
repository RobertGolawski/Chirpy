package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

var tokenSecret string = "supersecret"
var tokenSecretWrong string = "notsupersecret"
var duration time.Duration = time.Minute

func TestMakeJWT(t *testing.T) {
	id, err := uuid.Parse("123e4567-e89b-12d3-a456-426614174000")
	if err != nil {
		t.Fatalf("Some error happened: %v", err)
	}

	tokenString, err := MakeJWT(id, tokenSecret, duration)
	if err != nil {
		t.Fatalf("Error happened during making: %v", err)
	}

	t.Log("tokenString: ", tokenString)
	if tokenString == "" {
		t.Fatalf("Expected a non-empty tokenString")
	}
}

func TestValidateJWTHappy(t *testing.T) {
	id, err := uuid.Parse("123e4567-e89b-12d3-a456-426614174000")
	if err != nil {
		t.Fatalf("Some error happened: %v", err)
	}

	tokenString, err := MakeJWT(id, tokenSecret, duration)
	if err != nil {
		t.Fatalf("Error happened during making: %v", err)
	}
	t.Logf("tokenString: %s", tokenString)
	idValidated, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Fatalf("Error with validation %v", err)
	}

	if idValidated.String() != id.String() {
		t.Fatalf("Expected: %s but got: %s", id.String(), idValidated.String())
	}
}
func TestValidateJWTWrongSecret(t *testing.T) {
	id, err := uuid.Parse("123e4567-e89b-12d3-a456-426614174000")
	if err != nil {
		t.Fatalf("Some error happened: %v", err)
	}

	tokenString, err := MakeJWT(id, tokenSecret, duration)
	if err != nil {
		t.Fatalf("Error happened during making: %v", err)
	}
	t.Logf("tokenString: %s", tokenString)
	idValidated, err := ValidateJWT(tokenString, tokenSecretWrong)
	if err == nil {
		t.Fatalf("Error with validation, expected to not validate: %v", err)
	}

	if idValidated.String() == id.String() {
		t.Fatalf("Expected not: %s but got: %s", id.String(), idValidated.String())
	}
}
func TestValidateJWTExpired(t *testing.T) {
	duration = time.Second
	id, err := uuid.Parse("123e4567-e89b-12d3-a456-426614174000")
	if err != nil {
		t.Fatalf("Some error happened: %v", err)
	}

	tokenString, err := MakeJWT(id, tokenSecret, duration)
	if err != nil {
		t.Fatalf("Error happened during making: %v", err)
	}
	t.Logf("tokenString: %s", tokenString)
	time.Sleep(2 * time.Second)
	idValidated, err := ValidateJWT(tokenString, tokenSecretWrong)
	if err == nil {
		t.Fatalf("Expected token to be expired, but it was validated: %v", err)
	}

	if idValidated.String() == id.String() {
		t.Fatalf("Expected not: %s but got: %s", id.String(), idValidated.String())
	}
}

func TestGetBearerTokenHappy(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c")

	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}

	if token != "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c" {
		t.Fatalf("Expected the word 'token' but got: %v", token)
	}
}

func TestGetBearerTokenNoHeader(t *testing.T) {
	headers := http.Header{}

	token, err := GetBearerToken(headers)
	if err == nil {
		t.Fatalf("Expected an error but returned: %v", err)
	}

	if token != "" {
		t.Fatalf("Expected empty string with error but got: %v", token)
	}
}
