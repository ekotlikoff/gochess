package gateway

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"time"

	model "github.com/Ekotlikoff/gochess/internal/model"
	matchserver "github.com/Ekotlikoff/gochess/internal/server/backend/match"
	"github.com/gofrs/uuid"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	sessionCache *TTLMap

	//go:embed static
	webStaticFS embed.FS

	gatewayResponseMetric = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gochess",
			Subsystem: "gateway",
			Name:      "gateway_request_total",
			Help:      "Total number of requests serviced.",
		},
		[]string{"uri", "method", "status"},
	)

	gatewayResponseDurationMetric = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "gochess",
			Subsystem: "gateway",
			Name:      "gateway_request_duration",
			Help:      "Duration of requests serviced.",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1,
				2.5, 5, 10},
		},
		[]string{"uri", "method", "status"},
	)
)

func init() {
	sessionCache = NewTTLMap(50, 1800, 10)
	prometheus.MustRegister(gatewayResponseMetric)
	prometheus.MustRegister(gatewayResponseDurationMetric)
}

// Credentials are the credentialss for authentication
type Credentials struct {
	Username string
}

// Serve static files and proxy to the different backends
func Serve(httpBackend *url.URL, websocketBackend *url.URL, port int) {
	httpBackendProxy := httputil.NewSingleHostReverseProxy(httpBackend)
	wsBackendProxy := httputil.NewSingleHostReverseProxy(websocketBackend)
	wsBackendProxy.ModifyResponse = func(res *http.Response) error {
		gatewayResponseMetric.WithLabelValues(
			res.Request.URL.Path, res.Request.Method, res.Status).Inc()
		return nil
	}
	mux := http.NewServeMux()
	mux.Handle("/", prometheusMiddleware(http.HandlerFunc(handleWebRoot)))
	mux.Handle("/session", prometheusMiddleware(http.HandlerFunc(StartSession)))
	mux.Handle("/currentmatch", prometheusMiddleware(
		http.HandlerFunc(GetCurrentMatch)))
	// HTTP backend proxying
	mux.Handle("/http/match", prometheusMiddleware(httpBackendProxy))
	mux.Handle("/http/sync", prometheusMiddleware(httpBackendProxy))
	mux.Handle("/http/async", prometheusMiddleware(httpBackendProxy))
	// Websocket backend proxying
	mux.Handle("/ws", wsBackendProxy)
	// Prometheus metrics endpoint
	mux.Handle("/metrics", prometheusMiddleware(
		promhttp.Handler()))
	log.Println("Gateway server listening on port", port, "...")
	http.ListenAndServe(":"+strconv.Itoa(port), mux)
}

func handleWebRoot(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = "/static" + r.URL.Path // This is a hack to get the embedded path
	http.FileServer(http.FS(webStaticFS)).ServeHTTP(w, r)
}

// StartSession credit to https://www.sohamkamani.com/blog/2018/03/25/golang-session-authentication/
func StartSession(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	startSessionSpan := tracer.StartSpan("StartSession")
	defer startSessionSpan.Finish()
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
	newTokenSpan := tracer.StartSpan(
		"NewToken",
		opentracing.ChildOf(startSessionSpan.Context()),
	)
	sessionToken, err := uuid.NewV4()
	newTokenSpan.Finish()
	if err != nil {
		log.Println("Failed to generate session token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	sessionTokenStr := sessionToken.String()
	newPlayerSpan := tracer.StartSpan(
		"NewPlayer",
		opentracing.ChildOf(startSessionSpan.Context()),
	)
	player := matchserver.NewPlayer(creds.Username)
	newPlayerSpan.Finish()
	sessionCache.Put(sessionTokenStr, player)
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionTokenStr,
		Expires: time.Now().Add(1800 * time.Second),
	})
	w.WriteHeader(http.StatusOK)
}

// GetSession credit to https://www.sohamkamani.com/blog/2018/03/25/golang-session-authentication/
func GetSession(w http.ResponseWriter, r *http.Request) *matchserver.Player {
	tracer := opentracing.GlobalTracer()
	getSessionSpan := tracer.StartSpan("GetSession")
	defer getSessionSpan.Finish()
	c, err := r.Cookie("session_token")
	if err != nil {
		println("GET SESSION ERROR")
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
	println("SESSION token: " + sessionToken)
	getTokenSpan := tracer.StartSpan(
		"GetToken",
		opentracing.ChildOf(getSessionSpan.Context()),
	)
	player, err := sessionCache.Get(sessionToken)
	getTokenSpan.Finish()
	if err != nil {
		println("SESSION CACHE ERROR")
		log.Println("ERROR ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	} else if player == nil {
		println("PLAYER IS NIL")
		log.Println("No player found for token ", sessionToken)
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}
	return player
}

type CurrentMatch struct {
	BlackName                   string
	WhiteName                   string
	BlackRemainingTimeMs        int64
	WhiteRemainingTimeMs        int64
	Board                       model.Board
	Turn                        model.Color
	GameOver                    bool
	Result                      model.GameResult
	PreviousMove                model.Move
	PreviousMover               model.Piece
	BlackKing                   model.Piece
	WhiteKing                   model.Piece
	WhitePieces                 map[model.PieceType]uint8
	BlackPieces                 map[model.PieceType]uint8
	PositionHistory             map[string]uint8
	TurnsSinceCaptureOrPawnMove uint8
	RequestedDraw               bool
	RequestedDrawName           string
}

func currentMatchFromMatch(match *matchserver.Match) CurrentMatch {
	requestedDraw := match.GetRequestedDraw() != nil
	requestedDrawName := ""
	if requestedDraw {
		requestedDrawName = match.GetRequestedDraw().Name()
	}
	return CurrentMatch{
		BlackName:                   match.PlayerName(model.Black),
		WhiteName:                   match.PlayerName(model.White),
		BlackRemainingTimeMs:        match.PlayerRemainingTimeMs(model.Black),
		WhiteRemainingTimeMs:        match.PlayerRemainingTimeMs(model.White),
		Board:                       *match.Game.GetBoard(),
		Turn:                        match.Game.Turn(),
		GameOver:                    match.Game.GameOver(),
		Result:                      match.Game.Result(),
		PreviousMove:                match.Game.PreviousMove(),
		PreviousMover:               *match.Game.PreviousMover(),
		BlackKing:                   *match.Game.BlackKing(),
		WhiteKing:                   *match.Game.WhiteKing(),
		WhitePieces:                 match.Game.WhitePieces(),
		BlackPieces:                 match.Game.BlackPieces(),
		PositionHistory:             match.Game.PositionHistory(),
		TurnsSinceCaptureOrPawnMove: match.Game.TurnsSinceCaptureOrPawnMove(),
		RequestedDraw:               requestedDraw,
		RequestedDrawName:           requestedDrawName,
	}
}

type CurrentMatchResponse struct {
	Credentials Credentials
	Match       CurrentMatch
}

// GetCurrentMatch returns the client's username if a valid token is supplied, it
// also checks for a match the player is playing, if one exists it is returned.
func GetCurrentMatch(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	currentMatchSpan := tracer.StartSpan("CurrentMatch")
	player := GetSession(w, r)
	currentMatchResponse := CurrentMatchResponse{}
	if player == nil {
		return
	} else if player.GetMatch() == nil {
		currentMatchResponse = CurrentMatchResponse{
			Credentials: Credentials{Username: player.Name()},
		}
	} else {
		currentMatchResponse = CurrentMatchResponse{
			Credentials: Credentials{Username: player.Name()},
			Match:       currentMatchFromMatch(player.GetMatch()),
		}
	}
	if err := json.NewEncoder(w).Encode(currentMatchResponse); err != nil {
		fmt.Println(err)
		log.Println(err)
	}
	defer currentMatchSpan.Finish()
	w.WriteHeader(http.StatusOK)
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader and record the status for instrumentation
func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// prometheusMiddleware handles the request by passing it to the real
// handler and creating time series with the request details
func prometheusMiddleware(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := statusWriter{ResponseWriter: w}
		handler.ServeHTTP(&sw, r)
		duration := time.Since(start)
		gatewayResponseMetric.WithLabelValues(
			r.URL.Path, r.Method, fmt.Sprintf("%d", sw.status)).Inc()
		gatewayResponseDurationMetric.WithLabelValues(r.URL.Path, r.Method,
			fmt.Sprintf("%d", sw.status)).Observe(duration.Seconds())
	}
}

// SetQuiet logging
func SetQuiet() {
	log.SetOutput(ioutil.Discard)
}
