package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aditya-sutar-45/rss-aggregator/internal/auth"
	"github.com/aditya-sutar-45/rss-aggregator/internal/database"
	"github.com/google/uuid"
)

func (apiCfg *apiConfig) handlerRegister(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithErr(w, 400, fmt.Sprintf("error parsing json: %v\n", err))
		return
	}

	username := params.Username
	password := params.Password
	if username == "" || password == "" {
		respondWithErr(w, 400, fmt.Sprintf("username / password is empty: %v\n", err))
		return
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		respondWithErr(w, 500, fmt.Sprintf("something went worng!: %v\n", err))
		return
	}

	user, err := apiCfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:           uuid.New(),
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		Username:     username,
		PasswordHash: hashedPassword,
	})
	if err != nil {
		respondWithErr(w, 400, fmt.Sprintf("could not create user: %v\n", err))
		return
	}

	respondWithJSON(w, 201, databaseUserToUser(user))
}

func (apiCfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithErr(w, 400, fmt.Sprintf("error parsing json: %v\n", err))
		return
	}

	username := params.Username
	password := params.Password
	if username == "" || password == "" {
		respondWithErr(w, 400, fmt.Sprintf("username / password is empty: %v\n", err))
		return
	}

	user, err := apiCfg.DB.GetUserByUsername(r.Context(), username)
	if err != nil {
		respondWithErr(w, 404, "user not found")
		return
	}

	if !auth.CheckPasswordHash(password, user.PasswordHash) {
		respondWithErr(w, http.StatusUnauthorized, "incorrect username / password")
		return
	}

	cookie := http.Cookie{
		Name:     "apikey",
		Value:    user.ApiKey,
		Expires:  time.Now().Add(time.Hour * 24), // 1 day
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	respondWithJSON(w, 200, databaseUserToUser(user))
}

func (apiCfg *apiConfig) handlerLogout(w http.ResponseWriter, r *http.Request) {
	cookie := http.Cookie{
		Name:     "apikey",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie)

	type Response struct {
		Message string `json:"message"`
	}
	res := Response{
		Message: "logout success",
	}
	respondWithJSON(w, 200, res)
}

func (apiCfg *apiConfig) handlerCheckAuth(w http.ResponseWriter, r *http.Request, u database.User) {
	type Response struct {
		Message string `json:"message"`
	}
	res := Response{
		Message: "user logged in",
	}
	respondWithJSON(w, 200, res)
}

/*
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

	respondWithJSON(w, 201, databaseUserToUser(user))
}
*/

// authenticated endpoint
func (apiCfg *apiConfig) handlerGetUser(w http.ResponseWriter, r *http.Request, user database.User) {
	respondWithJSON(w, 200, databaseUserToUser(user))
}

func (apiConfig *apiConfig) handlerGetPostsForUser(w http.ResponseWriter, r *http.Request, user database.User) {
	posts, err := apiConfig.DB.GetPostsForUser(r.Context(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  10,
	})
	if err != nil {
		respondWithErr(w, 400, fmt.Sprintf("Could not get posts: %v", err))
		return
	}
	log.Printf("got %v posts\n", len(posts))

	respondWithJSON(w, 200, databasePostsToPosts(posts))
}
