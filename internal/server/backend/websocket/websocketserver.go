package websocketserver

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	matchserver "github.com/Ekotlikoff/gochess/internal/server/backend/match"
	gateway "github.com/Ekotlikoff/gochess/internal/server/frontend"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

// Serve the websocket server
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
	mux.Handle("/ws", makeWebsocketHandler(matchServer, cache))
	log.Println("WebsocketServer listening on port", port, "...")
	http.ListenAndServe(":"+strconv.Itoa(port), mux)
}

func makeWebsocketHandler(matchServer *matchserver.MatchingServer,
	cache *gateway.TTLMap) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.SetPrefix("WS Handler: ")
		player := getSession(w, r, cache)
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer c.Close()
		waitc := make(chan struct{})
		go readLoop(c, matchServer, player, waitc)
		writeLoop(c, player)
		<-waitc
		player.ClientDoneWithMatch()
	}
	return http.HandlerFunc(handler)
}

func writeLoop(c *websocket.Conn, player *matchserver.Player) {
	err := player.WaitForMatchStart()
	if err != nil {
		log.Println("FATAL: Failed to find match")
	}
	matchedResponse := matchserver.WebsocketResponse{
		WebsocketResponseType: matchserver.MatchStartT,
		MatchedResponse: matchserver.MatchedResponse{
			Color: player.Color(), OpponentName: player.MatchedOpponentName(),
			MaxTimeMs: player.MatchMaxTimeMs(),
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

func readLoop(c *websocket.Conn, matchServer *matchserver.MatchingServer,
	player *matchserver.Player, waitc chan struct{}) {
	defer c.Close()
	for {
		message := matchserver.WebsocketRequest{}
		if err := c.ReadJSON(&message); err != nil {
			log.Println("Read error:", err)
			close(waitc)
			return
		}
		switch message.WebsocketRequestType {
		case matchserver.RequestSyncT:
			player.MakeMoveWS(message.RequestSync)
		case matchserver.RequestAsyncT:
			if message.RequestAsync.Match {
				ctx, cancel := context.WithTimeout(
					context.Background(), 20*time.Millisecond)
				defer cancel()
				if !player.GetSearchingForMatch() &&
					!player.HasMatchStarted(ctx) {
					player.SetSearchingForMatch(true)
					matchServer.MatchPlayer(player)
					err := player.WaitForMatchStart()
					if err == nil {
						player.SetSearchingForMatch(false)
					} else {
						log.Println("FATAL: Failed to find match")
					}
				}
			} else {
				player.RequestAsync(message.RequestAsync)
			}
		}
	}
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
