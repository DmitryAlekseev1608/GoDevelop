package main

import (
	"net/http"
	"github.com/gorilla/mux"
	"context"
	"log"
	"encoding/json"
	"time"
)

func GetApp() http.Handler {
	router := mux.NewRouter()
	router.Use(authMiddleware)
	router.HandleFunc("/api/users", startRegister).Methods(http.MethodPost)
	router.HandleFunc("/api/users/login", startLogin).Methods(http.MethodPost)
	router.HandleFunc("/api/user", startUser).Methods(http.MethodGet)
	return router
}

func startRegister(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var curUser user
	err := curUser.encodeUser(r)
	if err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}
	var userResp user
	userResp.User.Email = curUser.User.Email
	userResp.User.CreatedAt = time.Now().Format("2006-01-02T15:04:05Z07:00")
	userResp.User.UpdatedAt = time.Now().Format("2006-01-02T15:04:05Z07:00")
	userResp.User.Username = curUser.User.Username
	ctx := context.Background()
	jsonData := userResp.decodeUser()
	err = redisClient.HSet(ctx, "user", curUser.User.Email, jsonData).Err()
	if err != nil {
		log.Fatalf("Error when register client %s: %v", curUser.User.Email, err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) 
	json.NewEncoder(w).Encode(userResp)
}


func startLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var curUser user
	err := curUser.encodeUser(r)
	if err != nil {
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
		return
	}
	ctx := context.Background()
	dataAllUsersFromRadis, err := redisClient.HGetAll(ctx, "user").Result()
	if err != nil {
		log.Fatalf("Error when get data %s: %v", curUser.User.Email, err)
	}
	dataUserFromRedis, _ := dataAllUsersFromRadis[curUser.User.Email]
	var dataUserFromRedisStruct user
	err = json.Unmarshal([]byte(dataUserFromRedis), &dataUserFromRedisStruct)
	if err != nil {
		log.Fatalf("Error when unmarshal data of user %s: %v", curUser.User.Email, err)
	}
	var userResp user
	userResp.User.Email = dataUserFromRedisStruct.User.Email
	userResp.User.CreatedAt = dataUserFromRedisStruct.User.CreatedAt
	userResp.User.UpdatedAt = dataUserFromRedisStruct.User.UpdatedAt
	userResp.User.Username = dataUserFromRedisStruct.User.Username
	newSession := &session{}
	newSession.createSession(dataUserFromRedisStruct.User.Email)
	userResp.User.Token = newSession.sessionID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userResp)
}


func startUser(w http.ResponseWriter, r *http.Request) {
	newSession := &session{}
	newSession.SessionFromContext(r.Context())
	ctx := context.Background()
	dataAllSessionsFromRadis, err := redisClient.HGetAll(ctx, "session").Result()
	if err != nil {
		log.Fatalf("Error when get data %s: %v", newSession.sessionID, err)
	}
	curUser, _ := dataAllSessionsFromRadis[newSession.sessionID]
	dataAllUsersFromRadis, err := redisClient.HGetAll(ctx, "user").Result()
	if err != nil {
		log.Fatalf("Error when get data %s: %v", newSession.userID, err)
	}
	dataUserFromRedis, _ := dataAllUsersFromRadis[curUser]
	var dataUserFromRedisStruct user
	err = json.Unmarshal([]byte(dataUserFromRedis), &dataUserFromRedisStruct)
	if err != nil {
		log.Fatalf("Error when unmarshal data of user %s: %v", curUser, err)
	}
	var userResp user
	userResp.User.Email = dataUserFromRedisStruct.User.Email
	userResp.User.CreatedAt = dataUserFromRedisStruct.User.CreatedAt
	userResp.User.UpdatedAt = dataUserFromRedisStruct.User.UpdatedAt
	userResp.User.Username = dataUserFromRedisStruct.User.Username
	userResp.User.Token = r.Header.Get("Authorization")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userResp)
}
