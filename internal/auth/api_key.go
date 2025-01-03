package auth

import (
	"errors"
	"net/http"
)

func GetAPIKey(header http.Header) (string, error) {
	key := header.Get("Authorization")
	if key == "" {
		return "", errors.New("no api key")
	}
	k := key[7:]

	return k, nil
}
