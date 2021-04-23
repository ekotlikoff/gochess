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
	matchServer *matchserver.MatchingServer, cache *gateway.TTLMap, port int,
) {
	mux := http.NewServeMux()
	mux.Handle("/http/match", makeSearchForMatchHandler(matchServer, cache))
	mux.Handle("/http/sync", makeSyncHandler(cache))
	mux.Handle("/http/async", makeAsyncHandler(cache))
	log.Println("HTTP server listening on port", port, "...")
	http.ListenAndServe(":"+strconv.Itoa(port), mux)
}

// SetQuiet logging
func SetQuiet() {
	log.SetOutput(ioutil.Discard)
}

func makeSearchForMatchHandler(
	matchServer *matchserver.MatchingServer, cache *gateway.TTLMap,
) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.SetPrefix("SearchForMatch: ")
		player := getSession(w, r, cache)
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

func makeSyncHandler(cache *gateway.TTLMap) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.SetPrefix("SyncHandler: ")
		player := getSession(w, r, cache)
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

func makeAsyncHandler(cache *gateway.TTLMap) http.Handler {
	handler := func(w http.ResponseWriter, r *http.Request) {
		log.SetPrefix("AsyncHandler: ")
		player := getSession(w, r, cache)
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
