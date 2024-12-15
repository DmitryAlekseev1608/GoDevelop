package main


import (
	"net/http"
	"github.com/go-redis/redis/v8"
	"context"
	"log"
)


type contextKey string


const (
	sessionKey = 1
)


var (
	noAuthUrls = map[string]struct{} {
		"/api/users": struct{}{},
		"/api/users/login": struct{}{},
	}
	redisClient *redis.Client
)


type session struct {
	sessionID 	string
	userID		string
}


func (s *session) checkSession (w http.ResponseWriter, r *http.Request) {
	receivedSessionID := r.Header.Get("Authorization")
	ctx := context.Background()
	dataAllSessionsFromRadis, err := redisClient.HGetAll(ctx, "session").Result()
	if err != nil {
		log.Fatalf("Error when get session %s: %v", receivedSessionID, err)
	}
	_, ok := dataAllSessionsFromRadis[receivedSessionID]
	if !ok {
		log.Println("CheckSession no rows")
		s = nil
		http.Error(w, "Failed to decode JSON", http.StatusBadRequest)
	} else {
		s.sessionID = receivedSessionID
	}
}


func (s *session) createSession (email string) {
	tokens := map[string]string{
		"golang@example.com": "token1",
		"golang_second@example.com": "token2",		
	}
	ctx := context.Background()
	err := redisClient.HSet(ctx, "session", "Token " + tokens[email], email).Err()
	if err != nil {
		log.Fatalf("Error when create session %s: %v", tokens[email], err)
	}
	s.sessionID = tokens[email]
	s.userID = email
}


func (s *session) SessionFromContext(ctx context.Context) {
	receivedSessionID, ok := ctx.Value(sessionKey).(string)
	if !ok {
		s = nil
	} else {
		s.sessionID = receivedSessionID
	}
}


func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redisClient = redis.NewClient(&redis.Options{
			Addr: "localhost:6380",
			Username: "user",
			Password: "password",
	
		})
		if _, ok := noAuthUrls[r.URL.Path]; ok {
			next.ServeHTTP(w, r)
			return
		}
		curSession := &session{}
		curSession.checkSession(w, r)
		ctx := context.WithValue(r.Context(), sessionKey, curSession.sessionID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
