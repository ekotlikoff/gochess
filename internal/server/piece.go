package server

type Piece struct {
	position  position
	color     Color
	pieceType PieceType
	active    bool
}

type PieceType uint8

const (
	Empty  = PieceType(iota)
	Rook   = PieceType(iota)
	Knight = PieceType(iota)
	Bishop = PieceType(iota)
	Queen  = PieceType(iota)
	King   = PieceType(iota)
	Pawn   = PieceType(iota)
)

func (piece Piece) String() string {
	switch piece.pieceType {
	case Rook:
		return "R"
	case Knight:
		return "K"
	case Bishop:
		return "B"
	case Queen:
		return "Q"
	case King:
		return "K"
	case Pawn:
		return "P"
	default:
		return "."
	}
}

func (piece *Piece) Position() position {
	return piece.position
}

func (piece *Piece) Color() Color {
	return piece.color
}

func (piece *Piece) Active() bool {
	return piece.active
}

func (piece *Piece) validMoves() []Move {
	return make([]Move, 0)
}
