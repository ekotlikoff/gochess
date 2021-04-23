package websocketserver

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	matchserver "github.com/Ekotlikoff/gochess/internal/server/backend/match"
	gateway "github.com/Ekotlikoff/gochess/internal/server/frontend"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

const (
	// Time allowed to write a message to the peer. TODO use this for write deadline as per gorilla ws examples.
	//writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 5 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

// Serve the websocket server
func Serve(
	matchServer *matchserver.MatchingServer, cache *gateway.TTLMap, port int,
) {
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
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		var response matchserver.WebsocketResponse
		//player.ChannelMutex.RLock()
		select {
		case ResponseSync := <-player.ResponseChanSync:
			response = matchserver.WebsocketResponse{
				WebsocketResponseType: matchserver.ResponseSyncT,
				ResponseSync:          ResponseSync,
			}
		case ResponseAsync := <-player.ResponseChanAsync:
			response = matchserver.WebsocketResponse{
				WebsocketResponseType: matchserver.ResponseAsyncT,
				ResponseAsync:         ResponseAsync,
			}
		case OpponentMove := <-player.OpponentPlayedMove:
			response = matchserver.WebsocketResponse{
				WebsocketResponseType: matchserver.OpponentPlayedMoveT,
				OpponentPlayedMove:    OpponentMove,
			}
		case <-ticker.C:
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				//player.ChannelMutex.RUnlock()
				return
			}
			continue
		}
		//player.ChannelMutex.RUnlock()
		err := c.WriteJSON(response)
		if err != nil {
			log.Println("Write error:", err)
			return
		} else if response.WebsocketResponseType == matchserver.ResponseAsyncT &&
			response.ResponseAsync.GameOver {
			return
		}
	}
}

func readLoop(c *websocket.Conn, matchServer *matchserver.MatchingServer,
	player *matchserver.Player, waitc chan struct{}) {
	defer c.Close()
	for {
		message := matchserver.WebsocketRequest{}
		if player.GetMatch() != nil && player.GetMatch().GameOver() {
			return
		}
		if err := c.ReadJSON(&message); err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Websocketserver read error: %v", err)
			}
			if player.GetMatch() != nil && !player.GetMatch().GameOver() {
				player.RequestChanAsync <- matchserver.RequestAsync{
					Resign: true,
				}
			}
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
						c.SetReadDeadline(time.Now().Add(pongWait))
						c.SetPongHandler(func(string) error {
							c.SetReadDeadline(time.Now().Add(pongWait))
							return nil
						})
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
