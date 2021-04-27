package httpserver

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/Ekotlikoff/gochess/internal/model"
	matchserver "github.com/Ekotlikoff/gochess/internal/server/backend/match"
	gateway "github.com/Ekotlikoff/gochess/internal/server/frontend"
)

// Serve the http server
func Serve(
	matchServer *matchserver.MatchingServer, port int,
) {
	mux := http.NewServeMux()
	mux.Handle("/http/match", makeSearchForMatchHandler(matchServer))
	mux.Handle("/http/sync", makeSyncHandler())
	mux.Handle("/http/async", makeAsyncHandler())
	log.Println("HTTP server listening on port", port, "...")
	http.ListenAndServe(":"+strconv.Itoa(port), mux)
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
			matchResponse :=
				matchserver.MatchedResponse{
					Color:        player.Color(),
					OpponentName: player.MatchedOpponentName(),
					MaxTimeMs:    player.MatchMaxTimeMs(),
				}
			json.NewEncoder(w).Encode(matchResponse)
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
				player.ClientDoneWithMatch()
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
