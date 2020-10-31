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
	game.Move(game.board[0][1], Move{0, 2})
	if debug {
		fmt.Println(game.board)
	}
	game.Move(game.board[0][6], Move{0, -2})
	if game.board[0][3] == nil || game.board[0][4] == nil {
		t.Error("Pawns did not move as expected")
	}
}

func TestCheckMate(t *testing.T) {
	game := NewGame()
	game.Move(game.board[4][1], Move{0, 2})
	game.Move(game.board[0][6], Move{0, -2})
	game.Move(game.board[3][0], Move{4, 4})
	game.Move(game.board[1][6], Move{0, -2})
	game.Move(game.board[5][0], Move{-3, 3})
	game.Move(game.board[2][6], Move{0, -2})
	game.Move(game.board[7][4], Move{-2, 2})
	if debug {
		fmt.Println(game.board)
	}
	if game.gameOver == false || game.winner != White {
		t.Error("Game should be over")
	}
}
