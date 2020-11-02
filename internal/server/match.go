package server

import (
	"errors"
	"gochess/internal/model"
	"time"
)

type Match struct {
	black     *Player
	white     *Player
	game      *model.Game
	gameOver  chan struct{}
	maxTimeMs int64
}

type MatchGenerator func(black *Player, white *Player) Match

func newMatch(black *Player, white *Player, maxTime int64) Match {
	black.color = model.Black
	white.color = model.White
	if black.name == white.name {
		black.name = black.name + "_black"
		white.name = white.name + "_white"
	}
	black.elapsedMs = 0
	white.elapsedMs = 0
	game := model.NewGame()
	return Match{black, white, &game, make(chan struct{}), maxTime}
}

func DefaultMatchGenerator(black *Player, white *Player) Match {
	return newMatch(black, white, 1200000)
}

func (match *Match) play() {
	go match.handleAsyncRequests()
	for !match.game.GameOver() {
		match.handleTurn()
	}
}

func (match *Match) handleTurn() {
	player := match.black
	opponent := match.white
	if match.game.Turn() != model.Black {
		player = match.white
		opponent = match.black
	}
	turnStart := time.Now()
	timeRemaining := match.maxTimeMs - player.elapsedMs
	timer := time.AfterFunc(time.Duration(timeRemaining)*time.Millisecond,
		match.handleTimeout(opponent))
	defer timer.Stop()
	request := RequestSync{}
	select {
	case request = <-player.requestChanSync:
	case <-match.gameOver:
		return
	}
	response := ResponseSync{moveSuccess: true}
	err := errors.New("")
	for err != nil {
		err = match.game.Move(request.position, request.move)
		if err != nil {
			player.responseChanSync <- ResponseSync{moveSuccess: false}
			select {
			case request = <-player.requestChanSync:
			case <-match.gameOver:
				return
			}
		}
	}
	player.responseChanSync <- response
	if match.game.GameOver() {
		result := match.game.Result()
		winner := match.black
		if result.Winner == model.White {
			winner = match.white
		}
		match.handleGameOver(result.Draw, false, false, winner.name)
	}
	player.elapsedMs += time.Now().Sub(turnStart).Milliseconds()
}

func (match *Match) handleTimeout(opponent *Player) func() {
	return func() {
		match.game.SetGameResult(opponent.color, false)
		match.handleGameOver(false, false, true, opponent.name)
	}
}

func (match *Match) handleAsyncRequests() {
	for !match.game.GameOver() {
		opponent := match.white
		request := RequestAsync{}
		select {
		case request = <-match.black.requestChanAsync:
		case request = <-match.white.requestChanAsync:
			opponent = match.black
		case <-match.gameOver:
			return
		}
		if request.resign {
			match.game.SetGameResult(opponent.color, false)
			match.handleGameOver(false, true, false, opponent.name)
			return
		} else if request.requestToDraw {
			// TODO handle draw
		}
	}
}

func (match *Match) handleGameOver(
	draw, resignation, timeout bool, winner string,
) {
	response := ResponseAsync{gameOver: true, draw: draw,
		resignation: resignation, timeout: timeout, winner: winner}
	go func() {
		match.black.responseChanAsync <- response
	}()
	go func() {
		match.white.responseChanAsync <- response
	}()
	close(match.gameOver)
}
