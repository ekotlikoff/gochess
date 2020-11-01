package server

import (
	"fmt"
	"gochess/internal/model"
	"time"
)

type Match struct {
	black     *Player
	white     *Player
	game      *model.Game
	maxTimeMs int64
}

func newMatch(black *Player, white *Player) Match {
	black.color = model.Black
	white.color = model.White
	black.elapsedMs = 0
	white.elapsedMs = 0
	game := model.NewGame()
	return Match{black, white, &game, 1200000}
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
	timeRemaining := match.maxTimeMs - player.elapsedMs
	timer := time.NewTimer(time.Duration(timeRemaining) * time.Millisecond)
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
	fmt.Println(player.name)
	player.responseChan <- Response{success: true}
	timer.Stop()
	player.elapsedMs += time.Now().Sub(turnStart).Milliseconds()
}
