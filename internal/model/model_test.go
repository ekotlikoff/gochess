package model

import (
	"testing"
)

func TestNewBoard(t *testing.T) {
	board := NewFullBoard()
	// fmt.Print(board)
	for file := 0; file < 8; file++ {
		blackPawn := board[file][6]
		if blackPawn.PieceType() != Pawn {
			t.Error("Expected pieceType = ", Pawn, " got ", blackPawn.PieceType())
		}
		if blackPawn.Color() != Black {
			t.Error("Expected color = ", Black, " got ", blackPawn.Color())
		}
		whitePawn := board[file][1]
		if whitePawn.PieceType() != Pawn {
			t.Error("Expected pieceType = ", Pawn, " got ", whitePawn.PieceType())
		}
		if whitePawn.Color() != White {
			t.Error("Expected color = ", White, " got ", whitePawn.Color())
		}
	}
}
