package model

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
	validMoves := piece.ValidMoves(board, previousMove, previousMover, false)
	for _, validMove := range validMoves {
		if validMove == move {
			return true
		}
	}
	return false
}

func (piece *Piece) validMoves(board *board) []Move {
	return piece.ValidMoves(board, Move{}, nil, false)
}

func (piece *Piece) ValidMoves(
	board *board, previousMove Move, previousMover *Piece, allThreatened bool,
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
			allThreatened,
		)...)
	}
	if !allThreatened {
		validMoves = append(validMoves, piece.getCastleMove(
			board, previousMove, previousMover,
		)...)
	}
	return validMoves
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
	board *board, previousMove Move, previousMover *Piece, allThreatened bool,
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
		(previousMove.y == 2 || previousMove.y == -2)
}

func (piece *Piece) validMovesSlide(
	move Move, previousMove Move, previousMover *Piece, board *board,
	maxSlide uint8, canSlideCapture bool, allThreatened bool,
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
				board, previousMove, previousMover, allThreatened,
			)...,
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

func AllThreatenedPositions(
	board *board, enemyColor Color, previousMove Move, previousMover *Piece,
) map[position]bool {
	out := map[position]bool{}
	// for each enemy piece
	for _, file := range board {
		for _, piece := range file {
			if piece != nil && piece.color == enemyColor {
				for _, position := range piece.ThreatenedPositions(
					board, previousMove, previousMover,
				) {
					out[position] = true
				}
			}
		}
	}
	return out
}

func (piece *Piece) ThreatenedPositions(
	board *board, previousMove Move, previousMover *Piece,
) []position {
	positions := []position{}
	moves := piece.ValidMoves(board, previousMove, previousMover, true)
	for _, move := range moves {
		threatenedX, threatenedY := addMoveToPosition(piece, move)
		positions = append(positions, position{threatenedX, threatenedY})
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
		return
	}
	noBlockLeft, noBlockRight := piece.noPiecesBlockingCastle(board)
	if !noBlockLeft && !noBlockRight {
		return
	}
	enemyColor := Black
	if piece.color == enemyColor {
		enemyColor = White
	}
	threatenedPositions := AllThreatenedPositions(
		board, enemyColor, previousMove, previousMover,
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
	threatenedPositions map[position]bool,
) (left, right bool) {
	left, right = true, true
	for i := int8(0); i < 3; i++ {
		leftX, leftY := addMoveToPosition(piece, Move{-i, 0})
		rightX, rightY := addMoveToPosition(piece, Move{i, 0})
		if threatenedPositions[position{leftX, leftY}] {
			left = false
		}
		if threatenedPositions[position{rightX, rightY}] {
			right = false
		}
	}
	return left, right
}
