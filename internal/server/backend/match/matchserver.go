package matchserver

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	pb "github.com/Ekotlikoff/gochess/api"
	"github.com/Ekotlikoff/gochess/internal/model"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

var (
	// PollingDefaultTimeout is the default timeout for http requests
	PollingDefaultTimeout time.Duration = 10 * time.Second

	matchingServerID = 0
)

const (
	// NullT is the WS response type for a null response
	NullT = WebsocketResponseType(iota)
	// RequestSyncT is the WS request type for a sync request
	RequestSyncT = WebsocketRequestType(iota)
	// RequestAsyncT is the WS request type for an async request
	RequestAsyncT = WebsocketRequestType(iota)
	// ResponseSyncT is the WS response type for a sync response
	ResponseSyncT = WebsocketResponseType(iota)
	// ResponseAsyncT is the WS response type for an async response
	ResponseAsyncT = WebsocketResponseType(iota)
	// OpponentPlayedMoveT is the WS response type for an opponent's move
	OpponentPlayedMoveT = WebsocketResponseType(iota)
)

type (
	// WebsocketResponseType represents the different type of responses
	// supported over the WS conn
	WebsocketResponseType uint8

	// WebsocketResponse is a struct for a response over the WS conn
	WebsocketResponse struct {
		WebsocketResponseType WebsocketResponseType
		ResponseSync          ResponseSync
		ResponseAsync         ResponseAsync
		OpponentPlayedMove    model.MoveRequest
	}

	// WebsocketRequestType represents the different type of requests supported
	// over the WS conn
	WebsocketRequestType uint8

	// WebsocketRequest is a struct for a request over the WS conn
	WebsocketRequest struct {
		WebsocketRequestType WebsocketRequestType
		RequestSync          model.MoveRequest
		RequestAsync         RequestAsync
	}

	// MatchDetails is a struct for the matched response
	MatchDetails struct {
		Color        model.Color
		OpponentName string
		MaxTimeMs    int64
	}

	// Player is a struct representing a matchserver client, containing channels
	// for communications between the client and the the matchserver
	Player struct {
		name                string
		color               model.Color
		elapsedMs           int64
		requestChanSync     chan model.MoveRequest
		ResponseChanSync    chan ResponseSync
		RequestChanAsync    chan RequestAsync
		ResponseChanAsync   chan ResponseAsync
		OpponentPlayedMove  chan model.MoveRequest
		matchStart          chan struct{}
		clientDoneWithMatch chan struct{}
		ChannelMutex        sync.RWMutex
		matchStartMutex     sync.RWMutex
		matchMutex          sync.RWMutex
		searchingForMatch   bool
		match               *Match
	}
)

// NewPlayer create a new player
func NewPlayer(name string) *Player {
	player := Player{name: name, color: model.Black}
	player.Reset()
	return &player
}

// Name get the player's name
func (player *Player) Name() string {
	return player.name
}

// GetSearchingForMatch get searching for match
func (player *Player) GetSearchingForMatch() bool {
	player.matchMutex.RLock()
	defer player.matchMutex.RUnlock()
	return player.searchingForMatch
}

// SetSearchingForMatch set searching for match
func (player *Player) SetSearchingForMatch(searchingForMatch bool) {
	player.matchMutex.Lock()
	defer player.matchMutex.Unlock()
	player.searchingForMatch = searchingForMatch
}

// GetMatch get the player's match
func (player *Player) GetMatch() *Match {
	player.matchMutex.RLock()
	defer player.matchMutex.RUnlock()
	return player.match
}

// SetMatch set the player's match
func (player *Player) SetMatch(match *Match) {
	player.matchMutex.Lock()
	defer player.matchMutex.Unlock()
	player.match = match
}

// MatchedOpponentName returns the matched opponent name
func (player *Player) MatchedOpponentName() string {
	player.matchMutex.RLock()
	defer player.matchMutex.RUnlock()
	opponentColor := model.Black
	if player.color == opponentColor {
		opponentColor = model.White
	}
	return player.GetMatch().PlayerName(opponentColor)
}

// MatchMaxTimeMs returns players max time in ms
func (player *Player) MatchMaxTimeMs() int64 {
	return player.GetMatch().maxTimeMs
}

// Color returns player color
func (player *Player) Color() model.Color {
	player.matchMutex.RLock()
	defer player.matchMutex.RUnlock()
	return player.color
}

// WaitForMatchStart player waits for match start
func (player *Player) WaitForMatchStart() error {
	player.matchStartMutex.RLock()
	defer player.matchStartMutex.RUnlock()
	select {
	case <-player.matchStart:
		return nil
	case <-time.After(120 * time.Second):
		return errors.New("Timeout")
	}
}

// HasMatchStarted player checks if their match has started
func (player *Player) HasMatchStarted(ctx context.Context) bool {
	player.matchStartMutex.RLock()
	defer player.matchStartMutex.RUnlock()
	select {
	case <-player.matchStart:
		return true
	case <-ctx.Done():
		return false
	}
}

// MakeMoveWS player (websocket client) make a move
func (player *Player) MakeMoveWS(pieceMove model.MoveRequest) {
	player.ChannelMutex.RLock()
	defer player.ChannelMutex.RUnlock()
	player.requestChanSync <- pieceMove
}

// MakeMove player makes a move
func (player *Player) MakeMove(pieceMove model.MoveRequest) bool {
	player.ChannelMutex.RLock()
	defer player.ChannelMutex.RUnlock()
	player.requestChanSync <- pieceMove
	response := <-player.ResponseChanSync
	return response.MoveSuccess
}

// GetSyncUpdate get the next sync update for a player
func (player *Player) GetSyncUpdate() *model.MoveRequest {
	player.ChannelMutex.RLock()
	defer player.ChannelMutex.RUnlock()
	select {
	case update := <-player.OpponentPlayedMove:
		return &update
	case <-time.After(PollingDefaultTimeout):
		return nil
	}
}

// RequestAsync player makes an async request
func (player *Player) RequestAsync(requestAsync RequestAsync) {
	player.ChannelMutex.RLock()
	defer player.ChannelMutex.RUnlock()
	player.RequestChanAsync <- requestAsync
}

// GetAsyncUpdate get the next async update for a player
func (player *Player) GetAsyncUpdate() *ResponseAsync {
	player.ChannelMutex.RLock()
	defer player.ChannelMutex.RUnlock()
	select {
	case update := <-player.ResponseChanAsync:
		return &update
	case <-time.After(PollingDefaultTimeout):
		return nil
	}
}

// Reset a player for their next match
func (player *Player) Reset() {
	player.ChannelMutex.Lock()
	defer player.ChannelMutex.Unlock()
	// If the player preexisted there may be a client waiting on the opponent's
	// move.
	if player.OpponentPlayedMove != nil {
		close(player.OpponentPlayedMove)
	}
	player.elapsedMs = 0
	player.SetMatch(nil)
	player.requestChanSync = make(chan model.MoveRequest, 1)
	player.ResponseChanSync = make(chan ResponseSync, 10)
	player.RequestChanAsync = make(chan RequestAsync, 1)
	player.ResponseChanAsync = make(chan ResponseAsync, 1)
	player.OpponentPlayedMove = make(chan model.MoveRequest, 10)
	player.matchStartMutex.Lock()
	defer player.matchStartMutex.Unlock()
	player.matchStart = make(chan struct{})
}

func (player *Player) startMatch() {
	player.ChannelMutex.Lock()
	defer player.ChannelMutex.Unlock()
	player.clientDoneWithMatch = make(chan struct{})
	player.matchStartMutex.RLock()
	defer player.matchStartMutex.RUnlock()
	close(player.matchStart)
	player.ResponseChanAsync <- ResponseAsync{
		Matched: true,
		MatchDetails: MatchDetails{
			Color:        player.Color(),
			OpponentName: player.MatchedOpponentName(),
			MaxTimeMs:    player.MatchMaxTimeMs(),
		},
	}
}

// ClientDoneWithMatch the client is now done with the match
func (player *Player) ClientDoneWithMatch() {
	player.ChannelMutex.RLock()
	defer player.ChannelMutex.RUnlock()
	close(player.clientDoneWithMatch)
}

// WaitForClientToBeDoneWithMatch block until the client is done with their
// match
func (player *Player) WaitForClientToBeDoneWithMatch() {
	player.ChannelMutex.RLock()
	defer player.ChannelMutex.RUnlock()
	<-player.clientDoneWithMatch
}

// ResponseSync represents a response to the client related to a move
type ResponseSync struct {
	MoveSuccess       bool
	ElapsedMs         int
	ElapsedMsOpponent int
}

// RequestAsync represents a request from the client unrelated to a move
type RequestAsync struct {
	Match, RequestToDraw, Resign bool
}

// ResponseAsync represents a response to the client unrelated to a move
type ResponseAsync struct {
	GameOver, RequestToDraw, Draw, Resignation, Timeout, Matched bool
	Winner                                                       string
	MatchDetails                                                 MatchDetails
}

// MatchingServer handles matching players and carrying out the game
type MatchingServer struct {
	id                        int
	liveMatches               []*Match
	liveMatchesMetric         prometheus.Gauge
	matchingQueueLengthMetric prometheus.Gauge
	mutex                     *sync.Mutex
	matchingPlayers           chan *Player
	pendingMatch              *sync.Mutex
	botMatchingEnabled        bool
	engineClient              pb.RustChessClient
	engineClientConn          *grpc.ClientConn
	maxMatchingDuration       time.Duration
}

// NewMatchingServer create a matching server with no engine
func NewMatchingServer() MatchingServer {
	matchingServer := MatchingServer{
		id: matchingServerID, mutex: &sync.Mutex{},
		matchingPlayers: make(chan *Player), pendingMatch: &sync.Mutex{},
	}
	matchingServerID++
	matchingQueueLengthMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "gochess",
		Subsystem: "matchserver",
		Name:      "matching_queue_length",
		Help:      "The number of players in the matching queue.",
		ConstLabels: prometheus.Labels{
			"matching_server_id": strconv.Itoa(matchingServer.id),
		},
	})
	prometheus.MustRegister(matchingQueueLengthMetric)
	matchingServer.matchingQueueLengthMetric = matchingQueueLengthMetric
	return matchingServer
}

// NewMatchingServerWithEngine create a matching server with an engine
func NewMatchingServerWithEngine(
	engineAddr string, maxMatchingDuration time.Duration,
	engineConnTimeout time.Duration,
) MatchingServer {
	matchingServer := NewMatchingServer()
	matchingServer.createEngineClient(engineAddr, engineConnTimeout)
	matchingServer.maxMatchingDuration = maxMatchingDuration
	return matchingServer
}

// LiveMatches current matches being played
func (matchingServer *MatchingServer) LiveMatches() []*Match {
	matchingServer.mutex.Lock()
	liveMatches := matchingServer.liveMatches
	matchingServer.mutex.Unlock()
	return liveMatches
}

func (matchingServer *MatchingServer) matchAndPlay(
	matchGenerator MatchGenerator, playServerID int,
) {
	var player1, player2 *Player
	maxMatchingTimer := time.NewTimer(0)
	<-maxMatchingTimer.C
	// Lock until a full match is found and started, thus avoiding unmatched
	// players stranded across goroutines.
	matchingServer.pendingMatch.Lock()
	for {
		select {
		case player := <-matchingServer.matchingPlayers:
			if player1 == nil {
				player1 = player
				if matchingServer.botMatchingEnabled {
					maxMatchingTimer.Reset(matchingServer.maxMatchingDuration)
				}
			} else if player2 == nil {
				player2 = player
				match := matchGenerator(player1, player2)
				matchingServer.matchingQueueLengthMetric.Sub(2)
				player1.SetMatch(&match)
				player2.SetMatch(&match)
				matchingServer.mutex.Lock()
				matchingServer.liveMatches =
					append(matchingServer.liveMatches, &match)
				matchingServer.mutex.Unlock()
				matchingServer.pendingMatch.Unlock()
				player1.startMatch()
				player2.startMatch()
				matchingServer.liveMatchesMetric.Inc()
				(&match).play()
				matchingServer.liveMatchesMetric.Dec()
				matchingServer.removeMatch(&match)
				player1, player2 = nil, nil
				matchingServer.pendingMatch.Lock()
			}
		case <-maxMatchingTimer.C:
			// The maxMatchingTimer has fired and we should match player1 with a
			// bot.
			botNames := [5]string{
				"jessica", "cherry", "gumdrop", "roland", "pumpkin",
			}
			botPlayer := NewPlayer(botNames[rand.Intn(len(botNames))] + "bot")
			go matchingServer.engineSession(botPlayer)
			matchingServer.matchingQueueLengthMetric.Inc()
			go (func() { matchingServer.matchingPlayers <- botPlayer })()
		}
	}
}

// StartMatchServers using default match generator
func (matchingServer *MatchingServer) StartMatchServers(
	maxConcurrentGames int, quit chan bool,
) {
	matchingServer.StartCustomMatchServers(
		maxConcurrentGames, DefaultMatchGenerator, quit,
	)
}

// StartCustomMatchServers using custom match generator
func (matchingServer *MatchingServer) StartCustomMatchServers(
	maxConcurrentGames int, matchGenerator MatchGenerator, quit chan bool,
) {
	matchingServer.liveMatchesMetric = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "gochess",
		Subsystem: "matchserver",
		Name:      "live_matches",
		Help:      "The number of live matches",
		ConstLabels: prometheus.Labels{
			"matching_server_id": strconv.Itoa(matchingServer.id),
		},
	})
	prometheus.MustRegister(matchingServer.liveMatchesMetric)
	log.Printf("Starting %d matchAndPlay threads ...", maxConcurrentGames)
	for i := 0; i < maxConcurrentGames; i++ {
		go matchingServer.matchAndPlay(matchGenerator, i)
	}
	<-quit // Wait to be told to exit.
	if matchingServer.engineClientConn != nil {
		matchingServer.engineClientConn.Close()
	}
}

func (matchingServer *MatchingServer) removeMatch(matchToRemove *Match) {
	liveMatches := matchingServer.liveMatches
	matchingServer.mutex.Lock()
	defer matchingServer.mutex.Unlock()
	for i, match := range liveMatches {
		if match == matchToRemove {
			if len(liveMatches) == 1 {
				matchingServer.liveMatches = nil
			} else {
				liveMatches[i] = liveMatches[len(liveMatches)-1]
				matchingServer.liveMatches = liveMatches[:len(liveMatches)-1]
				return
			}
		}
	}
}

// MatchPlayer queues the player for matching
func (matchingServer *MatchingServer) MatchPlayer(player *Player) {
	matchingServer.matchingPlayers <- player
	matchingServer.matchingQueueLengthMetric.Inc()
}
