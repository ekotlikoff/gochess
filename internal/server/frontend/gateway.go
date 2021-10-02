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
	"os"
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

type (
	// Gateway is the server that serves static files and proxies to the different
	// backends
	Gateway struct {
		HTTPBackend *url.URL
		WSBackend   *url.URL
		BasePath    string
		Port        int
	}

	// Credentials for authentication
	Credentials struct {
		Username string
	}
)

// Serve static files and proxy to the different backends
func (gw *Gateway) Serve() {
	httpBackendProxy := httputil.NewSingleHostReverseProxy(gw.HTTPBackend)
	wsBackendProxy := httputil.NewSingleHostReverseProxy(gw.WSBackend)
	wsBackendProxy.ModifyResponse = func(res *http.Response) error {
		gatewayResponseMetric.WithLabelValues(
			res.Request.URL.Path, res.Request.Method, res.Status).Inc()
		return nil
	}
	mux := http.NewServeMux()
	bp := gw.BasePath
	if len(bp) > 0 && (bp[len(bp)-1:] == "/" || bp[0:1] != "/") {
		panic("Invalid gateway base path")
	}
	mux.Handle(bp+"/", prometheusMiddleware(http.HandlerFunc(gw.handleWebRoot)))
	mux.Handle(bp+"/gochessclient.wasm", prometheusMiddleware(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, os.Getenv("HOME")+"/bin/gochessclient.wasm")
		})))
	mux.Handle(bp+"/session", prometheusMiddleware(http.HandlerFunc(Session)))
	// HTTP backend proxying
	mux.Handle(bp+"/http/match", prometheusMiddleware(httpBackendProxy))
	mux.Handle(bp+"/http/sync", prometheusMiddleware(httpBackendProxy))
	mux.Handle(bp+"/http/async", prometheusMiddleware(httpBackendProxy))
	// Websocket backend proxying
	mux.Handle(bp+"/ws", wsBackendProxy)
	// Prometheus metrics endpoint
	mux.Handle(bp+"/metrics", prometheusMiddleware(
		promhttp.Handler()))
	log.Println("Gateway server listening on port", gw.Port, "...")
	http.ListenAndServe(":"+strconv.Itoa(gw.Port), mux)
}

func (gw *Gateway) handleWebRoot(w http.ResponseWriter, r *http.Request) {
	bp := gw.BasePath
	if len(bp) > 0 && len(r.URL.Path) > len(bp) && r.URL.Path[0:len(bp)] == bp {
		r.URL.Path = "/static" + r.URL.Path[len(bp):]
	} else {
		r.URL.Path = "/static" + r.URL.Path // This is a hack to get the embedded path
	}
	http.FileServer(http.FS(webStaticFS)).ServeHTTP(w, r)
}

// Session credit to https://www.sohamkamani.com/blog/2018/03/25/golang-session-authentication/
func Session(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		getSession(w, r)
	} else if r.Method == http.MethodPost {
		newSession(w, r)
	}
}

func getSession(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	SessionSpan := tracer.StartSpan("GETSession")
	defer SessionSpan.Finish()
	player := GetSession(w, r)
	currentMatchResponse := SessionResponse{}
	if player == nil {
		return
	} else if player.GetMatch() == nil {
		log.Println("Found session,", player.Name())
		currentMatchResponse = SessionResponse{
			Credentials: Credentials{Username: player.Name()},
		}
	} else {
		currentMatchResponse = SessionResponse{
			Credentials: Credentials{Username: player.Name()},
			InMatch:     true,
			Match:       currentMatchFromMatch(player.GetMatch()),
		}
	}
	if err := json.NewEncoder(w).Encode(currentMatchResponse); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func newSession(w http.ResponseWriter, r *http.Request) {
	tracer := opentracing.GlobalTracer()
	SessionSpan := tracer.StartSpan("POSTSession")
	defer SessionSpan.Finish()
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
		opentracing.ChildOf(SessionSpan.Context()),
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
		opentracing.ChildOf(SessionSpan.Context()),
	)
	player := matchserver.NewPlayer(creds.Username)
	newPlayerSpan.Finish()
	log.Println("Adding to sessionCache,", creds.Username)
	err = sessionCache.Put(sessionTokenStr, player)
	if err != nil {
		log.Println("Failed to store session token in sessionCache")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionTokenStr,
		Expires: time.Now().Add(1800 * time.Second),
	})
}

// GetSession credit to https://www.sohamkamani.com/blog/2018/03/25/golang-session-authentication/
func GetSession(w http.ResponseWriter, r *http.Request) *matchserver.Player {
	tracer := opentracing.GlobalTracer()
	getSessionSpan := tracer.StartSpan("GetSession")
	defer getSessionSpan.Finish()
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			log.Println("session_token is not set")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Missing session_token"))
			return nil
		}
		log.Println("ERROR", err)
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	sessionToken := c.Value
	getTokenSpan := tracer.StartSpan(
		"GetToken",
		opentracing.ChildOf(getSessionSpan.Context()),
	)
	player, err := sessionCache.Get(sessionToken)
	getTokenSpan.Finish()
	if err != nil {
		log.Println("ERROR token is invalid")
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	} else if player == nil {
		log.Println("No player found for token ", sessionToken)
		w.WriteHeader(http.StatusUnauthorized)
		return nil
	}
	return player
}

// CurrentMatch serializable struct to bring client up to speed
type CurrentMatch struct {
	BlackName                   string
	WhiteName                   string
	MaxTimeMs                   int64
	BlackRemainingTimeMs        int64
	WhiteRemainingTimeMs        int64
	Board                       model.SerializableBoard
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

// SessionResponse serializable struct to send client's session
type SessionResponse struct {
	Credentials Credentials
	InMatch     bool
	Match       CurrentMatch
}

func currentMatchFromMatch(match *matchserver.Match) CurrentMatch {
	requestedDraw := match.GetRequestedDraw() != nil
	requestedDrawName := ""
	if requestedDraw {
		requestedDrawName = match.GetRequestedDraw().Name()
	}
	previousMoverPtr := match.Game.PreviousMover()
	previousMover := model.Piece{}
	if previousMoverPtr != nil {
		previousMover = *previousMoverPtr
	}
	return CurrentMatch{
		BlackName:                   match.PlayerName(model.Black),
		WhiteName:                   match.PlayerName(model.White),
		MaxTimeMs:                   match.MaxTimeMs(),
		BlackRemainingTimeMs:        match.PlayerRemainingTimeMs(model.Black),
		WhiteRemainingTimeMs:        match.PlayerRemainingTimeMs(model.White),
		Board:                       match.Game.GetSerializableBoard(),
		Turn:                        match.Game.Turn(),
		GameOver:                    match.Game.GameOver(),
		Result:                      match.Game.Result(),
		PreviousMove:                match.Game.PreviousMove(),
		PreviousMover:               previousMover,
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
