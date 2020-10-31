package server

import (
	"gochess/internal/model"
	"time"
)

const maxTimeMs = 1200000

type Match struct {
	black *Player
	white *Player
	game  *model.Game
}

func newMatch(black *Player, white *Player) Match {
	black.color = model.Black
	white.color = model.White
	black.elapsedMs = 0
	white.elapsedMs = 0
	game := model.NewGame()
	return Match{black, white, &game}
}

func (match *Match) play() {
	for !match.game.GameOver() {
		match.handleTurn()
	}
}

func (match *Match) handleTurn() {
	player := match.black
	if match.game.Turn() != model.Black {
		player = match.white
	}
	turnStart := time.Now()
	timer := time.NewTimer(time.Duration(maxTimeMs-player.elapsedMs) * time.Millisecond)
	go func() {
		<-timer.C
		// TODO handle timeout
	}()
	request := <-player.requestChan
	if request.requestToDraw || request.resign {
		// TODO handle draw/resignation
	} else {
		match.game.Move(request.position, request.move)
	}
	// TODO send response to player.responseChan
	timer.Stop()
	player.elapsedMs += time.Now().Sub(turnStart).Milliseconds()
}
