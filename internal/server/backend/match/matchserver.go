package matchserver

import (
	"errors"
	"github.com/Ekotlikoff/gochess/internal/model"
	"sync"
	"time"
)

var DefaultTimeout time.Duration = 10 * time.Second

const (
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
		name               string
		color              model.Color
		elapsedMs          int64
		requestChanSync    chan model.MoveRequest
		responseChanSync   chan ResponseSync
		requestChanAsync   chan RequestAsync
		responseChanAsync  chan ResponseAsync
		opponentPlayedMove chan model.MoveRequest
		matchStart         chan struct{}
		matchMutex         sync.RWMutex
		searchingForMatch  bool
		match              *Match
	}
)

func NewPlayer(name string) Player {
	player := Player{name: name, color: model.Black}
	player.Reset()
	return player
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

func (player *Player) GetGameover() bool {
	player.matchMutex.RLock()
	defer player.matchMutex.RUnlock()
	select {
	case <-player.match.gameOver:
		return true
	default:
		return false
	}
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
	default:
		return nil
	}
	return &response
}

func (player *Player) WaitForMatchStart() error {
	select {
	case <-player.matchStart:
		return nil
	case <-time.After(120 * time.Second):
		return errors.New("Timeout")
	}
}

func (player *Player) HasMatchStarted() bool {
	select {
	case <-player.matchStart:
		return true
	case <-time.After(DefaultTimeout):
		return false
	}
}

func (player *Player) MakeMoveWS(pieceMove model.MoveRequest) {
	player.requestChanSync <- pieceMove
}

func (player *Player) MakeMove(pieceMove model.MoveRequest) bool {
	player.requestChanSync <- pieceMove
	response := <-player.responseChanSync
	return response.MoveSuccess
}

func (player *Player) GetSyncUpdate() *model.MoveRequest {
	select {
	case update := <-player.opponentPlayedMove:
		return &update
	case <-time.After(DefaultTimeout):
		return nil
	}
}

func (player *Player) RequestAsync(requestAsync RequestAsync) {
	player.requestChanAsync <- requestAsync
}

func (player *Player) GetAsyncUpdate() *ResponseAsync {
	select {
	case update := <-player.responseChanAsync:
		return &update
	case <-time.After(DefaultTimeout):
		return nil
	}
}

func (player *Player) Reset() {
	// If the player preexisted there may be a client waiting on the opponent's
	// move.
	if player.opponentPlayedMove != nil {
		close(player.opponentPlayedMove)
	}
	player.elapsedMs = 0
	player.requestChanSync = make(chan model.MoveRequest, 1)
	player.responseChanSync = make(chan ResponseSync, 10)
	player.requestChanAsync = make(chan RequestAsync, 1)
	player.responseChanAsync = make(chan ResponseAsync, 1)
	player.opponentPlayedMove = make(chan model.MoveRequest, 10)
	player.matchStart = make(chan struct{})
}

type ResponseSync struct {
	MoveSuccess bool
}

type RequestAsync struct {
	RequestToDraw, Resign bool
}

type ResponseAsync struct {
	GameOver, RequestToDraw, Draw, Resignation, Timeout bool
	Winner                                              string
}

type MatchingServer struct {
	liveMatches  []*Match
	mutex        *sync.Mutex
	players      chan *Player
	pendingMatch *sync.Mutex
}

func NewMatchingServer() MatchingServer {
	return MatchingServer{
		mutex: &sync.Mutex{}, players: make(chan *Player),
		pendingMatch: &sync.Mutex{},
	}
}

func (matchingServer *MatchingServer) LiveMatches() []*Match {
	liveMatches := []*Match{}
	matchingServer.mutex.Lock()
	liveMatches = matchingServer.liveMatches
	matchingServer.mutex.Unlock()
	return liveMatches
}

func (matchingServer *MatchingServer) matchAndPlay(
	matchGenerator MatchGenerator, playServerID int,
) {
	var player1, player2 *Player
	// Lock until a full match is found and started, thus avoiding unmatched
	// players stranded across goroutines.
	matchingServer.pendingMatch.Lock()
	for player := range matchingServer.players {
		if player1 == nil {
			player1 = player
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
			close(player1.matchStart)
			close(player2.matchStart)
			(&match).play()
			matchingServer.removeMatch(&match)
			player1.SetMatch(nil)
			player2.SetMatch(nil)
			player1, player2 = nil, nil
			matchingServer.pendingMatch.Lock()
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
				liveMatches = liveMatches[:len(liveMatches)-1]
				return
			}
		}
	}
}

func (matchingServer *MatchingServer) MatchPlayer(player *Player) {
	matchingServer.players <- player
}
