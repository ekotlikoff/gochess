package model

type game struct {
	board         *board
	turn          Color
	gameOver      bool
	winner        Color
	previousMove  Move
	previousMover *Piece
	blackKing     *Piece
	whiteKing     *Piece
}

func (game *game) Move(piece *Piece, move Move) {
	if piece.color != game.turn {
		panic("It's not your turn")
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
		game.winner = game.turn
	} else if len(possibleEnemyMoves) == 0 {
		// TODO handle stalemate
	}
	game.previousMove = move
	game.previousMover = piece
	game.turn = enemyColor
}

func getOppositeColor(color Color) (opposite Color) {
	if color == Black {
		opposite = White
	} else {
		opposite = Black
	}
	return opposite
}

func NewGame() game {
	board := NewFullBoard()
	return game{
		&board, White, false, White, Move{}, nil, board[4][7], board[4][0],
	}
}

func (game *game) Board() *board {
	return game.board
}

func (game *game) Turn() Color {
	return game.turn
}

func (game *game) GameOver() bool {
	return game.gameOver
}

func (game *game) Winner() Color {
	return game.winner
}
