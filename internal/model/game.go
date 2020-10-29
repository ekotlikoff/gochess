package model

type game struct {
	board    board
	turn     Color
	gameOver bool
	winner   Color
}

func NewGame() game {
	return game{NewFullBoard(), White, false, White}
}

func (game *game) Board() board {
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
