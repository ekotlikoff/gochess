package websocketserver

import (
	"encoding/json"
	"github.com/Ekotlikoff/gochess/internal/server/match"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

var upgrader = websocket.Upgrader{} // use default options

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
	mux.HandleFunc("/session", runSession)
	http.ListenAndServe(":"+strconv.Itoa(port), mux)
}

func ServeRoot(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../webclient/assets/wasm_exec.html")
}

func runSession(w http.ResponseWriter, r *http.Request) {
	log.SetPrefix("runSession: ")
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
	if err != nil {
		log.Println("fatal: Failed to generate session token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer c.Close()
	player := matchserver.NewPlayer(creds.Username)
	err = player.WaitForMatchStart()
	if err != nil {
		log.Println("fatal: Failed to find match")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	go readLoop(c, player)
	go writeLoop(c, player)
}

func writeLoop(c *websocket.Conn, player matchserver.Player) {
	matchedResponse := WebsocketResponse{
		WebsocketResponseType: MatchStartT,
		MatchedResponse: MatchedResponse{
			player.Color(), player.MatchedOpponentName(),
			player.MatchMaxTimeMs(),
		},
	}
	c.WriteJSON(matchedResponse)
	for {
		if player.GetGameover() {
			return
		}
		message := player.GetWebsocketMessageToWrite()
		err := c.WriteJSON(*message)
		if err != nil {
			log.Println("Write error:", err)
		}
	}
}

func readLoop(c *websocket.Conn, player matchserver.Player) {
	for {
		if messageType, r, err := c.NextReader(); err != nil {
			defer c.Close()
			// TODO When match is over then return?
			message := WebsocketRequest{}
			c.ReadJSON(message)
			switch message.WebsocketRequestType {
			case RequestSyncT:
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
			case RequestAsyncT:
				//TODO
			}
		}
	}
}
