// Package auth
package auth

import (
	"errors"
	"net/http"
	"strings"
)

// GetAPIKey extracts and api key from the header
// Authorization: ApiKey
func GetAPIKey(headers http.Header) (string, error) {
	val := headers.Get("Authorization")
	if val == "" {
		return "", errors.New("no auth info found")
	}

	vals := strings.Split(val, " ")
	if len(vals) != 2 {
		return "", errors.New("auth header not in correct form")
	}

	if vals[0] != "ApiKey" {
		return "", errors.New("incorrect first part of auth header")
	}

	return vals[1], nil
}

func GetAPIKeyFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("apikey")
	if err != nil {
		return "", err
	}

	apiKey := cookie.Value

	return apiKey, nil
}
