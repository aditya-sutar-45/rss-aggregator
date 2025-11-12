package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aditya-sutar-45/rss-aggregator/internal/database"
	"github.com/google/uuid"
)

func (apiCfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name string `json:"username"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithErr(w, 400, fmt.Sprintf("error parsing json: %v\n", err))
		return
	}

	user, err := apiCfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Username:  params.Name,
	})
	if err != nil {
		respondWithErr(w, 400, fmt.Sprintf("could not create user: %v\n", err))
		return
	}

	respondWithJSON(w, 200, databaseUserToUser(user))
}
