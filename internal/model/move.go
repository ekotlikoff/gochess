package model

import (
	"errors"
	"fmt"
)

type Move struct {
	X, Y int8
}

func (move *Move) String() string {
	return fmt.Sprintf("%d,%d", move.X, move.Y)
}

var diagonalMoves = []Move{Move{1, 1}, Move{1, -1}, Move{-1, 1}, Move{-1, -1}}
var straightMoves = []Move{Move{0, 1}, Move{0, -1}, Move{1, 0}, Move{-1, 0}}

var moveMap = map[PieceType][]Move{
	Rook: straightMoves,
	Knight: {
		Move{1, 2}, Move{1, -2}, Move{-1, 2}, Move{-1, -2},
		Move{2, 1}, Move{2, -1}, Move{-2, 1}, Move{-2, -1},
	},
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

func (piece *Piece) takeMoveShort(board *board, move Move) {
	piece.takeMove(board, move, Move{}, nil, nil, nil)
}

func (piece *Piece) takeMove(
	board *board, move Move, previousMove Move, previousMover *Piece,
	king *Piece, promoteTo *PieceType,
) error {
	if !piece.IsMoveValid(board, move, previousMove, previousMover, king,
		promoteTo) {
		return errors.New("Piece attempted invalid move.")
	}
	piece.takeMoveUnsafe(board, move, previousMove, previousMover, promoteTo)
	piece.movesTaken += 1
	return nil
}

func (piece *Piece) takeMoveUnsafe(
	board *board, move Move, previousMove Move, previousMover *Piece,
	promoteTo *PieceType) (newPosition Position, capturedPiece *Piece,
	newCastledPosition Position, castledRook *Piece) {
	yDirection := int8(1)
	if piece.Color() == Black {
		yDirection *= -1
	}
	newX, newY := addMoveToPosition(piece, move)
	enPassantTargetY := uint8(int8(newY) + int8(-1*yDirection))
	enPassantTarget := &Piece{}
	if newX >= 0 && newX <= 7 && enPassantTargetY >= 0 && enPassantTargetY <= 7 {
		enPassantTarget = board[newX][enPassantTargetY]
	}
	isEnPassant := (piece.pieceType == Pawn && newX != piece.File() &&
		enPassantTarget != nil && enPassantTarget == previousMover &&
		enPassantTarget.pieceType == Pawn &&
		(previousMove.Y == 2 || previousMove.Y == -2))
	isCastle := piece.pieceType == King && (move.X < -1 || move.X > 1)
	if isEnPassant {
		capturedPiece = board[enPassantTarget.File()][enPassantTarget.Rank()]
		board[enPassantTarget.File()][enPassantTarget.Rank()] = nil
	} else if isCastle {
		if move.X < 0 {
			board[3][piece.Rank()] = board[0][piece.Rank()]
			board[0][piece.Rank()] = nil
			castledRook = board[3][piece.Rank()]
			newCastledPosition = Position{3, piece.Rank()}
		} else {
			board[5][piece.Rank()] = board[7][piece.Rank()]
			board[7][piece.Rank()] = nil
			castledRook = board[5][piece.Rank()]
			newCastledPosition = Position{5, piece.Rank()}
		}
		castledRook.position = newCastledPosition
	}
	if board[newX][newY] != nil {
		capturedPiece = board[newX][newY]
	}
	board[newX][newY] = piece
	board[piece.File()][piece.Rank()] = nil
	newPosition = Position{newX, newY}
	piece.position = newPosition
	if promoteTo != nil {
		piece.pieceType = *promoteTo
	}
	return newPosition, capturedPiece, newCastledPosition, castledRook
}

func (piece *Piece) IsMoveValid(
	board *board, move Move, previousMove Move, previousMover *Piece,
	king *Piece, promoteTo *PieceType,
) bool {
	validMoves :=
		piece.ValidMoves(board, previousMove, previousMover, false, king)
	for _, validMove := range validMoves {
		if validMove == move && piece.promotionValid(move, promoteTo) {
			return true
		}
	}
	return false
}

func (piece *Piece) validMoves(board *board) []Move {
	return piece.ValidMoves(board, Move{}, nil, false, nil)
}

func (piece *Piece) ValidMoves(
	board *board, previousMove Move, previousMover *Piece,
	allThreatened bool, king *Piece,
) []Move {
	validMoves := []Move{}
	baseMoves := moveMap[piece.PieceType()]
	canSlideCapture := true
	if piece.PieceType() == Pawn {
		canSlideCapture = false
	}
	for _, baseMove := range baseMoves {
		validMoves = append(validMoves, piece.validMovesSlide(
			baseMove, previousMove, previousMover, board,
			maxSlideMap[piece.PieceType()], canSlideCapture,
			allThreatened, king,
		)...)
	}
	if !allThreatened {
		validMoves = append(validMoves, piece.getCastleMove(
			board, previousMove, previousMover,
		)...)
	}
	return validMoves
}

func (piece *Piece) promotionValid(move Move, promoteTo *PieceType) bool {
	validPromoteTypes := map[PieceType]struct{}{
		Bishop: struct{}{}, Knight: struct{}{},
		Rook: struct{}{}, Queen: struct{}{},
	}
	_, newY := addMoveToPosition(piece, move)
	if piece.PieceType() == Pawn {
		if promoteTo == nil {
			return newY != 7 && newY != 0
		}
		_, promoteToIsValid := validPromoteTypes[*promoteTo]
		return (newY == 7 || newY == 0) && promoteToIsValid
	}
	return promoteTo == nil
}

func (piece *Piece) getCastleMove(
	board *board, previousMove Move, previousMover *Piece,
) []Move {
	out := []Move{}
	canLeft, canRight := piece.canCastle(board, previousMove, previousMover)
	if canLeft {
		out = append(out, Move{int8(-2), 0})
	}
	if canRight {
		out = append(out, Move{int8(2), 0})
	}
	return out
}

func addMoveToPosition(piece *Piece, move Move) (uint8, uint8) {
	newX := uint8(int8(piece.Position().File) + move.X)
	newY := uint8(int8(piece.Position().Rank) + move.Y)
	return newX, newY
}

func (piece *Piece) isMoveInBounds(move Move) bool {
	newX, newY := addMoveToPosition(piece, move)
	xInBounds := newX >= 0 && newX < 8
	yInBounds := newY >= 0 && newY < 8
	return xInBounds && yInBounds
}

func (piece *Piece) validCaptureMovesPawn(
	board *board, previousMove Move, previousMover *Piece, allThreatened bool,
	king *Piece,
) []Move {
	yDirection := int8(1)
	if piece.Color() == Black {
		yDirection *= -1
	}
	captureMoves := []Move{}
	for xDirection := int8(-1); xDirection <= 1; xDirection += 2 {
		captureMove := Move{xDirection, yDirection}
		wouldBeInCheck := func() bool {
			return !allThreatened && piece.wouldBeInCheck(
				board, captureMove, previousMove, previousMover, king,
			)
		}
		if !piece.isMoveInBounds(captureMove) || wouldBeInCheck() {
			continue
		}
		newX, newY := addMoveToPosition(piece, captureMove)
		pieceAtDest := board[newX][newY]
		enPassantTarget := board[newX][newY+uint8(-1*yDirection)]
		canEnPassant :=
			piece.canEnPassant(previousMove, previousMover, enPassantTarget)
		canCapture := pieceAtDest != nil && pieceAtDest.Color() != piece.Color()
		if canCapture || canEnPassant || allThreatened {
			captureMoves = append(captureMoves, captureMove)
		}
	}
	return captureMoves
}

func (piece *Piece) canEnPassant(
	previousMove Move, previousMover *Piece, enPassantTarget *Piece,
) bool {
	return enPassantTarget != nil && enPassantTarget == previousMover &&
		enPassantTarget.Color() != piece.Color() &&
		enPassantTarget.pieceType == Pawn &&
		(previousMove.Y == 2 || previousMove.Y == -2)
}

func (piece *Piece) validMovesSlide(
	move Move, previousMove Move, previousMover *Piece, board *board,
	maxSlide uint8, canSlideCapture bool, allThreatened bool, king *Piece,
) []Move {
	validSlides := []Move{}
	yDirectionModifier := int8(1)
	if piece.PieceType() == Pawn {
		if piece.Color() == Black {
			yDirectionModifier = int8(-1)
		}
		if piece.movesTaken > 0 {
			maxSlide = 1
		}
		validSlides = append(
			validSlides,
			piece.validCaptureMovesPawn(
				board, previousMove, previousMover, allThreatened, king,
			)...,
		)
	}
	for i := int8(1); i <= int8(maxSlide); i++ {
		slideMove := Move{move.X * i, move.Y * i * yDirectionModifier}
		wouldBeInCheck := func() bool {
			return !allThreatened && piece.wouldBeInCheck(
				board, slideMove, previousMove, previousMover, king,
			)
		}
		if !piece.isMoveInBounds(slideMove) || wouldBeInCheck() {
			continue
		}
		newX, newY := addMoveToPosition(piece, slideMove)
		pieceAtDest := board[newX][newY]
		destIsValidNoCapture := pieceAtDest == nil
		if destIsValidNoCapture && (piece.pieceType != Pawn || !allThreatened) {
			validSlides = append(validSlides, slideMove)
		} else {
			destIsValidCapture :=
				canSlideCapture && pieceAtDest.Color() != piece.Color()
			if destIsValidCapture {
				validSlides = append(validSlides, slideMove)
			}
			break
		}
	}
	return validSlides
}

func AllMoves(
	board *board, color Color, previousMove Move, previousMover *Piece,
	allThreatened bool, king *Piece,
) map[Position]bool {
	out := map[Position]bool{}
	// for each enemy piece
	for _, file := range board {
		for _, piece := range file {
			if piece != nil && piece.color == color {
				for _, position := range piece.Moves(
					board, previousMove, previousMover, allThreatened, king,
				) {
					out[position] = true
				}
			}
		}
	}
	return out
}

func (piece *Piece) Moves(
	board *board, previousMove Move, previousMover *Piece, allThreatened bool,
	king *Piece,
) []Position {
	positions := []Position{}
	moves :=
		piece.ValidMoves(board, previousMove, previousMover, allThreatened,
			king)
	for _, move := range moves {
		threatenedX, threatenedY := addMoveToPosition(piece, move)
		positions = append(positions, Position{threatenedX, threatenedY})
	}
	return positions
}

func (piece *Piece) canCastle(
	board *board, previousMove Move, previousMover *Piece,
) (castleLeft, castleRight bool) {
	if piece.pieceType != King || piece.movesTaken > 0 {
		return false, false
	}
	rookPieces := [2]*Piece{board[0][piece.Rank()], board[7][piece.Rank()]}
	castleLeft, castleRight = true, true
	for i, rook := range rookPieces {
		if rook == nil || rook.pieceType != Rook || rook.movesTaken != 0 {
			if i == 0 {
				castleLeft = false
			} else {
				castleRight = false
			}
		}
	}
	if !castleLeft && !castleRight {
		return false, false
	}
	noBlockLeft, noBlockRight := piece.noPiecesBlockingCastle(board)
	if !noBlockLeft && !noBlockRight {
		return false, false
	}
	enemyColor := getOppositeColor(piece.color)
	threatenedPositions := AllMoves(
		board, enemyColor, previousMove, previousMover, true, nil,
	)
	noCheckLeft, noCheckRight :=
		piece.wouldNotCastleThroughCheck(threatenedPositions)
	castleLeft = castleLeft && noBlockLeft && noCheckLeft
	castleRight = castleRight && noBlockRight && noCheckRight
	return castleLeft, castleRight
}

func (piece *Piece) noPiecesBlockingCastle(board *board) (left, right bool) {
	left, right = true, true
	for i := int8(1); i < 4; i++ {
		leftX, leftY := addMoveToPosition(piece, Move{-i, 0})
		rightX, rightY := addMoveToPosition(piece, Move{i, 0})
		if board[leftX][leftY] != nil {
			left = false
		}
		if i != 3 && board[rightX][rightY] != nil {
			right = false
		}
	}
	return left, right
}

func (piece *Piece) wouldNotCastleThroughCheck(
	threatenedPositions map[Position]bool,
) (left, right bool) {
	left, right = true, true
	for i := int8(0); i < 3; i++ {
		leftX, leftY := addMoveToPosition(piece, Move{-i, 0})
		rightX, rightY := addMoveToPosition(piece, Move{i, 0})
		if threatenedPositions[Position{leftX, leftY}] {
			left = false
		}
		if threatenedPositions[Position{rightX, rightY}] {
			right = false
		}
	}
	return left, right
}

func (piece *Piece) wouldBeInCheck(
	board *board, move Move, previousMove Move, previousMover *Piece,
	king *Piece,
) bool {
	if king == nil {
		return false
	}
	originalPosition := piece.position
	newPosition, capturedPiece, newCastledPosition, castledRook :=
		piece.takeMoveUnsafe(board, move, previousMove, previousMover, nil)
	wouldBeInCheck := king.isThreatened(board, move, piece)
	// Revert the move
	board[newPosition.File][newPosition.Rank] = nil
	if capturedPiece != nil {
		board[capturedPiece.File()][capturedPiece.Rank()] = capturedPiece
	}
	board[originalPosition.File][originalPosition.Rank] = piece
	piece.position = originalPosition
	if castledRook != nil {
		board[newCastledPosition.File][newCastledPosition.Rank] = nil
		board[newCastledPosition.File][newCastledPosition.Rank] = nil
		if castledRook.position.File == 5 {
			castledRook.position.File = 7
		} else {
			castledRook.position.File = 0
		}
	}
	return wouldBeInCheck
}

func (piece *Piece) isThreatened(board *board, previousMove Move,
	previousMover *Piece,
) bool {
	enemyColor := Black
	if piece.color == Black {
		enemyColor = White
	}
	threatenedPositions := AllMoves(
		board, enemyColor, previousMove, previousMover, true, nil,
	)
	inCheck := false
	if threatenedPositions[piece.position] {
		inCheck = true
	}
	return inCheck
}
