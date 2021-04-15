package matchserver

import (
	"context"
	"errors"
	pb "github.com/Ekotlikoff/gochess/api"
	"github.com/Ekotlikoff/gochess/internal/model"
	"google.golang.org/grpc"
	"math/rand"
	"sync"
	"time"
)

var PollingDefaultTimeout time.Duration = 10 * time.Second

const (
	NullT               = WebsocketResponseType(iota)
	MatchStartT         = WebsocketResponseType(iota)
	RequestSyncT        = WebsocketRequestType(iota)
	RequestAsyncT       = WebsocketRequestType(iota)
	ResponseSyncT       = WebsocketResponseType(iota)
	ResponseAsyncT      = WebsocketResponseType(iota)
	OpponentPlayedMoveT = WebsocketResponseType(iota)
)

type (
	WebsocketResponseType uint8

	WebsocketResponse struct {
		WebsocketResponseType WebsocketResponseType
		MatchedResponse       MatchedResponse
		ResponseSync          ResponseSync
		ResponseAsync         ResponseAsync
		OpponentPlayedMove    model.MoveRequest
	}

	WebsocketRequestType uint8

	WebsocketRequest struct {
		WebsocketRequestType WebsocketRequestType
		RequestSync          model.MoveRequest
		RequestAsync         RequestAsync
	}

	MatchedResponse struct {
		Color        model.Color
		OpponentName string
		MaxTimeMs    int64
	}

	Player struct {
		name                string
		color               model.Color
		elapsedMs           int64
		requestChanSync     chan model.MoveRequest
		responseChanSync    chan ResponseSync
		requestChanAsync    chan RequestAsync
		responseChanAsync   chan ResponseAsync
		opponentPlayedMove  chan model.MoveRequest
		matchStart          chan struct{}
		clientDoneWithMatch chan struct{}
		channelMutex        sync.RWMutex
		matchStartMutex     sync.RWMutex
		matchMutex          sync.RWMutex
		searchingForMatch   bool
		match               *Match
	}
)

func NewPlayer(name string) *Player {
	player := Player{name: name, color: model.Black}
	player.Reset()
	return &player
}

func (player *Player) Name() string {
	return player.name
}

func (player *Player) GetSearchingForMatch() bool {
	player.matchMutex.RLock()
	defer player.matchMutex.RUnlock()
	return player.searchingForMatch
}

func (player *Player) SetSearchingForMatch(searchingForMatch bool) {
	player.matchMutex.Lock()
	defer player.matchMutex.Unlock()
	player.searchingForMatch = searchingForMatch
}

func (player *Player) GetMatch() *Match {
	player.matchMutex.RLock()
	defer player.matchMutex.RUnlock()
	return player.match
}

func (player *Player) SetMatch(match *Match) {
	player.matchMutex.Lock()
	defer player.matchMutex.Unlock()
	player.match = match
}

func (player *Player) MatchedOpponentName() string {
	player.matchMutex.RLock()
	defer player.matchMutex.RUnlock()
	opponentColor := model.Black
	if player.color == opponentColor {
		opponentColor = model.White
	}
	return player.GetMatch().PlayerName(opponentColor)
}

func (player *Player) MatchMaxTimeMs() int64 {
	return player.GetMatch().MaxTimeMs()
}

func (player *Player) Color() model.Color {
	return player.color
}

func (player *Player) GetWebsocketMessageToWrite() *WebsocketResponse {
	player.channelMutex.RLock()
	defer player.channelMutex.RUnlock()
	var response WebsocketResponse
	select {
	case responseSync := <-player.responseChanSync:
		response = WebsocketResponse{
			WebsocketResponseType: ResponseSyncT,
			ResponseSync:          responseSync,
		}
	case responseAsync := <-player.responseChanAsync:
		response = WebsocketResponse{
			WebsocketResponseType: ResponseAsyncT,
			ResponseAsync:         responseAsync,
		}
	case opponentMove := <-player.opponentPlayedMove:
		response = WebsocketResponse{
			WebsocketResponseType: OpponentPlayedMoveT,
			OpponentPlayedMove:    opponentMove,
		}
	}
	return &response
}

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

func (player *Player) MakeMoveWS(pieceMove model.MoveRequest) {
	player.channelMutex.RLock()
	defer player.channelMutex.RUnlock()
	player.requestChanSync <- pieceMove
}

func (player *Player) MakeMove(pieceMove model.MoveRequest) bool {
	player.channelMutex.RLock()
	defer player.channelMutex.RUnlock()
	player.requestChanSync <- pieceMove
	response := <-player.responseChanSync
	return response.MoveSuccess
}

func (player *Player) GetSyncUpdate() *model.MoveRequest {
	player.channelMutex.RLock()
	defer player.channelMutex.RUnlock()
	select {
	case update := <-player.opponentPlayedMove:
		return &update
	case <-time.After(PollingDefaultTimeout):
		return nil
	}
}

func (player *Player) RequestAsync(requestAsync RequestAsync) {
	player.channelMutex.RLock()
	defer player.channelMutex.RUnlock()
	player.requestChanAsync <- requestAsync
}

func (player *Player) GetAsyncUpdate() *ResponseAsync {
	player.channelMutex.RLock()
	defer player.channelMutex.RUnlock()
	select {
	case update := <-player.responseChanAsync:
		return &update
	case <-time.After(PollingDefaultTimeout):
		return nil
	}
}

func (player *Player) Reset() {
	player.channelMutex.Lock()
	defer player.channelMutex.Unlock()
	// If the player preexisted there may be a client waiting on the opponent's
	// move.
	if player.opponentPlayedMove != nil {
		close(player.opponentPlayedMove)
	}
	player.elapsedMs = 0
	player.SetMatch(nil)
	player.requestChanSync = make(chan model.MoveRequest, 1)
	player.responseChanSync = make(chan ResponseSync, 10)
	player.requestChanAsync = make(chan RequestAsync, 1)
	player.responseChanAsync = make(chan ResponseAsync, 1)
	player.opponentPlayedMove = make(chan model.MoveRequest, 10)
	player.matchStartMutex.Lock()
	defer player.matchStartMutex.Unlock()
	player.matchStart = make(chan struct{})
}

func (player *Player) startMatch() {
	player.matchStartMutex.RLock()
	defer player.matchStartMutex.RUnlock()
	close(player.matchStart)
	player.channelMutex.Lock()
	defer player.channelMutex.Unlock()
	player.clientDoneWithMatch = make(chan struct{})
}

func (player *Player) ClientDoneWithMatch() {
	player.channelMutex.RLock()
	defer player.channelMutex.RUnlock()
	close(player.clientDoneWithMatch)
}

func (player *Player) WaitForClientToBeDoneWithMatch() {
	player.channelMutex.RLock()
	defer player.channelMutex.RUnlock()
	<-player.clientDoneWithMatch
}

type ResponseSync struct {
	MoveSuccess bool
}

type RequestAsync struct {
	Match, RequestToDraw, Resign bool
}

type ResponseAsync struct {
	GameOver, RequestToDraw, Draw, Resignation, Timeout bool
	Winner                                              string
}

type MatchingServer struct {
	liveMatches         []*Match
	mutex               *sync.Mutex
	players             chan *Player
	pendingMatch        *sync.Mutex
	botMatchingEnabled  bool
	engineClient        pb.RustChessClient
	engineClientConn    *grpc.ClientConn
	maxMatchingDuration time.Duration
}

func NewMatchingServer() MatchingServer {
	matchingServer := MatchingServer{
		mutex: &sync.Mutex{}, players: make(chan *Player),
		pendingMatch: &sync.Mutex{},
	}
	return matchingServer
}

func NewMatchingServerWithEngine(
	engineAddr string, maxMatchingDuration time.Duration,
	engineConnTimeout time.Duration,
) MatchingServer {
	matchingServer := MatchingServer{
		mutex: &sync.Mutex{}, players: make(chan *Player),
		pendingMatch: &sync.Mutex{},
	}
	matchingServer.createEngineClient(engineAddr, engineConnTimeout)
	matchingServer.maxMatchingDuration = maxMatchingDuration
	return matchingServer
}

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
		case player := <-matchingServer.players:
			if player1 == nil {
				player1 = player
				if matchingServer.botMatchingEnabled {
					maxMatchingTimer.Reset(matchingServer.maxMatchingDuration)
				}
			} else if player2 == nil {
				player2 = player
				match := matchGenerator(player1, player2)
				player1.SetMatch(&match)
				player2.SetMatch(&match)
				matchingServer.mutex.Lock()
				matchingServer.liveMatches =
					append(matchingServer.liveMatches, &match)
				matchingServer.mutex.Unlock()
				matchingServer.pendingMatch.Unlock()
				player1.startMatch()
				player2.startMatch()
				(&match).play()
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
			go (func() { matchingServer.players <- botPlayer })()
		}
	}
}

func (matchingServer *MatchingServer) StartMatchServers(
	maxConcurrentGames int, quit chan bool,
) {
	matchingServer.StartCustomMatchServers(
		maxConcurrentGames, DefaultMatchGenerator, quit,
	)
}

func (matchingServer *MatchingServer) StartCustomMatchServers(
	maxConcurrentGames int, matchGenerator MatchGenerator, quit chan bool,
) {
	// Start handlers
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

func (matchingServer *MatchingServer) MatchPlayer(player *Player) {
	matchingServer.players <- player
}
