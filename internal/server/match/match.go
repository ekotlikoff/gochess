package matchserver

import (
	"errors"
	"github.com/Ekotlikoff/gochess/internal/model"
	"sync"
	"time"
)

type Match struct {
	black         *Player
	white         *Player
	game          *model.Game
	gameOver      chan struct{}
	maxTimeMs     int64
	requestedDraw *Player
	mutex         sync.RWMutex
}

type MatchGenerator func(black *Player, white *Player) Match

func NewMatch(black *Player, white *Player, maxTimeMs int64) Match {
	black.color = model.Black
	white.color = model.White
	if black.name == white.name {
		black.name = black.name + "_black"
		white.name = white.name + "_white"
	}
	black.elapsedMs = 0
	white.elapsedMs = 0
	game := model.NewGame()
	return Match{black, white, &game, make(chan struct{}), maxTimeMs, nil,
		sync.RWMutex{}}
}

func DefaultMatchGenerator(black *Player, white *Player) Match {
	return NewMatch(black, white, 1200000)
}

func (match *Match) PlayerName(color model.Color) string {
	if color == model.Black {
		return match.black.Name()
	}
	return match.white.Name()
}

func (match *Match) GetRequestedDraw() *Player {
	match.mutex.RLock()
	defer match.mutex.RUnlock()
	return match.requestedDraw
}

func (match *Match) SetRequestedDraw(player *Player) {
	match.mutex.Lock()
	defer match.mutex.Unlock()
	match.requestedDraw = player
}

func (match *Match) MaxTimeMs() int64 {
	return match.maxTimeMs
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
	err := errors.New("")
	for err != nil {
		err = match.game.Move(model.MoveRequest{request.position, request.move})
		if err != nil {
			select {
			case player.responseChanSync <- ResponseSync{moveSuccess: false}:
			case <-match.gameOver:
				return
			}
			select {
			case request = <-player.requestChanSync:
			case <-match.gameOver:
				return
			}
		}
	}
	match.SetRequestedDraw(nil)
	player.responseChanSync <- ResponseSync{moveSuccess: true}
	opponent.opponentPlayedMove <- model.MoveRequest{request.position, request.move}
	if match.game.GameOver() {
		result := match.game.Result()
		winner := match.black
		if result.Winner == model.White {
			winner = match.white
		}
		match.handleGameOver(result.Draw, false, false, winner)
	}
	player.elapsedMs += time.Now().Sub(turnStart).Milliseconds()
}

func (match *Match) handleTimeout(opponent *Player) func() {
	return func() {
		match.handleGameOver(false, false, true, opponent)
	}
}

func (match *Match) handleAsyncRequests() {
	for !match.game.GameOver() {
		opponent := match.white
		player := match.black
		request := RequestAsync{}
		select {
		case request = <-match.black.requestChanAsync:
		case request = <-match.white.requestChanAsync:
			opponent = match.black
			player = match.white
		case <-match.gameOver:
			return
		}
		if request.Resign {
			match.handleGameOver(false, true, false, opponent)
			return
		} else if request.RequestToDraw {
			if match.GetRequestedDraw() == opponent {
				match.handleGameOver(true, false, false, opponent)
			} else if match.GetRequestedDraw() == player {
				// Consider the second requestToDraw a toggle.
				match.SetRequestedDraw(nil)
			} else {
				match.SetRequestedDraw(player)
				go func() {
					select {
					case opponent.responseChanAsync <- ResponseAsync{
						false, true, false, false, false, "",
					}:
					case <-time.After(10 * time.Second):
					}
				}()
			}
		}
	}
}

func (match *Match) handleGameOver(
	draw, resignation, timeout bool, winner *Player,
) {
	match.game.SetGameResult(winner.color, draw)
	winnerName := winner.name
	if draw {
		winnerName = ""
	}
	response := ResponseAsync{GameOver: true, Draw: draw,
		Resignation: resignation, Timeout: timeout, Winner: winnerName}
	for _, player := range [2]*Player{match.black, match.white} {
		thisPlayer := player
		go func() {
			select {
			case thisPlayer.responseChanAsync <- response:
			case <-time.After(5 * time.Second):
			}
		}()
	}
	close(match.gameOver)
}
