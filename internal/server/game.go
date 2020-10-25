package server

type game struct {
	board    board
	gameOver bool
	winner   Color
}

func NewGame() game {
	return game{board{}, false, White}

}
