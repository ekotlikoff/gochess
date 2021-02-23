package websocketserver

import (
	"encoding/json"
	"github.com/Ekotlikoff/gochess/internal/server/api/cache"
	"github.com/Ekotlikoff/gochess/internal/server/match"
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	sessionCache *cache.TTLMap
	upgrader     = websocket.Upgrader{} // use default options
)

func init() {
	sessionCache = cache.NewTTLMap(50, 1800, 10)
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

	mux.Handle("/", http.FileServer(http.Dir("./cmd/webserver/assets")))
	mux.HandleFunc("/session", StartSession)
	mux.Handle("/matchandplay", createMatchAndPlayHandler(matchServer))
	http.ListenAndServe(":"+strconv.Itoa(port), mux)
}

func ServeRoot(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../webclient/assets/wasm_exec.html")
}

// Credit to https://www.sohamkamani.com/blog/2018/03/25/golang-session-authentication/
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
	sessionCache.Put(sessionTokenStr, &player)
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionTokenStr,
		Expires: time.Now().Add(1800 * time.Second),
	})
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
	player, err := sessionCache.Get(sessionToken)
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

func createMatchAndPlayHandler(matchServer *matchserver.MatchingServer,
) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.SetPrefix("runSession: ")
		player := getSession(w, r)
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			return
		}
		defer c.Close()
		player.SetSearchingForMatch(true)
		matchServer.MatchPlayer(player)
		err = player.WaitForMatchStart()
		if err != nil {
			log.Println("fatal: Failed to find match")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		go readLoop(c, player)
		writeLoop(c, player)
	}
	return http.HandlerFunc(handler)
}

func writeLoop(c *websocket.Conn, player *matchserver.Player) {
	matchedResponse := matchserver.WebsocketResponse{
		WebsocketResponseType: matchserver.MatchStartT,
		MatchedResponse: matchserver.MatchedResponse{
			player.Color(), player.MatchedOpponentName(),
			player.MatchMaxTimeMs(),
		},
	}
	c.WriteJSON(&matchedResponse)
	for {
		if player.GetGameover() {
			return
		}
		if message := player.GetWebsocketMessageToWrite(); message != nil {
			err := c.WriteJSON(message)
			if err != nil {
				log.Println("Write error:", err)
				return
			}
		}
	}
}

func readLoop(c *websocket.Conn, player *matchserver.Player) {
	defer c.Close()
	for {
		message := matchserver.WebsocketRequest{}
		if err := c.ReadJSON(&message); err != nil {
			log.Println("Read error:", err)
			return
		}
		// TODO When match is over then return?
		switch message.WebsocketRequestType {
		case matchserver.RequestSyncT:
			player.MakeMoveWS(message.RequestSync)
		case matchserver.RequestAsyncT:
			//TODO
		}
	}
}
