package model

import (
	"bytes"
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

func TestLotsOfPawns(t *testing.T) {
	game := NewGame()
	if debug {
		fmt.Println(game.board)
	}
	game.Move(MoveRequest{Position{0, 1}, Move{0, 2}, nil})
	game.Move(MoveRequest{Position{0, 6}, Move{0, -2}, nil})
	game.Move(MoveRequest{Position{1, 1}, Move{0, 2}, nil})
	game.Move(MoveRequest{Position{1, 6}, Move{0, -2}, nil})
	game.Move(MoveRequest{Position{2, 1}, Move{0, 2}, nil})
	game.Move(MoveRequest{Position{2, 6}, Move{0, -2}, nil})
	game.Move(MoveRequest{Position{3, 1}, Move{0, 2}, nil})
	game.Move(MoveRequest{Position{3, 6}, Move{0, -2}, nil})
	game.Move(MoveRequest{Position{4, 1}, Move{0, 2}, nil})
	game.Move(MoveRequest{Position{4, 6}, Move{0, -2}, nil})
	game.Move(MoveRequest{Position{5, 1}, Move{0, 2}, nil})
	game.Move(MoveRequest{Position{5, 6}, Move{0, -2}, nil})
	for i := 0; i < 6; i++ {
		if game.board[i][3] == nil || game.board[i][4] == nil {
			t.Error("Pawns did not move as expected")
		}
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

func TestCheckMate2(t *testing.T) {
	game := NewGame()
	game.Move(MoveRequest{Position{6, 0}, Move{1, 2}, nil})
	game.Move(MoveRequest{Position{2, 6}, Move{0, -2}, nil})
	game.Move(MoveRequest{Position{7, 2}, Move{-1, -2}, nil})
	game.Move(MoveRequest{Position{2, 4}, Move{0, -1}, nil})
	game.Move(MoveRequest{Position{6, 0}, Move{1, 2}, nil})
	game.Move(MoveRequest{Position{2, 3}, Move{0, -1}, nil})
	game.Move(MoveRequest{Position{7, 2}, Move{-1, -2}, nil})
	game.Move(MoveRequest{Position{2, 2}, Move{-1, -1}, nil})
	game.Move(MoveRequest{Position{6, 0}, Move{1, 2}, nil})
	queen := Queen
	game.Move(MoveRequest{Position{1, 1}, Move{-1, -1}, &queen})
	game.Move(MoveRequest{Position{7, 2}, Move{-1, -2}, nil})
	game.Move(MoveRequest{Position{0, 0}, Move{1, 0}, nil})
	game.Move(MoveRequest{Position{6, 0}, Move{1, 2}, nil})
	game.Move(MoveRequest{Position{3, 7}, Move{-2, -2}, nil})
	game.Move(MoveRequest{Position{7, 2}, Move{-1, -2}, nil})
	game.Move(MoveRequest{Position{1, 5}, Move{0, -4}, nil})
	game.Move(MoveRequest{Position{6, 0}, Move{1, 2}, nil})
	game.Move(MoveRequest{Position{1, 1}, Move{1, -1}, nil})
	game.Move(MoveRequest{Position{3, 0}, Move{-1, 0}, nil})
	game.Move(MoveRequest{Position{1, 0}, Move{1, 0}, nil})
	if debug {
		fmt.Println(game.board)
	}
	if game.gameOver == false || game.result.Winner != Black {
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

func TestPositionEncoding(t *testing.T) {
	game := NewGame()
	if game.gameOver != false || game.result.Draw == true {
		t.Error("Game should not be over")
	}
	bytes1, _ := game.MarshalBinary()
	game.Move(MoveRequest{Position{4, 1}, Move{0, 2}, nil})
	game.Move(MoveRequest{Position{7, 6}, Move{0, -1}, nil})
	game.Move(MoveRequest{Position{4, 3}, Move{0, 1}, nil})
	game.Move(MoveRequest{Position{7, 5}, Move{0, -1}, nil})
	game.Move(MoveRequest{Position{4, 4}, Move{0, 1}, nil})
	game.Move(MoveRequest{Position{3, 6}, Move{0, -1}, nil})
	game.Move(MoveRequest{Position{4, 5}, Move{1, 1}, nil})
	game.Move(MoveRequest{Position{4, 7}, Move{-1, -1}, nil})
	bytes2, _ := game.MarshalBinary()
	if bytes.Compare(bytes1, bytes2) == 0 {
		t.Error("Positions should not be equivalent")
	}
	if debug {
		fmt.Println(game.board)
		fmt.Println(bytes1)
		fmt.Println(len(bytes1))
		fmt.Println(bytes2)
		fmt.Println(len(bytes2))
	}
	game.Move(MoveRequest{Position{3, 0}, Move{1, 1}, nil})
	game.Move(MoveRequest{Position{3, 6}, Move{-1, -1}, nil})
	game.Move(MoveRequest{Position{4, 1}, Move{-1, -1}, nil})
	game.Move(MoveRequest{Position{2, 5}, Move{1, 1}, nil})
	bytes3, _ := game.MarshalBinary()
	if bytes.Compare(bytes2, bytes3) != 0 {
		t.Error("Positions should be equivalent")
	}
}

func TestPositionEncodingCastle(t *testing.T) {
	game := NewGame()
	if game.gameOver != false || game.result.Draw == true {
		t.Error("Game should not be over")
	}
	game.Move(MoveRequest{Position{4, 1}, Move{0, 2}, nil})
	game.Move(MoveRequest{Position{4, 6}, Move{0, -2}, nil})
	game.Move(MoveRequest{Position{5, 0}, Move{-5, 5}, nil})
	game.Move(MoveRequest{Position{5, 7}, Move{-5, -5}, nil})
	game.Move(MoveRequest{Position{6, 0}, Move{1, 2}, nil})
	game.Move(MoveRequest{Position{6, 7}, Move{1, -2}, nil})
	positionWithCastle, _ := game.MarshalBinary()
	if debug {
		fmt.Println(game.board)
	}
	game.Move(MoveRequest{Position{4, 0}, Move{0, 1}, nil})
	game.Move(MoveRequest{Position{4, 7}, Move{0, -1}, nil})
	game.Move(MoveRequest{Position{4, 1}, Move{0, -1}, nil})
	game.Move(MoveRequest{Position{4, 6}, Move{0, 1}, nil})
	positionWithoutCastle, _ := game.MarshalBinary()
	if bytes.Compare(positionWithCastle, positionWithoutCastle) == 0 {
		t.Error("Positions should not be equivalent")
	}
}

func TestPositionEncodingEnPassant(t *testing.T) {
	game := NewGame()
	if game.gameOver != false || game.result.Draw == true {
		t.Error("Game should not be over")
	}
	game.Move(MoveRequest{Position{4, 1}, Move{0, 2}, nil})
	game.Move(MoveRequest{Position{7, 6}, Move{0, -2}, nil})
	game.Move(MoveRequest{Position{4, 3}, Move{0, 1}, nil})
	game.Move(MoveRequest{Position{3, 6}, Move{0, -2}, nil})
	positionWithEnPassant, _ := game.MarshalBinary()
	if debug {
		fmt.Println(game.board)
	}
	game.Move(MoveRequest{Position{3, 0}, Move{1, 1}, nil})
	game.Move(MoveRequest{Position{3, 7}, Move{-1, -1}, nil})
	game.Move(MoveRequest{Position{4, 1}, Move{-1, -1}, nil})
	game.Move(MoveRequest{Position{2, 6}, Move{1, 1}, nil})
	err := game.Move(MoveRequest{Position{4, 4}, Move{-1, 1}, nil})
	if err == nil {
		t.Error("En passant should no longer be an option")
	}
	positionWithoutEnPassant, _ := game.MarshalBinary()
	if bytes.Compare(positionWithEnPassant, positionWithoutEnPassant) == 0 {
		t.Error("Positions should not be equivalent")
	}
}

func TestDrawByRepetion(t *testing.T) {
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
	for i := 0; i < 2; i++ {
		game.Move(MoveRequest{Position{3, 0}, Move{1, 1}, nil})
		game.Move(MoveRequest{Position{3, 7}, Move{0, -1}, nil})
		game.Move(MoveRequest{Position{4, 1}, Move{-1, -1}, nil})
		err := game.Move(MoveRequest{Position{3, 6}, Move{0, 1}, nil})
		if err != nil {
			t.Error("Draw too early")
		}
	}
	if !game.gameOver || !game.result.Draw {
		t.Error("Game should be a draw")
	}
}

func TestDrawByFiftyMoveRule(t *testing.T) {
	game := NewGameNoPawns()
	if game.gameOver != false || game.result.Draw == true {
		t.Error("Game should not be over")
	}
	if debug {
		fmt.Println(game.board)
	}
	game.Move(MoveRequest{Position{2, 0}, Move{-2, 2}, nil})
	game.Move(MoveRequest{Position{5, 7}, Move{-5, -5}, nil})
	// 1 move without capture
	game.Move(MoveRequest{Position{3, 0}, Move{0, 1}, nil})
	game.Move(MoveRequest{Position{0, 2}, Move{5, 5}, nil})
	// 24 moves without capture
	for i := uint8(0); i < 6; i++ {
		// Move knight out and back
		game.Move(MoveRequest{Position{1, 0}, Move{1, 2}, nil})
		game.Move(MoveRequest{Position{1, 7}, Move{1, -2}, nil})
		game.Move(MoveRequest{Position{2, 2}, Move{-1, -2}, nil})
		game.Move(MoveRequest{Position{2, 5}, Move{-1, 2}, nil})
		// Move bishop out and back
		if i%2 == 0 {
			game.Move(MoveRequest{Position{5, 0}, Move{-1, 1}, nil})
			game.Move(MoveRequest{Position{5, 7}, Move{-1, -1}, nil})
		} else {
			game.Move(MoveRequest{Position{4, 1}, Move{1, -1}, nil})
			game.Move(MoveRequest{Position{4, 6}, Move{1, 1}, nil})
		}
		// Move rook
		game.Move(MoveRequest{Position{0, i}, Move{0, 1}, nil})
		err := game.Move(MoveRequest{Position{7, 7 - i}, Move{0, -1}, nil})
		if err != nil {
			t.Error("Draw too early")
		}
	}
	// 1 move without capture
	game.Move(MoveRequest{Position{0, 6}, Move{1, 0}, nil})
	game.Move(MoveRequest{Position{7, 1}, Move{-1, 0}, nil})
	// 5 moves without capture
	for i := uint8(0); i < 5; i++ {
		game.Move(MoveRequest{Position{1, 6 - i}, Move{0, -1}, nil})
		err := game.Move(MoveRequest{Position{6, i + 1}, Move{0, 1}, nil})
		if err != nil {
			t.Error("Draw too early")
		}
	}
	// 1 move without capture
	game.Move(MoveRequest{Position{1, 1}, Move{1, 0}, nil})
	game.Move(MoveRequest{Position{6, 6}, Move{-1, 0}, nil})
	// 16 moves without capture
	for i := uint8(0); i < 4; i++ {
		// Move knight out and back
		game.Move(MoveRequest{Position{1, 0}, Move{-1, 2}, nil})
		game.Move(MoveRequest{Position{1, 7}, Move{-1, -2}, nil})
		game.Move(MoveRequest{Position{0, 2}, Move{1, -2}, nil})
		game.Move(MoveRequest{Position{0, 5}, Move{1, 2}, nil})
		// Move bishop out and back
		if i%2 == 0 {
			game.Move(MoveRequest{Position{5, 0}, Move{-1, 1}, nil})
			game.Move(MoveRequest{Position{5, 7}, Move{-1, -1}, nil})
		} else {
			game.Move(MoveRequest{Position{4, 1}, Move{1, -1}, nil})
			game.Move(MoveRequest{Position{4, 6}, Move{1, 1}, nil})
		}
		// Move rook
		game.Move(MoveRequest{Position{2, i + 1}, Move{0, 1}, nil})
		err := game.Move(MoveRequest{Position{5, 6 - i}, Move{0, -1}, nil})
		if err != nil {
			t.Error("Draw too early")
		}
	}
	// 2 moves without capture
	game.Move(MoveRequest{Position{1, 0}, Move{-1, 2}, nil})
	game.Move(MoveRequest{Position{1, 7}, Move{-1, -2}, nil})
	game.Move(MoveRequest{Position{0, 2}, Move{1, -2}, nil})
	game.Move(MoveRequest{Position{0, 5}, Move{1, 2}, nil})
	if !game.gameOver || !game.result.Draw {
		t.Error("Game should be a draw")
	}
}
