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
	if game.turn == White {
		king = game.whiteKing
	}
	// TODO Check for checkmate and configure gameOver/winner
	piece.takeMove(
		game.board, move, game.previousMove, game.previousMover, king,
	)
	game.previousMove = move
	game.previousMover = piece
	if game.turn == Black {
		game.turn = White
	} else {
		game.turn = Black
	}
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
