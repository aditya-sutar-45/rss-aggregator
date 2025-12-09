package main

import (
	"fmt"
	"net/http"

	"github.com/aditya-sutar-45/rss-aggregator/internal/auth"
	"github.com/aditya-sutar-45/rss-aggregator/internal/database"
)

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

func (cfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := auth.GetAPIKeyFromCookie(r)
		if err != nil {
			respondWithErr(w, 403, fmt.Sprintf("no user found: %v\n", err))
			return
		}

		user, err := cfg.DB.GetUserByAPIKey(r.Context(), apiKey)
		if err != nil {
			respondWithErr(w, 404, fmt.Sprintf("User not found: %v\n", err))
			return
		}

		handler(w, r, user)
	}
}
