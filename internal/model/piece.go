package model

import (
	"bytes"
	"encoding/binary"
)

type (
	// Piece a chess piece
	Piece struct {
		PieceType  PieceType
		Position   Position
		Color      Color
		MovesTaken uint16
	}

	// PieceType the various piece types
	PieceType uint8
)

// NewPiece return a new piece
func NewPiece(pieceType PieceType, position Position, color Color) *Piece {
	return &Piece{pieceType, position, color, uint16(0)}
}

const (
	// Rook piece type
	Rook = PieceType(iota)
	// Knight piece type
	Knight = PieceType(iota)
	// Bishop piece type
	Bishop = PieceType(iota)
	// Queen piece type
	Queen = PieceType(iota)
	// King piece type
	King = PieceType(iota)
	// Pawn piece type
	Pawn = PieceType(iota)
)

// String return the piece's string
func (piece *Piece) String() string {
	if piece == nil {
		return "-"
	}
	switch piece.PieceType {
	case Rook:
		if piece.Color == White {
			return "\u2656"
		}
		return "\u265C"
	case Knight:
		if piece.Color == White {
			return "\u2658"
		}
		return "\u265E"
	case Bishop:
		if piece.Color == White {
			return "\u2657"
		}
		return "\u265D"
	case Queen:
		if piece.Color == White {
			return "\u2655"
		}
		return "\u265B"
	case King:
		if piece.Color == White {
			return "\u2654"
		}
		return "\u265A"
	case Pawn:
		if piece.Color == White {
			return "\u2659"
		}
		return "\u265F"
	default:
		return "-"
	}
}

// ClientString return the piece's client string
func (piece *Piece) ClientString() string {
	if piece == nil {
		return ""
	}
	out := ""
	if piece.Color == Black {
		out += "b"
	} else {
		out += "w"
	}
	switch piece.PieceType {
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

// Value return the piece's point value
func (piece *Piece) Value() int8 {
	if piece == nil {
		return 0
	}
	switch piece.PieceType {
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

// MarshalBinary return the piece's representation as a byte array
func (piece *Piece) MarshalBinary(
	board *Board, previousMove Move, previousMover *Piece, king *Piece,
) (data []byte, err error) {
	// Ensure temporary options (castle/en passant) are taken into account
	// for draw by repetition.
	temporaryMoveRights := uint8(0)
	if piece.PieceType == Pawn {
		validMoves :=
			piece.ValidMoves(board, previousMove, previousMover, false, king)
		temporaryMoveRights = uint8(len(validMoves))
	} else if piece.PieceType == King {
		castleLeft, castleRight := piece.hasCastleRights(board)
		if castleLeft {
			temporaryMoveRights++
		}
		if castleRight {
			temporaryMoveRights++
		}
	}
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, []byte{
		byte(piece.PieceType),
		byte(piece.Color),
		piece.Position.File,
		piece.Position.Rank,
		temporaryMoveRights,
	})
	return buf.Bytes(), err
}

// StringSimple return the piece's simple string representation
func (piece *Piece) StringSimple() string {
	if piece == nil {
		return "-"
	}
	colorPurple := "\033[35m"
	colorWhite := "\033[37m"
	colorReset := "\033[0m"
	out := colorPurple
	if piece.Color == White {
		out = colorWhite
	}
	switch piece.PieceType {
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

// File return the piece's file
func (piece *Piece) File() uint8 {
	return piece.Position.File
}

// Rank return the piece's rank
func (piece *Piece) Rank() uint8 {
	return piece.Position.Rank
}
