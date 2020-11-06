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

func Serve() {
	//		http.HandleFunc()
}

func CreateStartSessionHandler(matchServer matchserver.Match) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {

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
		// TODO kick off a goroutine where we start matching, and once matched
		// send a response to the client that they've matched and what their
		// color is.
		//  matchServer.MatchPlayer(player)
		// Set the client session_token and an expiry time equal to the cache ttl
		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   sessionTokenStr,
			Expires: time.Now().Add(1800 * time.Second),
		})
	}
	return handler
}

func Welcome(w http.ResponseWriter, r *http.Request) {
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

	// We then get the name of the user from our cache, where we set the session token
	player, err := cache.Get(sessionToken)
	if err != nil {
		// If there is an error fetching from cache, return an internal server error status
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if player == nil {
		// If the session token is not present in cache, return an unauthorized error
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// Finally, return the welcome message to the user
	w.Write([]byte(fmt.Sprintf("Welcome %s!", player.Name())))
}
