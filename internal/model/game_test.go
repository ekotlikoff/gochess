package model

import (
	"fmt"
	"testing"
)

const debug bool = false

func TestNewGame(t *testing.T) {
	game := NewGame()
	if game.turn != White {
		t.Error("Expected turn = ", White, " got ", game.turn)
	}
	if game.blackKing.pieceType != King || game.whiteKing.pieceType != King {
		t.Error("Expected kings got ", game.blackKing.pieceType)
	}
}

func TestMoves(t *testing.T) {
	game := NewGame()
	if debug {
		fmt.Println(game.board)
	}
	game.Move(Position{0, 1}, Move{0, 2})
	if debug {
		fmt.Println(game.board)
	}
	game.Move(Position{0, 6}, Move{0, -2})
	if game.board[0][3] == nil || game.board[0][4] == nil {
		t.Error("Pawns did not move as expected")
	}
}

func TestCheckMate(t *testing.T) {
	game := NewGame()
	game.Move(Position{4, 1}, Move{0, 2})
	game.Move(Position{0, 6}, Move{0, -2})
	game.Move(Position{3, 0}, Move{4, 4})
	game.Move(Position{1, 6}, Move{0, -2})
	game.Move(Position{5, 0}, Move{-3, 3})
	game.Move(Position{2, 6}, Move{0, -2})
	game.Move(Position{7, 4}, Move{-2, 2})
	if debug {
		fmt.Println(game.board)
	}
	if game.gameOver == false || game.result.Winner != White {
		t.Error("Game should be over")
	}
}

func TestStalemate(t *testing.T) {
	game := NewGameNoPawns()
	if game.gameOver != false || game.result.Draw == true {
		t.Error("Game should not be over")
	}
	for i := 0; i < 3; i++ {
		game.board[i][0] = nil
		game.board[7-i][0] = nil
	}
	game.board[3][0] = nil
	game.Move(Position{4, 0}, Move{1, 0})
	game.Move(Position{7, 7}, Move{0, -6})
	game.Move(Position{5, 0}, Move{1, 0})
	game.Move(Position{0, 7}, Move{0, -6})
	game.Move(Position{6, 0}, Move{-1, 0})
	game.Move(Position{3, 7}, Move{0, -6})
	game.Move(Position{5, 0}, Move{1, 0})
	game.Move(Position{3, 1}, Move{1, 0})
	if debug {
		fmt.Println(game.board)
	}
	if game.gameOver == false || game.result.Draw != true {
		t.Error("Game should be a draw")
	}
}
