package model

import "errors"

type Game struct {
	board         *board
	turn          Color
	gameOver      bool
	result        gameResult
	previousMove  Move
	previousMover *Piece
	blackKing     *Piece
	whiteKing     *Piece
}

type gameResult struct {
	winner Color
	draw   bool
}

var ErrGameOver = errors.New("The game is over")

func (game *Game) Move(position Position, move Move) error {
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
	piece.takeMove(
		game.board, move, game.previousMove, game.previousMover, king,
	)
	enemyColor := getOppositeColor(piece.color)
	possibleEnemyMoves := AllMoves(
		game.board, enemyColor, move, piece, false, enemyKing,
	)
	if len(possibleEnemyMoves) == 0 &&
		enemyKing.isThreatened(game.board, move, piece) {
		game.gameOver = true
		game.result.winner = game.turn
	} else if len(possibleEnemyMoves) == 0 {
		game.gameOver = true
		game.result.draw = true
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
		&board, White, false, gameResult{}, Move{}, nil, board[4][7], board[4][0],
	}
}

func NewGameNoPawns() Game {
	board := NewBoardNoPawns()
	return Game{
		&board, White, false, gameResult{}, Move{}, nil, board[4][7], board[4][0],
	}
}

func (game *Game) Board() *board {
	return game.board
}

func (game *Game) Turn() Color {
	return game.turn
}

func (game *Game) GameOver() bool {
	return game.gameOver
}

func (game *Game) SetGameResult(winner Color, draw bool) {
	game.gameOver = true
	game.result.winner = winner
	game.result.draw = draw
}

func (game *Game) Winner() Color {
	return game.result.winner
}
