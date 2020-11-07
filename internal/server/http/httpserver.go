package httpserver

// Credit to https://www.sohamkamani.com/blog/2018/03/25/golang-session-authentication/
import (
	//	"gochess/internal/server/match"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/satori/go.uuid"
	"gochess/internal/model"
	"gochess/internal/server/match"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var cache *TTLMap

func init() {
	cache = NewTTLMap(50, 1800, 10)
}

type Credentials struct {
	Username string `json:"username"`
}

func Serve(matchServer *matchserver.MatchingServer) {
	http.HandleFunc("/", StartSession)
	http.Handle("/match/", createSearchForMatchHandler(matchServer))
	http.HandleFunc("/sync/", moveHandler)
	http.ListenAndServe(":80", nil)
}

func StartSession(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Starting session")
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil || creds.Username == "" {
		fmt.Println("Failed to decode the body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Create a new random session token
	sessionToken, err := uuid.NewV4()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	sessionTokenStr := sessionToken.String()
	player := matchserver.NewPlayer(creds.Username)
	fmt.Println(player.Name())
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
		c, err := r.Cookie("session_token")
		if err != nil {
			if err == http.ErrNoCookie {
				// If the cookie is not set, return an unauthorized status
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// For any other type of error, return a bad request status
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sessionToken := c.Value
		player, err := cache.Get(sessionToken)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else if player == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		matchServer.MatchPlayer(player) // Block until matched
		fmt.Fprintf(w, "Matched!  Player color=%v", player.Color())
	}
	return http.HandlerFunc(handler)
}

func moveHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// If the cookie is not set, return an unauthorized status
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// For any other type of error, return a bad request status
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sessionToken := c.Value
	player, err := cache.Get(sessionToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if player == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	// Read body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	pieceMove, err := parseMoveBody(body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	success := player.MakeMove(pieceMove)
	if !success {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func parseMoveBody(body []byte) (matchserver.PieceMove, error) {
	// Match "(3,1)(0,2)" as PieceMove{Position{3,1}, Move{0,2}}
	re := regexp.MustCompile(`^\((\d?),(\d?)\)\((\d?),(\d?)\)$`)
	match := re.FindAllString(string(body), 4)
	if len(match) != 4 {
		return matchserver.PieceMove{}, errors.New("Failed to parse body")
	}
	posX, errX := strconv.ParseUint(match[0], 10, 8)
	posY, errY := strconv.ParseUint(match[1], 10, 8)
	moveX, errMX := strconv.ParseInt(match[2], 10, 8)
	moveY, errMY := strconv.ParseInt(match[3], 10, 8)
	if errX != nil || errY != nil || errMX != nil || errMY != nil {
		return matchserver.PieceMove{}, errors.New("Failed to parse ints")
	}
	return matchserver.PieceMove{
		model.Position{uint8(posX), uint8(posY)},
		model.Move{int8(moveX), int8(moveY)}}, nil
}
