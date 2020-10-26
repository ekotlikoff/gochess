package server

import (
	"fmt"
	"testing"
)

const debug bool = false

func TestValidMovesPawnUnmoved(t *testing.T) {
	game := NewGame()
	board := game.Board()
	if debug {
		fmt.Println(board)
	}
	validMoves := board[0][1].ValidMoves(&board)
	if (len(validMoves) != 2 || validMoves[0] != Move{0, 1} || validMoves[1] != Move{0, 2}) {
		t.Error("Expected moves = {0, 1}, {0, 2} got ", validMoves)
	}
	validMoves = board[0][6].ValidMoves(&board)
	if (len(validMoves) != 2 || validMoves[0] != Move{0, -1} || validMoves[1] != Move{0, -2}) {
		t.Error("Expected moves = {0, -1}, {0, -2} got ", validMoves)
	}
}

func TestValidMovesPawnCapture(t *testing.T) {
	game := NewGame()
	board := game.Board()
	board[1][5] = board[1][1]
	board[1][1] = nil
	board[1][5].position.Rank = 5
	if debug {
		fmt.Println(board)
	}
	validMoves := board[1][5].ValidMoves(&board)
	if (len(validMoves) != 2 || validMoves[0] != Move{-1, 1} || validMoves[1] != Move{1, 1}) {
		t.Error("Expected capture moves got ", validMoves)
	}
}

func TestValidMovesPawnMoved(t *testing.T) {
	game := NewGame()
	board := game.Board()
	board[0][2] = board[0][1]
	board[0][1] = nil
	board[0][2].position.Rank = 2
	if debug {
		fmt.Println(board)
	}
	validMoves := board[0][2].ValidMoves(&board)
	if (len(validMoves) != 1 || validMoves[0] != Move{0, 1}) {
		t.Error("Expected moves = {0, 1} got ", validMoves)
	}
}

func TestValidMovesRook(t *testing.T) {
	game := NewGame()
	board := game.Board()
	board[0][5] = board[0][1]
	board[0][1] = nil
	board[0][5].position.Rank = 5
	if debug {
		fmt.Println(board)
	}
	validMoves := board[0][0].ValidMoves(&board)
	if (len(validMoves) != 4 || validMoves[3] != Move{0, 4}) {
		t.Error("Expected rook moves, got ", validMoves)
	}
}

// TODO TestValidMovesRookCapture
