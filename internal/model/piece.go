package model

import (
	"bytes"
	"encoding/binary"
)

type (
	Piece struct {
		pieceType  PieceType
		position   Position
		color      Color
		movesTaken uint16
	}

	PieceType uint8
)

func NewPiece(pieceType PieceType, position Position, color Color) *Piece {
	return &Piece{pieceType, position, color, uint16(0)}
}

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

func (piece *Piece) ClientString() string {
	if piece == nil {
		return ""
	}
	out := ""
	if piece.Color() == Black {
		out += "b"
	} else {
		out += "w"
	}
	switch piece.pieceType {
	case Rook:
		out += "r"
	case Knight:
		out += "n"
	case Bishop:
		out += "b"
	case Queen:
		out += "q"
	case King:
		out += "k"
	case Pawn:
		out += "p"
	}
	return out
}

func (piece *Piece) Value() int8 {
	if piece == nil {
		return 0
	}
	switch piece.pieceType {
	case Queen:
		return 9
	case Rook:
		return 5
	case Knight, Bishop:
		return 3
	case Pawn:
		return 1
	}
	return 0
}

func (piece *Piece) MarshalBinary(
	board *board, previousMove Move, previousMover *Piece, king *Piece,
) (data []byte, err error) {
	// Ensure temporary options (castle/en passant) are taken into account
	// for draw by repetition.
	temporaryMoveRights := uint8(0)
	if piece.pieceType == Pawn {
		validMoves :=
			piece.ValidMoves(board, previousMove, previousMover, false, king)
		temporaryMoveRights = uint8(len(validMoves))
	} else if piece.pieceType == King {
		castleLeft, castleRight := piece.hasCastleRights(board)
		if castleLeft {
			temporaryMoveRights += 1
		}
		if castleRight {
			temporaryMoveRights += 1
		}
	}
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, []byte{
		byte(piece.pieceType),
		byte(piece.color),
		piece.position.File,
		piece.position.Rank,
		temporaryMoveRights,
	})
	return buf.Bytes(), err
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

func (piece *Piece) Position() Position {
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
