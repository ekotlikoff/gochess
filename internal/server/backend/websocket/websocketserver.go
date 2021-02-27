package websocketserver

import (
	"github.com/Ekotlikoff/gochess/internal/server/backend/match"
	"github.com/Ekotlikoff/gochess/internal/server/frontend"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

var upgrader = websocket.Upgrader{} // use default options

func Serve(
	matchServer *matchserver.MatchingServer, cache *gateway.TTLMap, port int,
	logFile *string, quiet bool,
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
	mux.Handle("/ws/matchandplay", makeMatchAndPlayHandler(matchServer, cache))
	http.ListenAndServe(":"+strconv.Itoa(port), mux)
}

func getSession(w http.ResponseWriter, r *http.Request, cache *gateway.TTLMap,
) *matchserver.Player {
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

func makeMatchAndPlayHandler(matchServer *matchserver.MatchingServer,
	cache *gateway.TTLMap) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.SetPrefix("runSession: ")
		player := getSession(w, r, cache)
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			w.WriteHeader(http.StatusInternalServerError)
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
		switch message.WebsocketRequestType {
		case matchserver.RequestSyncT:
			player.MakeMoveWS(message.RequestSync)
		case matchserver.RequestAsyncT:
			player.RequestAsync(message.RequestAsync)
		}
	}
}