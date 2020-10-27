package server

type Piece struct {
	pieceType  PieceType
	position   position
	color      Color
	movesTaken uint16
}

func NewPiece(pieceType PieceType, position position, color Color) *Piece {
	return &Piece{pieceType, position, color, uint16(0)}
}

type PieceType uint8

const (
	Rook   = PieceType(iota)
	Knight = PieceType(iota)
	Bishop = PieceType(iota)
	Queen  = PieceType(iota)
	King   = PieceType(iota)
	Pawn   = PieceType(iota)
)

func (piece *Piece) String() string {
	if piece == nil {
		return "-"
	}
	switch piece.pieceType {
	case Rook:
		if piece.Color() == White {
			return "\u2656"
		} else {
			return "\u265C"
		}
	case Knight:
		if piece.Color() == White {
			return "\u2658"
		} else {
			return "\u265E"
		}
	case Bishop:
		if piece.Color() == White {
			return "\u2657"
		} else {
			return "\u265D"
		}
	case Queen:
		if piece.Color() == White {
			return "\u2655"
		} else {
			return "\u265B"
		}
	case King:
		if piece.Color() == White {
			return "\u2654"
		} else {
			return "\u265A"
		}
	case Pawn:
		if piece.Color() == White {
			return "\u2659"
		} else {
			return "\u265F"
		}
	default:
		return "-"
	}
}

func (piece *Piece) StringSimple() string {
	if piece == nil {
		return "-"
	}
	colorPurple := "\033[35m"
	colorWhite := "\033[37m"
	colorReset := "\033[0m"
	out := colorPurple
	if piece.Color() == White {
		out = colorWhite
	}
	switch piece.pieceType {
	case Rook:
		out += "r"
	case Knight:
		out += "k"
	case Bishop:
		out += "b"
	case Queen:
		out += "Q"
	case King:
		out += "K"
	case Pawn:
		out += "p"
	}
	return out + colorReset
}

func (piece *Piece) Position() position {
	return piece.position
}

func (piece *Piece) File() uint8 {
	return piece.position.File
}

func (piece *Piece) Rank() uint8 {
	return piece.position.Rank
}

func (piece *Piece) PieceType() PieceType {
	return piece.pieceType
}

func (piece *Piece) Color() Color {
	return piece.color
}
