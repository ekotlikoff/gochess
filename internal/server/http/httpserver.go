package httpserver

// Credit to https://www.sohamkamani.com/blog/2018/03/25/golang-session-authentication/
import (
	//	"gochess/internal/server/match"
	"encoding/json"
	"fmt"
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

type httpHandler func(http.ResponseWriter, *http.Request)

func Serve(matchServer *matchserver.MatchingServer) {
	http.HandleFunc("/", StartSession)
	http.HandleFunc("/match/", createSearchForMatchHandler(matchServer))
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
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionTokenStr,
		Expires: time.Now().Add(1800 * time.Second),
	})
}

func createSearchForMatchHandler(
	matchServer *matchserver.MatchingServer,
) httpHandler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				// If the cookie is not set, return an unauthorized status
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// For any other type of error, return a bad request status
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sessionToken := c.Value
		player, err := cache.Get(sessionToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else if player == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		matchServer.MatchPlayer(player) // Block until matched
		fmt.Fprintf(w, "Matched!  Player color=\v", player.Color())
	}
	return handler
}
