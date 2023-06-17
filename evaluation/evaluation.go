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

const (
	KING   int = 0
	QUEEN  int = 1
	ROOK   int = 2
	BISHOP int = 3
	KNIGHT int = 4
	PAWN   int = 5
)

// pieceScoreMap maps a pieceType with a centipawn score value
var pieceScoreMap = []int{
	20000, // King
	900,   // Queen
	500,   // Rook
	330,   // Bishop
	320,   // Knight
	100,   // Pawn
}

// Pieces-Square tables (from black's perspective view) eg -> pawnBonus[0] == pawnBonus["h8"]
// Aditional bonus/penalties based on the piece location in the board
var pawnBonus = []int{
	0, 0, 0, 0, 0, 0, 0, 0,
	50, 50, 50, 50, 50, 50, 50, 50,
	10, 10, 20, 30, 30, 20, 10, 10,
	5, 5, 10, 25, 25, 10, 5, 5,
	0, 0, 0, 20, 20, 0, 0, 0,
	5, -5, -10, 0, 0, -10, -5, 5,
	5, 10, 10, -20, -20, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightBonus = []int{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, 0, 0, 0, 0, -20, -40,
	-30, 0, 10, 15, 15, 10, 0, -30,
	-30, 5, 15, 20, 20, 15, 5, -30,
	-30, 0, 15, 20, 20, 15, 0, -30,
	-30, 5, 10, 15, 15, 10, 5, -30,
	-40, -20, 0, 5, 5, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}

var bishopBonus = []int{
	-20, -10, -10, -10, -10, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 10, 10, 5, 0, -10,
	-10, 5, 5, 10, 10, 5, 5, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 5, 0, 0, 0, 0, 5, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}

var rookBonus = []int{
	0, 0, 0, 0, 0, 0, 0, 0,
	5, 10, 10, 10, 10, 10, 10, 5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	0, 0, 0, 5, 5, 0, 0, 0,
}

var queenBonus = []int{
	-20, -10, -10, -5, -5, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-5, 0, 5, 5, 5, 5, 0, -5,
	0, 0, 5, 5, 5, 5, 0, -5,
	-10, 5, 5, 5, 5, 5, 0, -10,
	-10, 0, 5, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}

var kingMiddleGameBonus = []int{
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-30, -40, -40, -50, -50, -40, -40, -30,
	-20, -30, -30, -40, -40, -30, -30, -20,
	-10, -20, -20, -20, -20, -20, -20, -10,
	20, 20, 0, 0, 0, 0, 20, 20,
	20, 30, 10, 0, 0, 10, 30, 20,
}

// Endgame criteria
// Both sides have no queens or
// Every side which has a queen has additionally no other pieces or one minorpiece maximum.
var kingEndGameBonus = []int{
	-50, -40, -30, -20, -20, -30, -40, -50,
	-30, -20, -10, 0, 0, -10, -20, -30,
	-30, -10, 20, 30, 30, 20, -10, -30,
	-30, -10, 30, 40, 40, 30, -10, -30,
	-30, -10, 30, 40, 40, 30, -10, -30,
	-30, -10, 20, 30, 30, 20, -10, -30,
	-30, -30, 0, 0, 0, 0, -30, -30,
	-50, -30, -30, -30, -30, -30, -30, -50,
}

// Evaluate scores out the position in centipawns for a given a position of the board
func Evaluate(pos board.Position) (score int) {
	// Set checkmate as 10 kings value
	// TODO: should evaluate checkmate in number of moves. eg mate in 2 is better than a mate in 3
	if pos.Checkmate(board.WHITE) {
		return -10 * pieceScoreMap[0]
	}
	if pos.Checkmate(board.BLACK) {
		return 10 * pieceScoreMap[0]
	}

	// Sum the score of the pieces of the side and then subtract them
	whiteMaterial := calculateScore(pos, board.WHITE)
	blackMaterial := calculateScore(pos, board.BLACK)

	score = (whiteMaterial - blackMaterial)
	return
}

// calculateScore returns the material count of each side
func calculateScore(pos board.Position, side rune) (score int) {
	bitboards := pos.Bitboards(side)

	score += kingScore(bitboards[0], side, pos.IsEndgame())
	score += queenScore(bitboards[1], side)
	score += rookScore(bitboards[2], side)
	score += bishopScore(bitboards[3], side)
	score += knightScore(bitboards[4], side)
	score += pawnScore(bitboards[5], side)
	return
}

// HELPERS that calculates the score of the passed pieces

func kingScore(bitboard board.Bitboard, side rune, endGame bool) (score int) {
	// Select proper function
	bitScanFunction := getBitscanFunction(side)
	score += pieceScoreMap[KING]

	if bitboard > 0 {
		if endGame {
			score += kingEndGameBonus[bitScanFunction(bitboard)]
		} else {
			score += kingMiddleGameBonus[bitScanFunction(bitboard)]
		}
	}
	return
}

func queenScore(bitboard board.Bitboard, side rune) (score int) {
	bitScanFunction := getBitscanFunction(side)

	for bitboard > 0 {
		queen := board.BitboardFromIndex(board.Bsf(bitboard))
		score += pieceScoreMap[QUEEN]
		score += queenBonus[bitScanFunction(queen)]

		bitboard &= ^queen
	}

	return
}

func rookScore(bitboard board.Bitboard, side rune) (score int) {
	bitScanFunction := getBitscanFunction(side)

	for bitboard > 0 {
		rook := board.BitboardFromIndex(board.Bsf(bitboard))
		score += pieceScoreMap[ROOK]
		score += rookBonus[bitScanFunction(rook)]

		bitboard &= ^rook
	}

	return
}

func bishopScore(bitboard board.Bitboard, side rune) (score int) {
	bitScanFunction := getBitscanFunction(side)

	for bitboard > 0 {
		bishop := board.BitboardFromIndex(board.Bsf(bitboard))
		score += pieceScoreMap[BISHOP]
		score += bishopBonus[bitScanFunction(bishop)]

		bitboard &= ^bishop
	}

	return
}

func knightScore(bitboard board.Bitboard, side rune) (score int) {
	bitScanFunction := getBitscanFunction(side)

	for bitboard > 0 {
		knight := board.BitboardFromIndex(board.Bsf(bitboard))
		score += pieceScoreMap[KNIGHT]
		score += knightBonus[bitScanFunction(knight)]

		bitboard &= ^knight
	}

	return
}

func pawnScore(bitboard board.Bitboard, side rune) (score int) {
	bitScanFunction := getBitscanFunction(side)

	for bitboard > 0 {
		pawn := board.BitboardFromIndex(board.Bsf(bitboard))
		score += pieceScoreMap[PAWN]
		score += pawnBonus[bitScanFunction(pawn)]

		bitboard &= ^pawn
	}
	return
}

// getBitscanFunction returns a bitscan function depending on which side are we
// checking the score
func getBitscanFunction(side rune) func(bitboard board.Bitboard) int {
	bitScanFunction := board.Bsr
	if side == board.BLACK {
		bitScanFunction = board.Bsf
	}
	return bitScanFunction
}
