package model

const (
	Black = Color(iota)
	White = Color(iota)
)

type (
	Color uint8

	Position struct {
		File, Rank uint8
	}

	board [8][8]*Piece
)

func NewPosition(file, rank uint8) Position {
	switch {
	case file >= 8:
		panic("Error: Invalid file " + string(file))
	case rank >= 8:
		panic("Error: Invalid rank " + string(rank))
	}
	return Position{file, rank}
}

func NewFullBoard() board {
	var board board
	// Create the pawns.
	for i := uint8(0); i < 16; i++ {
		color := Black
		rank := uint8(6)
		if i > 7 {
			color = White
			rank = 1
		}
		piece := NewPiece(Pawn, NewPosition(uint8(i%8), rank), color)
		board[piece.position.File][piece.position.Rank] = piece
	}
	// Create the rest.
	createTheBackLine(&board)
	return board
}

func NewBoardNoPawns() board {
	var board board
	createTheBackLine(&board)
	return board
}

func createTheBackLine(board *board) {
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

func (board board) Piece(pos Position) *Piece {
	return board[pos.File][pos.Rank]
}

func (board board) String() string {
	out := ""
	for rank := 7; rank >= 0; rank-- {
		for file := 0; file < 8; file++ {
			out += board[file][rank].String() + " "
		}
		out += "\n"
	}
	return out
}
