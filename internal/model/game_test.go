package model

import (
	"testing"
)

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
	game.Move(game.board[0][1], Move{0, 2})
	game.Move(game.board[0][6], Move{0, -2})
	if game.board[0][3] == nil || game.board[0][4] == nil {
		t.Error("Pawns did not move as expected")
	}
}
