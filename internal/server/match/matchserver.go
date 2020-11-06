package matchserver

import (
	"gochess/internal/model"
	"sync"
)

type Player struct {
	name              string
	color             model.Color
	elapsedMs         int64
	requestChanSync   chan RequestSync
	responseChanSync  chan ResponseSync
	requestChanAsync  chan RequestAsync
	responseChanAsync chan ResponseAsync
}

func NewPlayer(name string) Player {
	return Player{
		name, model.Black, int64(0),
		make(chan RequestSync), make(chan ResponseSync),
		make(chan RequestAsync), make(chan ResponseAsync),
	}
}

func (player *Player) Name() string {
	return player.name
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
