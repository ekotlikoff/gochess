package matchserver

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/Ekotlikoff/gochess/internal/model"
)

const clientGameoverNotifGracePeriod = 1 * time.Second

type (
	// Match is a struct representing a game between two players
	Match struct {
		black         *Player
		white         *Player
		Game          *model.Game
		gameOverChan  chan struct{}
		matchOverChan chan struct{}
		maxTimeMs     int64
		requestedDraw *Player
		mutex         sync.RWMutex
	}

	// MatchGenerator takes two players and creates a match
	MatchGenerator func(black *Player, white *Player) Match
)

// NewMatch create a new match between two players
func NewMatch(black *Player, white *Player, maxTimeMs int64) Match {
	black.color = model.Black
	white.color = model.White
	if black.name == white.name {
		black.name = black.name + "_black"
		white.name = white.name + "_white"
	}
	game := model.NewGame()
	return Match{black, white, game, make(chan struct{}), make(chan struct{}),
		maxTimeMs, nil, sync.RWMutex{}}
}

// Create a new match between two players with no pawns
func newMatchNoPawns(black *Player, white *Player, maxTimeMs int64) Match {
	black.color = model.Black
	white.color = model.White
	if black.name == white.name {
		black.name = black.name + "_black"
		white.name = white.name + "_white"
	}
	game := model.NewGameNoPawns()
	return Match{black, white, game, make(chan struct{}), make(chan struct{}),
		maxTimeMs, nil, sync.RWMutex{}}
}

// DefaultMatchGenerator default match generator
func DefaultMatchGenerator(p1 *Player, p2 *Player) Match {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	if r.Intn(2) > 0 {
		return NewMatch(p1, p2, 1200000)
	}
	return NewMatch(p2, p1, 1200000)
}

// CreateCustomMatchGenerator create a generator that creates matches with a
// custom match length in seconds
func CreateCustomMatchGenerator(matchPlayerTimeSeconds int) MatchGenerator {
	return func(p1 *Player, p2 *Player) Match {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		if r.Intn(2) > 0 {
			return NewMatch(p1, p2, int64(matchPlayerTimeSeconds*1000))
		}
		return NewMatch(p2, p1, int64(matchPlayerTimeSeconds*1000))
	}
}

// PlayerName get the player name corresponding to the input color
func (match *Match) PlayerName(color model.Color) string {
	match.mutex.RLock()
	defer match.mutex.RUnlock()
	if color == model.Black {
		return match.black.Name()
	}
	return match.white.Name()
}

// MaxTimeMs get match max time
func (match *Match) MaxTimeMs() int64 {
	match.mutex.RLock()
	defer match.mutex.RUnlock()
	return match.maxTimeMs
}

// PlayerRemainingTimeMs get the player corresponding to the input color
func (match *Match) PlayerRemainingTimeMs(color model.Color) int64 {
	match.mutex.RLock()
	defer match.mutex.RUnlock()
	if color == model.Black {
		return match.maxTimeMs - match.black.elapsedMs
	}
	return match.maxTimeMs - match.white.elapsedMs
}

// GameOver check if the game is over
func (match *Match) GameOver() bool {
	match.mutex.RLock()
	defer match.mutex.RUnlock()
	return match.Game.GameOver()
}

// GetRequestedDraw get the current player who has requested a draw
func (match *Match) GetRequestedDraw() *Player {
	match.mutex.RLock()
	defer match.mutex.RUnlock()
	return match.requestedDraw
}

// SetRequestedDraw store the fact that the player has requested a draw
func (match *Match) SetRequestedDraw(player *Player) {
	match.mutex.Lock()
	defer match.mutex.Unlock()
	match.requestedDraw = player
}

func (match *Match) play() {
	waitc := make(chan struct{})
	go match.handleAsyncRequests(waitc)
	for !match.Game.GameOver() {
		match.handleTurn()
	}
	<-waitc
	match.matchOver()
}

func (match *Match) matchOver() {
	close(match.matchOverChan)
}

func (match *Match) waitForMatchOver() {
	<-match.matchOverChan
}

func (match *Match) handleTurn() {
	player := match.black
	opponent := match.white
	if match.Game.Turn() != model.Black {
		player = match.white
		opponent = match.black
	}
	turnStart := time.Now()
	timeRemaining := match.maxTimeMs - player.elapsedMs
	timer := time.AfterFunc(time.Duration(timeRemaining)*time.Millisecond,
		match.handleTimeout(opponent))
	defer timer.Stop()
	request := model.MoveRequest{}
	select {
	case request = <-player.requestChanSync:
	case <-match.gameOverChan:
		return
	}
	err := errors.New("")
	for err != nil {
		err = match.Game.Move(request)
		if err != nil {
			select {
			case player.ResponseChanSync <- ResponseSync{MoveSuccess: false}:
			case <-match.gameOverChan:
				return
			}
			select {
			case request = <-player.requestChanSync:
			case <-match.gameOverChan:
				return
			}
		}
	}
	if !timer.Stop() {
		return
	}
	match.SetRequestedDraw(nil)
	player.elapsedMs += time.Since(turnStart).Milliseconds()
	player.ResponseChanSync <- ResponseSync{
		MoveSuccess: true, ElapsedMs: int(player.elapsedMs),
		ElapsedMsOpponent: int(opponent.elapsedMs),
	}
	opponent.OpponentPlayedMove <- request
	if match.Game.GameOver() {
		result := match.Game.Result()
		winner := match.black
		if result.Winner == model.White {
			winner = match.white
		}
		match.handleGameOver(result.Draw, false, false, winner)
	}
}

func (match *Match) handleTimeout(opponent *Player) func() {
	return func() {
		onlyKing := match.Game.OnlyKing(model.Black)
		if opponent.color == model.White {
			onlyKing = match.Game.OnlyKing(model.White)
		}
		if onlyKing {
			match.handleGameOver(true, false, true, opponent)
		} else {
			match.handleGameOver(false, false, true, opponent)
		}
	}
}

func (match *Match) handleAsyncRequests(waitc chan struct{}) {
	defer close(waitc)
	for !match.Game.GameOver() {
		opponent := match.white
		player := match.black
		request := RequestAsync{}
		select {
		case request = <-match.black.RequestChanAsync:
		case request = <-match.white.RequestChanAsync:
			opponent = match.black
			player = match.white
		case <-match.gameOverChan:
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
				go func() {
					select {
					case opponent.ResponseChanAsync <- ResponseAsync{
						false, true, false, false, false, false, "", MatchDetails{},
					}:
					case <-match.gameOverChan:
					}
				}()
			} else {
				match.SetRequestedDraw(player)
				go func() {
					select {
					case opponent.ResponseChanAsync <- ResponseAsync{
						false, true, false, false, false, false, "", MatchDetails{},
					}:
					case <-match.gameOverChan:
					}
				}()
			}
		}
	}
}

func (match *Match) handleGameOver(
	draw, resignation, timeout bool, winner *Player,
) {
	match.mutex.Lock()
	defer match.mutex.Unlock()
	select {
	case <-match.gameOverChan:
		return
	default:
		break
	}
	match.Game.SetGameResult(winner.color, draw)
	winnerName := winner.name
	if draw {
		winnerName = ""
	}
	response := ResponseAsync{GameOver: true, Draw: draw,
		Resignation: resignation, Timeout: timeout, Winner: winnerName}
	var wg sync.WaitGroup
	for _, player := range [2]*Player{match.black, match.white} {
		thisPlayer := player
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case thisPlayer.ResponseChanAsync <- response:
			case <-time.After(clientGameoverNotifGracePeriod):
			}
		}()
	}
	wg.Wait()
	close(match.gameOverChan)
}
