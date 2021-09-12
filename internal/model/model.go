package model

import "strconv"

const (
	// Black is one of the two chess colors
	Black = Color(iota)
	// White is one of the two chess colors
	White = Color(iota)
)

type (
	// Color of a piece or player
	Color uint8

	// Position is a struct representing a chess piece's position
	Position struct {
		File, Rank uint8
	}

	// Board is a chess board of pieces
	Board [8][8]*Piece
	// SerializableBoard is a chess board of pieces that can be serialized
	SerializableBoard []Piece
)

// NewPosition creates a new position
func NewPosition(file, rank uint8) Position {
	switch {
	case file >= 8:
		panic("Error: Invalid file " + string(file))
	case rank >= 8:
		panic("Error: Invalid rank " + string(rank))
	}
	return Position{file, rank}
}

func newFullBoard() Board {
	var board Board
	// Create the pawns.
	for i := uint8(0); i < 16; i++ {
		color := Black
		rank := uint8(6)
		if i > 7 {
			color = White
			rank = 1
		}
		piece := NewPiece(Pawn, NewPosition(uint8(i%8), rank), color)
		board[piece.Position.File][piece.Position.Rank] = piece
	}
	// Create the rest.
	createTheBackLine(&board)
	return board
}

func newBoardNoPawns() Board {
	var board Board
	createTheBackLine(&board)
	return board
}

func newBoardFromSerializableBoard(serializableBoard SerializableBoard) *Board {
	var board Board
	for _, piece := range serializableBoard {
		board[piece.File()][piece.Rank()] = &piece
	}
	return &board
}

func createTheBackLine(board *Board) {
	for i := uint8(0); i < 4; i++ {
		board[i][7] = NewPiece(PieceType(i), NewPosition(uint8(i), 7), Black)
		board[i][0] = NewPiece(PieceType(i), NewPosition(uint8(i), 0), White)
		pieceType := PieceType(i)
		if i == 3 {
			pieceType = PieceType(i + 1)
		}
		board[7-i][7] = NewPiece(pieceType, NewPosition(uint8(7-i), 7), Black)
		board[7-i][0] = NewPiece(pieceType, NewPosition(uint8(7-i), 0), White)
	}
}

// Piece get a piece from the board
func (board Board) Piece(pos Position) *Piece {
	return board[pos.File][pos.Rank]
}

func (board Board) String() string {
	out := ""
	for rank := 7; rank >= 0; rank-- {
		for file := 0; file < 8; file++ {
			out += board[file][rank].String() + " "
		}
		out += "\n"
	}
	return out
}

func (pos Position) String() string {
	return strconv.Itoa(int(pos.File)) + "," + strconv.Itoa(int(pos.Rank))
}
