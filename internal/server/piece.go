package server

type Piece struct {
	position  position
	color     Color
	pieceType PieceType
	active    bool
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

func (piece *Piece) PieceType() PieceType {
	return piece.pieceType
}

func (piece *Piece) Color() Color {
	return piece.color
}

func (piece *Piece) Active() bool {
	return piece.active
}

type Move struct {
	x, y int8
}

var diagonalMoves = []Move{Move{1, 1}, Move{1, -1}, Move{-1, 1}, Move{-1, -1}}
var straightMoves = []Move{Move{0, 1}, Move{0, -1}, Move{1, 0}, Move{-1, 0}}

var moveMap = map[PieceType][]Move{
	Rook: straightMoves,
	Knight: {Move{1, 2}, Move{1, -2}, Move{-1, 2}, Move{-1, -2},
		Move{2, 1}, Move{2, -1}, Move{-2, 1}, Move{-2, -1}},
	Bishop: diagonalMoves,
	Queen:  append(diagonalMoves, straightMoves...),
	King:   append(diagonalMoves, straightMoves...),
	Pawn:   {Move{0, 1}},
}

var maxSlideMap = map[PieceType]uint8{
	Rook:   7,
	Knight: 1,
	Bishop: 7,
	Queen:  7,
	King:   1,
	Pawn:   2,
}

func (piece *Piece) ValidMoves(board *board) []Move {
	validMoves := []Move{}
	if piece == nil {
		return validMoves
	}
	baseMoves := moveMap[piece.PieceType()]
	for i := 0; i < len(baseMoves); i++ {
		validMoves = append(validMoves, piece.validMoves(
			baseMoves[i],
			board,
			maxSlideMap[piece.PieceType()],
			false,
		)...)
	}
	return validMoves
}

func (piece *Piece) isMoveInBounds(move Move) bool {
	newX := uint8(int8(piece.Position().File) + move.x)
	xInBounds := newX >= 0 && newX < 8
	newY := uint8(int8(piece.Position().Rank) + move.y)
	yInBounds := newY >= 0 && newY < 8
	return xInBounds && yInBounds
}

func (piece *Piece) validCaptureMovesPawn(board *board) []Move {
	yDirection := int8(1)
	if piece.Color() == Black {
		yDirection *= -1
	}
	captureMoves := []Move{}
	for xDirection := int8(-1); xDirection <= 1; xDirection += 2 {
		captureMove := Move{xDirection, yDirection}
		if !piece.isMoveInBounds(captureMove) {
			break
		}
		destX := uint8(int8(piece.Position().File) + captureMove.x)
		destY := uint8(int8(piece.Position().Rank) + captureMove.y)
		pieceAtDest := board[destX][destY]
		if pieceAtDest != nil && pieceAtDest.Color() != piece.Color() {
			captureMoves = append(captureMoves, captureMove)
		}
	}
	return captureMoves
}

func (piece *Piece) validMoves(
	move Move, board *board, maxSlide uint8, canCapture bool,
) []Move {
	validSlides := []Move{}
	yDirectionModifier := int8(1)
	if piece.PieceType() == Pawn {
		originalRank := uint8(1)
		if piece.Color() == Black {
			originalRank = uint8(6)
			yDirectionModifier = int8(-1)
		}
		if piece.Position().Rank != originalRank {
			maxSlide = 1
		}
		validSlides = append(validSlides, piece.validCaptureMovesPawn(board)...)
	}
	for i := int8(1); i <= int8(maxSlide); i++ {
		slideMove := Move{move.x * i, move.y * i * yDirectionModifier}
		if !piece.isMoveInBounds(slideMove) {
			break
		}
		destX := uint8(int8(piece.Position().File) + slideMove.x)
		destY := uint8(int8(piece.Position().Rank) + slideMove.y)
		pieceAtDest := board[destX][destY]
		destIsValid := pieceAtDest == nil ||
			(pieceAtDest.Color() != piece.Color() && canCapture)
		if destIsValid {
			validSlides = append(validSlides, slideMove)
		} else {
			break
		}
	}
	return validSlides
}
