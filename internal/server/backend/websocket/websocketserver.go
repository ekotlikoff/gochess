package websocketserver

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	matchserver "github.com/ekotlikoff/gochess/internal/server/backend/match"
	gateway "github.com/ekotlikoff/gochess/internal/server/frontend"
	"github.com/gorilla/websocket"
	opentracing "github.com/opentracing/opentracing-go"
	opentracinglog "github.com/opentracing/opentracing-go/log"
)

var upgrader = websocket.Upgrader{} // use default options

const (
	// Time allowed to write a message to the peer. TODO use this for write deadline as per gorilla ws examples.
	//writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 5 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 7) / 10
)

// WSBackend handles websocket connections
type WSBackend struct {
	MatchServer *matchserver.MatchingServer
	BasePath    string
	Port        int
}

// Serve the websocket server
func (backend *WSBackend) Serve() {
	bp := backend.BasePath
	if len(bp) > 0 && (bp[len(bp)-1:] == "/" || bp[0:1] != "/") {
		panic("Invalid gateway base path")
	}
	mux := http.NewServeMux()
	mux.Handle(bp+"/ws", makeWebsocketHandler(backend.MatchServer))
	log.Println("WebsocketServer listening on port", backend.Port, "...")
	http.ListenAndServe(":"+strconv.Itoa(backend.Port), mux)
}

func makeWebsocketHandler(matchServer *matchserver.MatchingServer,
) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		tracer := opentracing.GlobalTracer()
		wsHandlerSpan := tracer.StartSpan("WebsocketConnection")
		defer wsHandlerSpan.Finish()
		player := gateway.GetSession(w, r)
		if player == nil {
			log.Println("No player found for session")
			return
		}
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade error:", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer c.Close()
		waitc := make(chan struct{})
		player.ClientConnectToPlayer()
		defer player.ClientDisconnectFromPlayer()
		// playerMutex is used to ensure we are not interacting with the player
		// while resetting it
		playerMutex := &sync.Mutex{}
		go readLoop(c, matchServer, player, wsHandlerSpan, waitc, playerMutex)
		writeLoop(c, player, wsHandlerSpan, playerMutex)
		<-waitc
		log.Println("Websocketserver disconnecting from client: " + player.Name())
	}
	return http.HandlerFunc(handler)
}

func writeLoop(c *websocket.Conn, player *matchserver.Player,
	span opentracing.Span, playerMutex *sync.Mutex) {
	tracer := opentracing.GlobalTracer()
	err := player.WaitForMatchStart()
	if err != nil {
		log.Println("FATAL: Failed to find match")
		return
	}
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	for {
		getResSpan := tracer.StartSpan(
			"GetResponse",
			opentracing.ChildOf(span.Context()),
		)
		resType := ""
		var response matchserver.WebsocketResponse
		select {
		case ResponseSync := <-player.ResponseChanSync:
			resType = "sync"
			response = matchserver.WebsocketResponse{
				WebsocketResponseType: matchserver.ResponseSyncT,
				ResponseSync:          ResponseSync,
			}
		case ResponseAsync := <-player.ResponseChanAsync:
			resType = "async"
			response = matchserver.WebsocketResponse{
				WebsocketResponseType: matchserver.ResponseAsyncT,
				ResponseAsync:         ResponseAsync,
			}
		case OpponentMove := <-player.OpponentPlayedMove:
			resType = "opponentPlayedMove"
			response = matchserver.WebsocketResponse{
				WebsocketResponseType: matchserver.OpponentPlayedMoveT,
				OpponentPlayedMove:    OpponentMove,
			}
		case <-ticker.C:
			getResSpan.LogFields(opentracinglog.String("resType", "ping"))
			if err := c.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("FATAL Write PingMessage error:", err)
				getResSpan.Finish()
				return
			}
			getResSpan.Finish()
			continue
		}
		getResSpan.LogFields(opentracinglog.String("resType", resType))
		getResSpan.Finish()
		err := c.WriteJSON(response)
		if err != nil {
			log.Println("Write error:", err)
			return
		} else if response.WebsocketResponseType == matchserver.ResponseAsyncT &&
			response.ResponseAsync.GameOver {
			playerMutex.Lock()
			// The biggest potential delay here is the matchserver waiting up till
			// timeout for a misbehaving client to read its gameover async response
			// (match.go handleGameOver).  During this delay we are not writing ping
			// messages, so if the matchserver waits too long for other clients this
			// ws connection could die.
			player.WaitForMatchOver()
			player.Reset()
			playerMutex.Unlock()
		}
	}
}

func readLoop(c *websocket.Conn, matchServer *matchserver.MatchingServer,
	player *matchserver.Player, span opentracing.Span, waitc chan struct{},
	playerMutex *sync.Mutex) {
	defer c.Close()
	tracer := opentracing.GlobalTracer()
	for {
		message := matchserver.WebsocketRequest{}
		readWSSpan := tracer.StartSpan(
			"ReadWS",
			opentracing.ChildOf(span.Context()),
		)
		if err := c.ReadJSON(&message); err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Websocketserver read error: %v", err)
			}
			/* TODO instead of immediate resignation let the matchserver know
			that there is no client connected, then matchserver can handle when
			to time them out
			if player.GetMatch() != nil && !player.GetMatch().GameOver() {
				player.RequestChanAsync <- matchserver.RequestAsync{
					Resign: true,
				}
			}
			*/
			close(waitc)
			readWSSpan.LogFields(opentracinglog.String("readType", "ConnClosed"))
			readWSSpan.Finish()
			return
		}
		readWSSpan.LogFields(opentracinglog.String("readType", "MessageReceived"))
		readWSSpan.Finish()
		switch message.WebsocketRequestType {
		case matchserver.RequestSyncT:
			playerMutex.Lock()
			makeMoveSpan := tracer.StartSpan(
				"MakeMove",
				opentracing.ChildOf(span.Context()),
			)
			player.MakeMoveWS(message.RequestSync)
			makeMoveSpan.Finish()
			playerMutex.Unlock()
		case matchserver.RequestAsyncT:
			playerMutex.Lock()
			if message.RequestAsync.Match {
				ctx, cancel := context.WithTimeout(
					context.Background(), 20*time.Millisecond)
				defer cancel()
				if !player.GetSearchingForMatch() &&
					!player.HasMatchStarted(ctx) {
					player.SetSearchingForMatch(true)
					matchServer.MatchPlayer(player)
					waitForMatchSpan := tracer.StartSpan(
						"WaitForMatchStart",
						opentracing.ChildOf(span.Context()),
					)
					err := player.WaitForMatchStart()
					waitForMatchSpan.Finish()
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
				requestAsyncSpan := tracer.StartSpan(
					"RequestAsync",
					opentracing.ChildOf(span.Context()),
				)
				player.RequestAsync(message.RequestAsync)
				requestAsyncSpan.Finish()
			}
			playerMutex.Unlock()
		}
	}
}
