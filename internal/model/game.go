package model

import (
	"bytes"
	"encoding/binary"
	"errors"
	"sync"
)

type (
	// Game is a struct representing a chess game
	Game struct {
		board                       *Board
		turn                        Color
		gameOver                    bool
		result                      GameResult
		previousMove                Move
		previousMover               *Piece
		blackKing                   *Piece
		whiteKing                   *Piece
		whitePieces                 map[PieceType]uint8
		blackPieces                 map[PieceType]uint8
		positionHistory             map[string]uint8
		turnsSinceCaptureOrPawnMove uint8
		mutex                       sync.RWMutex
	}

	// GameResult is a struct representing the result of a game
	GameResult struct {
		Winner Color
		Draw   bool
	}

	// MoveRequest is a move request that can be applied to a game
	MoveRequest struct {
		Position  Position
		Move      Move
		PromoteTo *PieceType
	}
)

// Move make a move
func (game *Game) Move(moveRequest MoveRequest) error {
	game.mutex.Lock()
	defer game.mutex.Unlock()
	move := moveRequest.Move
	piece := game.board[moveRequest.Position.File][moveRequest.Position.Rank]
	err := game.isMoveRequestValid(piece)
	if err != nil {
		return err
	}
	king, enemyKing := game.getKings()
	capturedPiece, err := piece.takeMove(game.board, move, game.previousMove,
		game.previousMover, king, moveRequest.PromoteTo)
	if err != nil {
		return err
	}
	game.handleCapturedPiece(piece, capturedPiece)
	drawByRepetion, err := game.updatePositionHistory()
	if err != nil {
		return err
	}
	drawByInsufficientMaterial := game.isDrawByInsufficientMaterial()
	drawByFiftyMoveRule := game.turnsSinceCaptureOrPawnMove >= 100
	enemyColor := getOppositeColor(piece.color)
	possibleEnemyMoves := AllMoves(
		game.board, enemyColor, move, piece, false, enemyKing,
	)
	if len(possibleEnemyMoves) == 0 &&
		enemyKing.isThreatened(game.board, move, piece) {
		game.gameOver = true
		game.result.Winner = game.turn
	} else if len(possibleEnemyMoves) == 0 || drawByRepetion ||
		drawByFiftyMoveRule || drawByInsufficientMaterial {
		game.gameOver = true
		game.result.Draw = true
	}
	game.previousMove = move
	game.previousMover = piece
	game.turn = enemyColor
	return nil
}

func (game *Game) handleCapturedPiece(piece *Piece, capturedPiece *Piece) {
	if piece.pieceType != Pawn && capturedPiece == nil {
		game.turnsSinceCaptureOrPawnMove++
	} else {
		game.clearPositionHistory()
		game.turnsSinceCaptureOrPawnMove = 0
		if capturedPiece == nil {
			return
		}
		if capturedPiece.color == Black {
			game.blackPieces[capturedPiece.pieceType]--
		} else {
			game.whitePieces[capturedPiece.pieceType]--
		}
	}
}

func (game *Game) isDrawByInsufficientMaterial() bool {
	return (game.OnlyKing(Black) && noMajorPiecesOrPawns(game.whitePieces) &&
		maxOneMinorPiece(game.whitePieces)) ||
		(game.OnlyKing(White) && noMajorPiecesOrPawns(game.blackPieces) &&
			maxOneMinorPiece(game.blackPieces))
}

func noMajorPiecesOrPawns(pieces map[PieceType]uint8) bool {
	return pieces[Queen] == 0 && pieces[Rook] == 0 && pieces[Pawn] == 0
}

func maxOneMinorPiece(pieces map[PieceType]uint8) bool {
	return (pieces[Bishop] == 0 && pieces[Knight] <= 1) ||
		(pieces[Bishop] <= 1 && pieces[Knight] == 0)
}

// OnlyKing returns if the player has only a king
func (game *Game) OnlyKing(color Color) bool {
	pieces := game.blackPieces
	if color == White {
		pieces = game.whitePieces
	}
	return noMajorPiecesOrPawns(pieces) && pieces[Bishop] == 0 &&
		pieces[Knight] == 0
}

func (game *Game) isMoveRequestValid(piece *Piece) error {
	if game.gameOver {
		return errors.New("the game is over")
	} else if piece == nil {
		return errors.New("cannot move nil piece")
	} else if piece.color != game.turn {
		return errors.New("it's not your turn")
	}
	return nil
}

func (game *Game) getKings() (king, enemyKing *Piece) {
	king = game.blackKing
	enemyKing = game.whiteKing
	if game.turn == White {
		king = game.whiteKing
		enemyKing = game.blackKing
	}
	return
}

func (game *Game) updatePositionHistory() (bool, error) {
	position, err := game.MarshalBinary()
	if err != nil {
		return false, err
	}
	game.positionHistory[string(position)]++
	return game.positionHistory[string(position)] > 2, nil
}

func (game *Game) clearPositionHistory() {
	game.positionHistory = make(map[string]uint8)
}

// MarshalBinary represent the game as a byte array
func (game *Game) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, game.turn)
	if err != nil {
		return nil, err
	}
	for _, file := range game.board {
		for _, piece := range file {
			if piece != nil {
				king := game.blackKing
				if piece.color == White {
					king = game.whiteKing
				}
				bytes, err := piece.MarshalBinary(
					game.board, game.previousMove, game.previousMover, king)
				if err != nil {
					return nil, err
				}
				err = binary.Write(buf, binary.BigEndian, bytes)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return buf.Bytes(), nil
}

func getOppositeColor(color Color) (opposite Color) {
	if color == Black {
		opposite = White
	} else {
		opposite = Black
	}
	return opposite
}

// NewGame create a new game
func NewGame() *Game {
	board := newFullBoard()
	return createGame(board)
}

// NewGameNoPawns create a new game with no pawns
func NewGameNoPawns() *Game {
	board := newBoardNoPawns()
	return createGame(board)
}

func createGame(board Board) *Game {
	game := Game{
		board: &board, blackKing: board[4][7], whiteKing: board[4][0],
		positionHistory: make(map[string]uint8),
		blackPieces:     make(map[PieceType]uint8),
		whitePieces:     make(map[PieceType]uint8),
	}
	for _, file := range board {
		for _, piece := range file {
			if piece != nil {
				if piece.color == Black {
					game.blackPieces[piece.pieceType]++
				} else {
					game.whitePieces[piece.pieceType]++
				}
			}
		}
	}
	game.updatePositionHistory()
	game.turn = White
	return &game
}

// BoardString get the board's string
func (game *Game) BoardString() string {
	game.mutex.RLock()
	defer game.mutex.RUnlock()
	return game.board.String()
}

// GetBoard get the game's board
func (game *Game) GetBoard() *Board {
	game.mutex.RLock()
	defer game.mutex.RUnlock()
	return game.board
}

// PointAdvantage get the color's point advantage
func (game *Game) PointAdvantage(color Color) int8 {
	game.mutex.RLock()
	defer game.mutex.RUnlock()
	var points int8 = 0
	for _, file := range game.board {
		for _, piece := range file {
			if piece != nil {
				if piece.Color() == color {
					points += piece.Value()
				} else {
					points -= piece.Value()
				}
			}
		}
	}
	return points
}

// Turn get the current turn
func (game *Game) Turn() Color {
	game.mutex.RLock()
	defer game.mutex.RUnlock()
	return game.turn
}

// GameOver get whether the game is over
func (game *Game) GameOver() bool {
	game.mutex.RLock()
	defer game.mutex.RUnlock()
	return game.gameOver
}

// SetGameResult set the game's result
func (game *Game) SetGameResult(winner Color, draw bool) {
	game.mutex.Lock()
	defer game.mutex.Unlock()
	game.gameOver = true
	game.result.Winner = winner
	game.result.Draw = draw
}

// Result get the game's result
func (game *Game) Result() GameResult {
	game.mutex.RLock()
	defer game.mutex.RUnlock()
	return game.result
}

func (mr MoveRequest) String() string {
	return "Position: " + mr.Position.String() + ", Move: " + mr.Move.String()
}
