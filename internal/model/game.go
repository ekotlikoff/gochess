package model

import (
	"errors"
	"sync"
)

type Game struct {
	board         *board
	turn          Color
	gameOver      bool
	result        GameResult
	previousMove  Move
	previousMover *Piece
	blackKing     *Piece
	whiteKing     *Piece
	mutex         sync.RWMutex
}

type GameResult struct {
	Winner Color
	Draw   bool
}

type MoveRequest struct {
	Position  Position
	Move      Move
	PromoteTo *PieceType
}

var ErrGameOver = errors.New("The game is over")

func (game *Game) Move(moveRequest MoveRequest) error {
	game.mutex.Lock()
	defer game.mutex.Unlock()
	position := moveRequest.Position
	move := moveRequest.Move
	piece := game.board[position.File][position.Rank]
	if game.gameOver {
		return ErrGameOver
	} else if piece == nil {
		return errors.New("Cannot move nil piece")
	} else if piece.color != game.turn {
		return errors.New("It's not your turn")
	}
	king := game.blackKing
	enemyKing := game.whiteKing
	if game.turn == White {
		king = game.whiteKing
		enemyKing = game.blackKing
	}
	err := piece.takeMove(game.board, move, game.previousMove,
		game.previousMover, king, moveRequest.PromoteTo)
	if err != nil {
		return err
	}
	enemyColor := getOppositeColor(piece.color)
	possibleEnemyMoves := AllMoves(
		game.board, enemyColor, move, piece, false, enemyKing,
	)
	if len(possibleEnemyMoves) == 0 &&
		enemyKing.isThreatened(game.board, move, piece) {
		game.gameOver = true
		game.result.Winner = game.turn
	} else if len(possibleEnemyMoves) == 0 {
		game.gameOver = true
		game.result.Draw = true
	}
	game.previousMove = move
	game.previousMover = piece
	game.turn = enemyColor
	return nil
}

func getOppositeColor(color Color) (opposite Color) {
	if color == Black {
		opposite = White
	} else {
		opposite = Black
	}
	return opposite
}

func NewGame() Game {
	board := NewFullBoard()
	return Game{
		board: &board, turn: White,
		blackKing: board[4][7], whiteKing: board[4][0],
	}
}

func NewGameNoPawns() Game {
	board := NewBoardNoPawns()
	return Game{
		board: &board, turn: White,
		blackKing: board[4][7], whiteKing: board[4][0],
	}
}

func (game *Game) Board() *board {
	game.mutex.RLock()
	defer game.mutex.RUnlock()
	return game.board
}

func (game *Game) PointAdvantage(color Color) int8 {
	game.mutex.RLock()
	defer game.mutex.RUnlock()
	var points int8 = 0
	for _, file := range game.board {
		for _, piece := range file {
			if piece != nil {
				if piece.Color() == color {
					points += piece.Value()
				} else {
					points -= piece.Value()
				}
			}
		}
	}
	return points
}

func (game *Game) Turn() Color {
	game.mutex.RLock()
	defer game.mutex.RUnlock()
	return game.turn
}

func (game *Game) GameOver() bool {
	game.mutex.RLock()
	defer game.mutex.RUnlock()
	return game.gameOver
}

func (game *Game) SetGameResult(winner Color, draw bool) {
	game.mutex.Lock()
	defer game.mutex.Unlock()
	game.gameOver = true
	game.result.Winner = winner
	game.result.Draw = draw
}

func (game *Game) Result() GameResult {
	game.mutex.RLock()
	defer game.mutex.RUnlock()
	return game.result
}
