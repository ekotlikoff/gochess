package server

import (
	"gochess/internal/model"
	"sync"
)

const maxConcurrentGames = 10

type Player struct {
	name         string
	color        model.Color
	elapsedMs    int64
	requestChan  <-chan Request
	responseChan chan<- Response
}

type Request struct {
	position      model.Position
	move          model.Move
	requestToDraw bool
	resign        bool
}

type Response struct {
	success  bool
	gameOver bool
	winner   string
}

type MatchingServer struct {
	liveMatches []*Match
	mutex       *sync.Mutex
}

func NewMatchingServer() MatchingServer {
	return MatchingServer{nil, &sync.Mutex{}}
}

var matchingChan = make(chan<- Player)

func (matchingServer *MatchingServer) matchAndPlay(players chan *Player) {
	var player1, player2 *Player
	for player := range players {
		if player1 == nil {
			player2 = player
		} else if player2 == nil {
			player1 = player
		} else {
			match := newMatch(player1, player2)
			matchingServer.mutex.Lock()
			matchingServer.liveMatches = append(matchingServer.liveMatches, &match)
			matchingServer.mutex.Unlock()
			(&match).play()
			matchingServer.mutex.Lock()
			matchingServer.removeMatch(&match)
			matchingServer.mutex.Unlock()
		}
		player1, player = nil, nil
	}
}

func (matchingServer *MatchingServer) Serve(players chan *Player, quit chan bool) {
	// Start handlers
	for i := 0; i < maxConcurrentGames; i++ {
		go matchingServer.matchAndPlay(players)
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
