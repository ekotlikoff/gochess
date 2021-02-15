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
	game.Move(MoveRequest{Position{0, 1}, Move{0, 2}, nil})
	if debug {
		fmt.Println(game.board)
	}
	game.Move(MoveRequest{Position{0, 6}, Move{0, -2}, nil})
	if game.board[0][3] == nil || game.board[0][4] == nil {
		t.Error("Pawns did not move as expected")
	}
}

func TestPoints(t *testing.T) {
	game := NewGame()
	if debug {
		fmt.Println(game.board)
	}
	if game.PointAdvantage(Black) != 0 {
		t.Error("Expected zero point advantage")
	}
	game.Move(MoveRequest{Position{0, 1}, Move{0, 2}, nil})
	game.Move(MoveRequest{Position{1, 6}, Move{0, -2}, nil})
	game.Move(MoveRequest{Position{0, 3}, Move{1, 1}, nil})
	if debug {
		fmt.Println(game.board)
	}
	if game.PointAdvantage(Black) != -1 || game.PointAdvantage(White) != 1 {
		t.Error("Expected point advantage == 1")
	}
}

func TestCheckMate(t *testing.T) {
	game := NewGame()
	game.Move(MoveRequest{Position{4, 1}, Move{0, 2}, nil})
	game.Move(MoveRequest{Position{0, 6}, Move{0, -2}, nil})
	game.Move(MoveRequest{Position{3, 0}, Move{4, 4}, nil})
	game.Move(MoveRequest{Position{1, 6}, Move{0, -2}, nil})
	game.Move(MoveRequest{Position{5, 0}, Move{-3, 3}, nil})
	game.Move(MoveRequest{Position{2, 6}, Move{0, -2}, nil})
	game.Move(MoveRequest{Position{7, 4}, Move{-2, 2}, nil})
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
	game.Move(MoveRequest{Position{4, 0}, Move{1, 0}, nil})
	game.Move(MoveRequest{Position{7, 7}, Move{0, -6}, nil})
	game.Move(MoveRequest{Position{5, 0}, Move{1, 0}, nil})
	game.Move(MoveRequest{Position{0, 7}, Move{0, -6}, nil})
	game.Move(MoveRequest{Position{6, 0}, Move{-1, 0}, nil})
	game.Move(MoveRequest{Position{3, 7}, Move{0, -6}, nil})
	game.Move(MoveRequest{Position{5, 0}, Move{1, 0}, nil})
	game.Move(MoveRequest{Position{3, 1}, Move{1, 0}, nil})
	if debug {
		fmt.Println(game.board)
	}
	if game.gameOver == false || game.result.Draw != true {
		t.Error("Game should be a draw")
	}
}

func TestEnPassant(t *testing.T) {
	game := NewGame()
	if game.gameOver != false || game.result.Draw == true {
		t.Error("Game should not be over")
	}
	game.Move(MoveRequest{Position{4, 1}, Move{0, 2}, nil})
	game.Move(MoveRequest{Position{7, 6}, Move{0, -2}, nil})
	game.Move(MoveRequest{Position{4, 3}, Move{0, 1}, nil})
	game.Move(MoveRequest{Position{3, 6}, Move{0, -2}, nil})
	if debug {
		fmt.Println(game.board)
	}
	if game.board[3][4] == nil {
		t.Error("Pawn should have moved as expected")
	}
	game.Move(MoveRequest{Position{4, 4}, Move{-1, 1}, nil})
	if game.board[3][4] != nil || game.board[3][5] == nil {
		t.Error("Pawn should have been taken en passant")
	}
}

func TestPromotion(t *testing.T) {
	game := NewGame()
	if game.gameOver != false || game.result.Draw == true {
		t.Error("Game should not be over")
	}
	promotionType := Rook
	game.Move(MoveRequest{Position{4, 1}, Move{0, 2}, nil})
	game.Move(MoveRequest{Position{7, 6}, Move{0, -1}, nil})
	game.Move(MoveRequest{Position{4, 3}, Move{0, 1}, nil})
	game.Move(MoveRequest{Position{7, 5}, Move{0, -1}, nil})
	game.Move(MoveRequest{Position{4, 4}, Move{0, 1}, nil})
	game.Move(MoveRequest{Position{3, 6}, Move{0, -1}, nil})
	game.Move(MoveRequest{Position{4, 5}, Move{1, 1}, nil})
	game.Move(MoveRequest{Position{4, 7}, Move{-1, -1}, nil})
	err := game.Move(MoveRequest{Position{5, 6}, Move{1, 1}, nil})
	if err == nil || game.board[6][7].PieceType() != Knight {
		t.Error("Pawn should not have moved since the move is invalid without promotion.")
	}
	game.Move(MoveRequest{Position{5, 6}, Move{1, 1}, &promotionType})
	if debug {
		fmt.Println(game.board)
	}
	if game.board[6][7] == nil || game.board[6][7].PieceType() != Rook {
		t.Error("Pawn should have been promoted")
	}
}
