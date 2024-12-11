package auth

import (
	"crypto/rand"
	"encoding/hex"
	"log"
)

func MakeRefreshToken() (string, error) {
	b := make([]byte, 256)

	_, err := rand.Read(b)
	if err != nil {
		log.Fatalf("Error during make refresh token: %v", err)
	}
	h := hex.EncodeToString(b)

	return h, nil
}
