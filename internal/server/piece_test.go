package server

import (
	"fmt"
	"testing"
)

const debug bool = false

func movePiece(board *board, oldX uint8, oldY uint8, newX uint8, newY uint8) {
	board[newX][newY] = board[oldX][oldY]
	board[oldX][oldY] = nil
	board[newX][newY].position.File = newX
	board[newX][newY].position.Rank = newY
}

func TestValidMovesPawnUnmoved(t *testing.T) {
	board := NewFullBoard()
	if debug {
		fmt.Println(board)
	}
	validMoves := board[0][1].validMoves(&board)
	if (len(validMoves) != 2 || validMoves[0] != Move{0, 1} || validMoves[1] != Move{0, 2}) {
		t.Error("Expected moves = {0, 1}, {0, 2} got ", validMoves)
	}
	validMoves = board[0][6].validMoves(&board)
	if (len(validMoves) != 2 || validMoves[0] != Move{0, -1} || validMoves[1] != Move{0, -2}) {
		t.Error("Expected moves = {0, -1}, {0, -2} got ", validMoves)
	}
	movePiece(&board, 0, 6, 0, 2)
	validMoves = board[0][1].validMoves(&board)
	if len(validMoves) != 0 {
		t.Error("Expected no valid moves got ", validMoves)
	}
}

func TestValidMovesPawnCapture(t *testing.T) {
	board := NewFullBoard()
	movePiece(&board, 1, 1, 1, 5)
	if debug {
		fmt.Println(board)
	}
	validMoves := board[1][5].validMoves(&board)
	if (len(validMoves) != 2 || validMoves[0] != Move{-1, 1} || validMoves[1] != Move{1, 1}) {
		t.Error("Expected capture moves got ", validMoves)
	}
}

func TestValidMovesPawnEnPassant(t *testing.T) {
	board := NewFullBoard()
	movePiece(&board, 4, 6, 4, 3)
	board[3][1].takeMoveShort(&board, Move{0, 2})
	if debug {
		fmt.Println(board)
	}
	validMoves := board[4][3].ValidMoves(&board, Move{0, 2}, board[3][3], false)
	if (len(validMoves) != 2 || validMoves[0] != Move{-1, -1}) {
		t.Error("Expected en passant moves got ", validMoves)
	}
	board[4][3].takeMove(&board, Move{-1, -1}, Move{0, 2}, board[3][3])
	if board[3][3] != nil {
		t.Error("Expected en passant to take the en passant target")
	}
}

func TestValidMovesPawnNoEnPassant(t *testing.T) {
	board := NewFullBoard()
	movePiece(&board, 4, 6, 4, 3)
	board[3][1].takeMoveShort(&board, Move{0, 1})
	board[3][2].takeMoveShort(&board, Move{0, 1})
	if debug {
		fmt.Println(board)
	}
	validMoves := board[4][3].ValidMoves(&board, Move{0, 1}, board[3][3], false)
	if (len(validMoves) != 1 || validMoves[0] != Move{0, -1}) {
		t.Error("Expected no en passant moves got ", validMoves)
	}
}

func TestValidMovesPawnMoved(t *testing.T) {
	board := NewFullBoard()
	board[0][1].takeMoveShort(&board, Move{0, 1})
	if debug {
		fmt.Println(board)
	}
	validMoves := board[0][2].validMoves(&board)
	if (len(validMoves) != 1 || validMoves[0] != Move{0, 1}) {
		t.Error("Expected moves = {0, 1} got ", validMoves)
	}
}

func TestThreatenedPositionsPawn(t *testing.T) {
	board := NewFullBoard()
	if debug {
		fmt.Println(board)
	}
	threatenedPositions := board[1][1].ThreatenedPositions(&board, Move{}, nil)
	if len(threatenedPositions) != 2 {
		t.Error("Expected pawn threatens, got ", threatenedPositions)
	}
	movePiece(&board, 1, 1, 5, 5)
	threatenedPositions = board[5][5].ThreatenedPositions(&board, Move{}, nil)
	if len(threatenedPositions) != 2 {
		t.Error("Expected pawn threatens, got ", threatenedPositions)
	}
}

func TestValidMovesRook(t *testing.T) {
	board := NewFullBoard()
	movePiece(&board, 0, 1, 0, 5)
	if debug {
		fmt.Println(board)
	}
	validMoves := board[0][0].validMoves(&board)
	if (len(validMoves) != 4 || validMoves[3] != Move{0, 4}) {
		t.Error("Expected rook moves, got ", validMoves)
	}
}

func TestValidMovesRookCapture(t *testing.T) {
	board := NewFullBoard()
	movePiece(&board, 0, 1, 1, 3)
	if debug {
		fmt.Println(board)
	}
	validMoves := board[0][0].validMoves(&board)
	if (len(validMoves) != 6 || validMoves[5] != Move{0, 6}) {
		t.Error("Expected rook moves, got ", validMoves)
	}
}

func TestValidMovesRookMultiple(t *testing.T) {
	board := NewFullBoard()
	movePiece(&board, 0, 0, 4, 4)
	if debug {
		fmt.Println(board)
	}
	validMoves := board[4][4].validMoves(&board)
	if len(validMoves) != 11 {
		t.Error("Expected rook moves, got ", validMoves)
	}
}

func TestThreatenedPositionsRook(t *testing.T) {
	board := NewFullBoard()
	if debug {
		fmt.Println(board)
	}
	movePiece(&board, 0, 0, 4, 4)
	threatenedPositions := board[4][4].ThreatenedPositions(&board, Move{}, nil)
	if len(threatenedPositions) != 11 {
		t.Error("Expected rook threatens, got ", threatenedPositions)
	}
}

func TestValidMovesKnight(t *testing.T) {
	board := NewFullBoard()
	if debug {
		fmt.Println(board)
	}
	validMoves := board[1][0].validMoves(&board)
	if (len(validMoves) != 2 || validMoves[0] != Move{1, 2}) {
		t.Error("Expected knight moves, got ", validMoves)
	}
}

func TestValidMovesKnightMultple(t *testing.T) {
	board := NewFullBoard()
	movePiece(&board, 6, 7, 3, 3)
	if debug {
		fmt.Println(board)
	}
	validMoves := board[3][3].validMoves(&board)
	if (len(validMoves) != 8 || validMoves[0] != Move{1, 2}) {
		t.Error("Expected knight moves, got ", validMoves)
	}
}

func TestThreatenedPositionsKnight(t *testing.T) {
	board := NewFullBoard()
	if debug {
		fmt.Println(board)
	}
	movePiece(&board, 6, 7, 3, 3)
	threatenedPositions := board[3][3].ThreatenedPositions(&board, Move{}, nil)
	if len(threatenedPositions) != 8 {
		t.Error("Expected knight threatens, got ", threatenedPositions)
	}
}

func TestValidMovesBishop(t *testing.T) {
	board := NewFullBoard()
	movePiece(&board, 1, 1, 1, 2)
	movePiece(&board, 3, 1, 3, 2)
	if debug {
		fmt.Println(board)
	}
	validMoves := board[2][0].validMoves(&board)
	if len(validMoves) != 7 {
		t.Error("Expected bishop moves, got ", validMoves)
	}
}

func TestValidMovesBishopMultiple(t *testing.T) {
	board := NewFullBoard()
	movePiece(&board, 2, 0, 4, 4)
	if debug {
		fmt.Println(board)
	}
	validMoves := board[4][4].validMoves(&board)
	if len(validMoves) != 8 {
		t.Error("Expected bishop moves, got ", validMoves)
	}
}

func TestValidMovesQueenMultiple(t *testing.T) {
	board := NewFullBoard()
	movePiece(&board, 3, 0, 4, 4)
	if debug {
		fmt.Println(board)
	}
	validMoves := board[4][4].validMoves(&board)
	if len(validMoves) != 19 {
		t.Error("Expected queen moves, got ", validMoves)
	}
}

func TestValidMovesKing(t *testing.T) {
	board := NewFullBoard()
	movePiece(&board, 4, 7, 4, 5)
	if debug {
		fmt.Println(board)
	}
	validMoves := board[4][5].validMoves(&board)
	if len(validMoves) != 5 {
		t.Error("Expected king moves, got ", validMoves)
	}
}

func TestValidMovesKingMultiple(t *testing.T) {
	board := NewFullBoard()
	movePiece(&board, 4, 7, 1, 2)
	if debug {
		fmt.Println(board)
	}
	validMoves := board[1][2].validMoves(&board)
	if len(validMoves) != 8 {
		t.Error("Expected king moves, got ", validMoves)
	}
}

func TestValidMovesKingCastle(t *testing.T) {
	board := NewFullBoard()
	movePiece(&board, 6, 0, 4, 5)
	movePiece(&board, 5, 0, 4, 6)
	movePiece(&board, 1, 0, 4, 6)
	movePiece(&board, 2, 0, 4, 6)
	movePiece(&board, 3, 0, 4, 6)
	if debug {
		fmt.Println(board)
	}
	validMoves := board[4][0].validMoves(&board)
	if len(validMoves) != 4 {
		t.Error("Expected king moves, got ", validMoves)
	}
}
