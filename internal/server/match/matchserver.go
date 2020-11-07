package matchserver

import (
	"gochess/internal/model"
	"sync"
)

type Player struct {
	name               string
	color              model.Color
	elapsedMs          int64
	requestChanSync    chan RequestSync
	responseChanSync   chan ResponseSync
	requestChanAsync   chan RequestAsync
	responseChanAsync  chan ResponseAsync
	opponentPlayedMove chan PieceMove
}

type PieceMove struct {
	Position model.Position
	Move     model.Move
}

func NewPlayer(name string) Player {
	return Player{
		name, model.Black, int64(0),
		make(chan RequestSync, 1), make(chan ResponseSync, 10),
		make(chan RequestAsync, 1), make(chan ResponseAsync, 1),
		make(chan PieceMove, 10),
	}
}

func (player *Player) Name() string {
	return player.name
}

func (player *Player) Color() model.Color {
	return player.color
}

func (player *Player) MakeMove(pieceMove PieceMove) bool {
	player.requestChanSync <- RequestSync{pieceMove.Position, pieceMove.Move}
	response := <-player.responseChanSync
	return response.moveSuccess
}

func (player *Player) GetOpponentMove() PieceMove {
	return <-player.opponentPlayedMove
}

func (player *Player) Stop() {
	close(player.requestChanSync)
	close(player.responseChanSync)
	close(player.requestChanAsync)
	close(player.responseChanAsync)
}

type RequestSync struct {
	position model.Position
	move     model.Move
}

type ResponseSync struct {
	moveSuccess bool
}

type RequestAsync struct {
	requestToDraw, resign bool
}

type ResponseAsync struct {
	gameOver, requestToDraw, draw, resignation, timeout bool
	winner                                              string
}

type MatchingServer struct {
	liveMatches []*Match
	mutex       *sync.Mutex
	players     chan *Player
}

func NewMatchingServer() MatchingServer {
	return MatchingServer{mutex: &sync.Mutex{}, players: make(chan *Player)}
}

func (matchingServer *MatchingServer) LiveMatches() []*Match {
	liveMatches := []*Match{}
	matchingServer.mutex.Lock()
	liveMatches = matchingServer.liveMatches
	matchingServer.mutex.Unlock()
	return liveMatches
}

func (matchingServer *MatchingServer) matchAndPlay(
	matchGenerator MatchGenerator,
) {
	var player1, player2 *Player
	for player := range matchingServer.players {
		if player1 == nil {
			player1 = player
		} else if player2 == nil {
			player2 = player
			match := matchGenerator(player1, player2)
			matchingServer.mutex.Lock()
			matchingServer.liveMatches = append(matchingServer.liveMatches, &match)
			matchingServer.mutex.Unlock()
			(&match).play()
			matchingServer.mutex.Lock()
			matchingServer.removeMatch(&match)
			matchingServer.mutex.Unlock()
			player1, player = nil, nil
		}
	}
}

func (matchingServer *MatchingServer) Serve(
	maxConcurrentGames int, quit chan bool,
) {
	matchingServer.ServeCustomMatch(
		maxConcurrentGames, DefaultMatchGenerator, quit,
	)
}

func (matchingServer *MatchingServer) ServeCustomMatch(
	maxConcurrentGames int, matchGenerator MatchGenerator, quit chan bool,
) {
	// Start handlers
	for i := 0; i < maxConcurrentGames; i++ {
		go matchingServer.matchAndPlay(matchGenerator)
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
