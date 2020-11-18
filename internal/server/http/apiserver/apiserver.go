package apiserver

// Credit to https://www.sohamkamani.com/blog/2018/03/25/golang-session-authentication/
import (
	"encoding/json"
	"fmt"
	"github.com/satori/go.uuid"
	"gochess/internal/model"
	"gochess/internal/server/match"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var cache *TTLMap

func init() {
	cache = NewTTLMap(50, 1800, 10)
}

type Credentials struct {
	Username string
}

func Serve(
	matchServer *matchserver.MatchingServer, port int, logFile *string,
	quiet bool,
) {
	if logFile != nil {
		file, err := os.OpenFile(*logFile, os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatal(err)
		}
		log.SetOutput(file)
	}
	if quiet {
		log.SetOutput(ioutil.Discard)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", StartSession)
	mux.Handle("/match", createSearchForMatchHandler(matchServer))
	mux.HandleFunc("/sync", SyncHandler)
	mux.HandleFunc("/async", AsyncHandler)
	http.ListenAndServe(":"+strconv.Itoa(port), mux)
}

func StartSession(w http.ResponseWriter, r *http.Request) {
	log.SetPrefix("StartSession: ")
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		log.Println("Bad request", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	} else if creds.Username == "" {
		log.Println("Missing username")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing username"))
		return
	}
	// Create a new random session token
	sessionToken, err := uuid.NewV4()
	if err != nil {
		log.Println("Failed to generate session token")
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
) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.SetPrefix("SearchForMatch: ")
		player := getSession(w, r)
		if player == nil {
			return
		}
		matchServer.MatchPlayer(player)
		player.WaitForMatchStart()
		fmt.Fprintf(w, "Matched!  Player color=%v", player.Color())
	}
	return http.HandlerFunc(handler)
}

func SyncHandler(w http.ResponseWriter, r *http.Request) {
	log.SetPrefix("SyncHandler: ")
	player := getSession(w, r)
	if player == nil {
		return
	}
	switch r.Method {
	case "GET":
		pieceMove := player.GetSyncUpdate()
		fmt.Fprintf(w, "Opponent move=%v", pieceMove)
	case "POST":
		var moveRequest model.MoveRequest
		err := json.NewDecoder(r.Body).Decode(&moveRequest)
		if err != nil {
			log.Println("Failed to parse move body ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		success := player.MakeMove(moveRequest)
		if !success {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

func AsyncHandler(w http.ResponseWriter, r *http.Request) {
	log.SetPrefix("AsyncHandler: ")
	player := getSession(w, r)
	if player == nil {
		return
	}
	switch r.Method {
	case "GET":
		asyncUpdate := player.GetAsyncUpdate()
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(asyncUpdate); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	case "POST":
		var requestAsync matchserver.RequestAsync
		err := json.NewDecoder(r.Body).Decode(&requestAsync)
		if err != nil {
			log.Println("Bad request", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		player.RequestAsync(requestAsync)
	}
}

func getSession(w http.ResponseWriter, r *http.Request) *matchserver.Player {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			log.Println("session_token is not set")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Missing session_token"))
			return nil
		}
		log.Println("ERROR ", err)
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	sessionToken := c.Value
	player, err := cache.Get(sessionToken)
	if err != nil {
		log.Println("ERROR ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	} else if player == nil {
		log.Println("No player found for token ", sessionToken)
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}
	return player
}
