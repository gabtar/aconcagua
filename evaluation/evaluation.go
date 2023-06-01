// evaluation contains all functions related to position evaluation in the engine
package evaluation

import (
	"github.com/gabtar/aconcagua/board"
)

// Base Centipawn score values
// P = 100
// N = 320
// B = 330
// R = 500
// Q = 900
// K = 20000

// pieceScoreMap maps a pieceType with a centipawn score value
var pieceScoreMap = map[int]float64{
	board.WHITE_KING:   20000.0,
	board.BLACK_KING:   20000.0,
	board.WHITE_QUEEN:  900.0,
	board.BLACK_QUEEN:  900.0,
	board.WHITE_ROOK:   500.0,
	board.BLACK_ROOK:   500.0,
	board.WHITE_BISHOP: 330.0,
	board.BLACK_BISHOP: 330.0,
	board.WHITE_KNIGHT: 320.0,
	board.BLACK_KNIGHT: 320.0,
	board.WHITE_PAWN:   100.0,
	board.BLACK_PAWN:   100.0,
}

// Evaluate scores out the position in centipawns for a given a position of the board
// TODO: Add piece value based on piece-square tables
func Evaluate(pos board.Position) (score float64) {
	// Set checkmate as 10 kings value
	// TODO: should evaluate checkmate in number of moves. eg mate in 2 is better than a mate in 3
	if pos.Checkmate(board.WHITE) {
		return -10 * pieceScoreMap[board.WHITE_KING]
	}
	if pos.Checkmate(board.BLACK) {
		return 10 * pieceScoreMap[board.BLACK_KING]
	}

	// Sum the score of the pieces of the side and then subtract them
	whiteMaterial := calculateMaterial(pos, board.WHITE)
	blackMaterial := calculateMaterial(pos, board.BLACK)

	score = (whiteMaterial - blackMaterial)
	return
}

// calculateMaterial returns the material count of each side
func calculateMaterial(pos board.Position, side rune) (material float64) {
	pieces := []int{board.WHITE_KING, board.WHITE_QUEEN, board.WHITE_ROOK, board.WHITE_BISHOP, board.WHITE_KNIGHT, board.WHITE_PAWN}
	if side != board.WHITE {
		pieces = []int{board.BLACK_KING, board.BLACK_QUEEN, board.BLACK_ROOK, board.BLACK_BISHOP, board.BLACK_KNIGHT, board.BLACK_PAWN}
	}

	for _, piece := range pieces {
		material += pieceScoreMap[piece] * float64(pos.NumberOfPieces(piece))
	}
	return
}
