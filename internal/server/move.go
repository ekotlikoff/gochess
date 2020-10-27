package server

type Move struct {
	x, y int8
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
	piece.takeMove(board, move, Move{}, nil)
}

func (piece *Piece) takeMove(
	board *board, move Move, previousMove Move, previousMover *Piece,
) {
	if !piece.IsMoveValid(board, move, previousMove, previousMover) {
		panic("Error: Piece attempted invalid move.")
	}
	yDirection := int8(1)
	if piece.Color() == Black {
		yDirection *= -1
	}
	newX, newY := addMoveToPosition(piece, move)
	enPassantTarget := board[newX][newY+uint8(-1*yDirection)]
	isEnPassant := (piece.pieceType == Pawn &&
		newX != piece.File() &&
		enPassantTarget == previousMover &&
		enPassantTarget.pieceType == Pawn &&
		(previousMove.y == 2 || previousMove.y == -2))
	if isEnPassant {
		board[enPassantTarget.File()][enPassantTarget.Rank()] = nil
	}
	board[newX][newY] = piece
	board[piece.File()][piece.Rank()] = nil
	newPosition := position{newX, newY}
	piece.position = newPosition
	piece.movesTaken += 1
}

func (piece *Piece) IsMoveValid(
	board *board, move Move, previousMove Move, previousMover *Piece,
) bool {
	validMoves := piece.ValidMoves(board, previousMove, previousMover)
	for _, validMove := range validMoves {
		if validMove == move {
			return true
		}
	}
	return false
}

func (piece *Piece) validMoves(board *board) []Move {
	return piece.ValidMoves(board, Move{}, nil)
}

func (piece *Piece) ValidMoves(
	board *board, previousMove Move, previousMover *Piece,
) []Move {
	validMoves := []Move{}
	if piece == nil {
		return validMoves
	}
	baseMoves := moveMap[piece.PieceType()]
	canSlideCapture := true
	if piece.PieceType() == Pawn {
		canSlideCapture = false
	}
	for _, baseMove := range baseMoves {
		validMoves = append(validMoves, piece.validMovesSlide(
			baseMove,
			previousMove,
			previousMover,
			board,
			maxSlideMap[piece.PieceType()],
			canSlideCapture,
		)...)
	}
	return validMoves
}

func addMoveToPosition(piece *Piece, move Move) (uint8, uint8) {
	newX := uint8(int8(piece.Position().File) + move.x)
	newY := uint8(int8(piece.Position().Rank) + move.y)
	return newX, newY
}

func (piece *Piece) isMoveInBounds(move Move) bool {
	newX, newY := addMoveToPosition(piece, move)
	xInBounds := newX >= 0 && newX < 8
	yInBounds := newY >= 0 && newY < 8
	return xInBounds && yInBounds
}

func (piece *Piece) validCaptureMovesPawn(
	board *board, previousMove Move, previousMover *Piece,
) []Move {
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
		newX, newY := addMoveToPosition(piece, captureMove)
		pieceAtDest := board[newX][newY]
		enPassantTarget := board[newX][newY+uint8(-1*yDirection)]
		canEnPassant :=
			piece.canEnPassant(previousMove, previousMover, enPassantTarget)
		canCapture := pieceAtDest != nil && pieceAtDest.Color() != piece.Color()
		if canCapture || canEnPassant {
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
		(previousMove.y == 2 || previousMove.y == -2)
}

func (piece *Piece) validMovesSlide(
	move Move, previousMove Move, previousMover *Piece, board *board,
	maxSlide uint8, canSlideCapture bool,
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
			piece.validCaptureMovesPawn(board, previousMove, previousMover)...,
		)
	}
	for i := int8(1); i <= int8(maxSlide); i++ {
		slideMove := Move{move.x * i, move.y * i * yDirectionModifier}
		if !piece.isMoveInBounds(slideMove) {
			break
		}
		newX, newY := addMoveToPosition(piece, slideMove)
		pieceAtDest := board[newX][newY]
		destIsValidNoCapture := pieceAtDest == nil
		if destIsValidNoCapture {
			validSlides = append(validSlides, slideMove)
		} else {
			destIsValidCapture :=
				pieceAtDest.Color() != piece.Color() && canSlideCapture
			if destIsValidCapture {
				validSlides = append(validSlides, slideMove)
			}
			break
		}
	}
	return validSlides
}

func getThreatenedPositions() map[Position]bool {
	out := map[Position]bool{}
	// for each enemy piece
	// get validMoves
	// put validMoveDestinationPosition in return map
	return out
}

func (piece *Piece) canCastle(board *board) (casteLeft, castleRight bool) {
	if piece.pieceType != King || piece.movesTaken > 0 {
		return false, false
	}
	rookPieces := [2]*Piece{board[0][piece.Rank()], board[7][piece.Rank()]}
	casteLeft, castleRight = true
	for i, rook := range rookPieces {
		if rook.pieceType != Rook || rook.movesTaken != 0 {
			if i == 0 {
				castleLeft = false
			} else {
				castleRight = false
			}
		}
	}
	if !castleLeft && !castleRight {
		return
	}
	// getThreatenedPositions()
	// TODO if king is threatened, return false, false
	hitLeft, hitRight = false
	for i := 1; i < 5; i++ {
		if castleLeft && !hitLeft {
			// TODO if space is threatened, set false
			castleLeft, hitLeft =
				isCastleStillPossible(board, -i, rookPieces[0])
		}
		if castleRight && !hitRight {
			// TODO if space is threatened, set false
			castleRight, hitRight =
				isCastleStillPossible(board, i, rookPieces[1])
		}
	}
}

func (piece *Piece) isCastleStillPossible(
	board *board, xDiff uint8, rook Piece,
) (canStill, hit bool) {
	if board[piece.File()+xDiff][piece.Rank()] != nil &&
		board[piece.File()+xDiff][piece.Rank()] != rook {
		return
	} else if board[piece.File()-i][piece.Rank()] == rookPieces[0] {
		hit = true
	}
	canStill = true
}
