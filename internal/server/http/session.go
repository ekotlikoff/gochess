package httpserver

// Credit to https://www.sohamkamani.com/blog/2018/03/25/golang-session-authentication/

import (
	"encoding/json"
	"github.com/satori/go.uuid"
	"gochess/internal/server/match"
	"net/http"
	"time"
)

var cache *TTLMap

func init() {
	cache = NewTTLMap(50, 1800, 10)
}

type Credentials struct {
	Username string `json:"username"`
}

func StartSession(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Create a new random session token
	sessionToken, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	sessionTokenStr := sessionToken.String()
	player := matchserver.NewPlayer(creds.Username)
	cache.Put(sessionTokenStr, &player)

	// TODO start matchmaking before returning 200 to client.
	// Need a match server to start matching on
	// 1) Create a Server() func for the http server.  This func will
	// take in a matching server as an arg
	// 2) Rename and wrap StartSession in a closure where we pass in the match
	// server and use http.HandleFunc to serve this initial endpoint.

	// Set the client session_token and an expiry time equal to the cache ttl
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionTokenStr,
		Expires: time.Now().Add(1800 * time.Second),
	})
}
