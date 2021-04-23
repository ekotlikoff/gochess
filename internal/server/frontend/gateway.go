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

	matchserver "github.com/Ekotlikoff/gochess/internal/server/backend/match"
	"github.com/gofrs/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// SessionCache is the cache of all user sessions
	SessionCache *TTLMap

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
	SessionCache = NewTTLMap(50, 1800, 10)
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
	log.SetPrefix("StartSession: ")
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
	// Create a new random session token
	sessionToken, err := uuid.NewV4()
	if err != nil {
		log.Println("Failed to generate session token")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	sessionTokenStr := sessionToken.String()
	player := matchserver.NewPlayer(creds.Username)
	SessionCache.Put(sessionTokenStr, player)
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionTokenStr,
		Expires: time.Now().Add(1800 * time.Second),
	})
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
