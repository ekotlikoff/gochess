package server

type Color uint8

const (
	Black = Color(iota)
	White = Color(iota)
)

type position struct {
	File, Rank uint8
}

func NewPosition(file, rank uint8) position {
	switch {
	case file >= 8:
		panic("Invalid file " + string(file))
	case rank >= 8:
		panic("Invalid rank " + string(rank))
	}
	return position{file, rank}
}

type Move struct {
	Piece       Piece
	NewPosition position
}

type board [8][8]Piece

func NewFullBoard() board {
	var board [8][8]Piece
	// Create the pawns.
	for i := uint8(0); i < 16; i++ {
		color := Black
		rank := uint8(6)
		if i > 7 {
			color = White
			rank = 1
		}
		piece := Piece{NewPosition(uint8(i%8), rank), color, Pawn, true}
		board[piece.position.File][piece.position.Rank] = piece
	}
	// Create the rest.
	for i := uint8(0); i < 4; i++ {
		board[i][7] = Piece{NewPosition(uint8(i), 7), Black, PieceType(i + 1), true}
		board[i][0] = Piece{NewPosition(uint8(i), 0), White, PieceType(i + 1), true}
		pieceType := PieceType(i + 1)
		if i == 3 {
			pieceType = PieceType(i + 2)
		}
		board[7-i][7] = Piece{NewPosition(uint8(7-i), 7), Black, pieceType, true}
		board[7-i][0] = Piece{NewPosition(uint8(7-i), 0), White, pieceType, true}
	}
	return board
}

func (board board) String() string {
	out := ""
	for rank := 7; rank >= 0; rank-- {
		for file := 0; file < 8; file++ {
			out += board[file][rank].String()
		}
		out += "\n"
	}
	return out
}
