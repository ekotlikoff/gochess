package httpserver

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/ekotlikoff/gochess/internal/model"
	matchserver "github.com/ekotlikoff/gochess/internal/server/backend/match"
	gateway "github.com/ekotlikoff/gochess/internal/server/frontend"
)

// HTTPBackend handles http requests
type HTTPBackend struct {
	MatchServer *matchserver.MatchingServer
	BasePath    string
	Port        int
}

// Serve the http server
func (backend *HTTPBackend) Serve() {
	bp := backend.BasePath
	if len(bp) > 0 && (bp[len(bp)-1:] == "/" || bp[0:1] != "/") {
		panic("Invalid gateway base path")
	}
	mux := http.NewServeMux()
	mux.Handle(bp+"/http/match", makeSearchForMatchHandler(backend.MatchServer))
	mux.Handle(bp+"/http/sync", makeSyncHandler())
	mux.Handle(bp+"/http/async", makeAsyncHandler())
	log.Println("HTTP server listening on port", backend.Port, "...")
	http.ListenAndServe(":"+strconv.Itoa(backend.Port), mux)
}

// SetQuiet logging
func SetQuiet() {
	log.SetOutput(ioutil.Discard)
}

func makeSearchForMatchHandler(
	matchServer *matchserver.MatchingServer,
) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		player := gateway.GetSession(w, r)
		if player == nil {
			return
		} else if !player.GetSearchingForMatch() {
			player.Reset()
			player.SetSearchingForMatch(true)
			matchServer.MatchPlayer(player)
		}
		ctx, cancel :=
			context.WithTimeout(context.Background(), matchserver.PollingDefaultTimeout)
		defer cancel()
		if player.HasMatchStarted(ctx) {
			player.SetSearchingForMatch(false)
		} else {
			// Return HTTP 202 until match starts.
			w.WriteHeader(http.StatusAccepted)
		}
	}
	return http.HandlerFunc(handler)
}

func makeSyncHandler() http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		player := gateway.GetSession(w, r)
		if player == nil {
			return
		}
		switch r.Method {
		case "GET":
			pieceMove := player.GetSyncUpdate()
			if pieceMove != nil {
				json.NewEncoder(w).Encode(*pieceMove)
				return
			}
			// Return HTTP 204 if no update.
			w.WriteHeader(http.StatusNoContent)
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
	return http.HandlerFunc(handler)
}

func makeAsyncHandler() http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		player := gateway.GetSession(w, r)
		if player == nil {
			return
		}
		switch r.Method {
		case "GET":
			asyncUpdate := player.GetAsyncUpdate()
			if asyncUpdate == nil {
				// Return HTTP 204 if no update.
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(asyncUpdate); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if asyncUpdate.GameOver {
				player.WaitForMatchOver()
				player.Reset()
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
	return http.HandlerFunc(handler)
}
