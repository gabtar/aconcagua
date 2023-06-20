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

// Matches the pos.Bitboards order
const (
	WHITE_KING = iota
	WHITE_QUEEN
	WHITE_ROOK
	WHITE_BISHOP
	WHITE_KNIGHT
	WHITE_PAWN
	BLACK_KING
	BLACK_QUEEN
	BLACK_ROOK
	BLACK_BISHOP
	BLACK_KNIGHT
	BLACK_PAWN
	WHITE_KING_ENDGAME
	BLACK_KING_ENDGAME
)

// pieceScoreValue maps a piece with a centipawn score value (using the pos.Bitboards order)
var pieceScoreValue = [12]int{
	20000,
	900,
	500,
	330,
	320,
	100,
	-20000,
	-900,
	-500,
	-330,
	-320,
	-100,
}

var piecesSquareTables [14][64]int = loadPiecesSquareTables()

// Pieces-Square tables for white's pieces (but from black's perspective view) eg -> pawnBonus[0] == pawnBonus["h8"]
var pawnBonus = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	50, 50, 50, 50, 50, 50, 50, 50,
	10, 10, 20, 30, 30, 20, 10, 10,
	5, 5, 10, 25, 25, 10, 5, 5,
	0, 0, 0, 20, 20, 0, 0, 0,
	5, -5, -10, 0, 0, -10, -5, 5,
	5, 10, 10, -20, -20, 10, 10, 5,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightBonus = [64]int{
	-50, -40, -30, -30, -30, -30, -40, -50,
	-40, -20, 0, 0, 0, 0, -20, -40,
	-30, 0, 10, 15, 15, 10, 0, -30,
	-30, 5, 15, 20, 20, 15, 5, -30,
	-30, 0, 15, 20, 20, 15, 0, -30,
	-30, 5, 10, 15, 15, 10, 5, -30,
	-40, -20, 0, 5, 5, 0, -20, -40,
	-50, -40, -30, -30, -30, -30, -40, -50,
}

var bishopBonus = [64]int{
	-20, -10, -10, -10, -10, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 10, 10, 5, 0, -10,
	-10, 5, 5, 10, 10, 5, 5, -10,
	-10, 0, 10, 10, 10, 10, 0, -10,
	-10, 10, 10, 10, 10, 10, 10, -10,
	-10, 5, 0, 0, 0, 0, 5, -10,
	-20, -10, -10, -10, -10, -10, -10, -20,
}

var rookBonus = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	5, 10, 10, 10, 10, 10, 10, 5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	-5, 0, 0, 0, 0, 0, 0, -5,
	0, 0, 0, 5, 5, 0, 0, 0,
}

var queenBonus = [64]int{
	-20, -10, -10, -5, -5, -10, -10, -20,
	-10, 0, 0, 0, 0, 0, 0, -10,
	-10, 0, 5, 5, 5, 5, 0, -10,
	-5, 0, 5, 5, 5, 5, 0, -5,
	0, 0, 5, 5, 5, 5, 0, -5,
	-10, 5, 5, 5, 5, 5, 0, -10,
	-10, 0, 5, 0, 0, 0, 0, -10,
	-20, -10, -10, -5, -5, -10, -10, -20,
}

var kingMiddleGameBonus = [64]int{
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
var kingEndGameBonus = [64]int{
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
		return -10 * pieceScoreValue[0]
	}
	if pos.Checkmate(board.BLACK) {
		return 10 * pieceScoreValue[0]
	}

	for pieceType, bb := range pos.Bitboards {
		for bb > 0 {
			sqIndex := board.Bsf(bb)
			sqBB := board.Bitboard(0b1 << sqIndex)
			isEndgame := isEndgame(pos)

			// TODO: Refactor?
			if pieceType == WHITE_KING && isEndgame {
				score += pieceScoreValue[pieceType]
				score += piecesSquareTables[WHITE_KING_ENDGAME][sqIndex]
			} else if pieceType == BLACK_KING && isEndgame {
				score += pieceScoreValue[pieceType]
				score += piecesSquareTables[BLACK_KING_ENDGAME][sqIndex]
			} else {
				score += pieceScoreValue[pieceType]
				score += piecesSquareTables[pieceType][sqIndex]
			}

			bb ^= sqBB
		}
	}

	return
}

// reverses an array of integer
func reverse(arr [64]int) [64]int {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

func invertSign(arr [64]int) [64]int {
	for i := 0; i < 64; i++ {
		arr[i] = -arr[i]
	}
	return arr
}

func loadPiecesSquareTables() (pq [14][64]int) {

	// White pieces
	pq[WHITE_KING] = reverse(kingMiddleGameBonus)
	pq[WHITE_QUEEN] = reverse(queenBonus)
	pq[WHITE_ROOK] = reverse(rookBonus)
	pq[WHITE_BISHOP] = reverse(bishopBonus)
	pq[WHITE_KNIGHT] = reverse(knightBonus)
	pq[WHITE_PAWN] = reverse(pawnBonus)

	// Black pieces
	pq[BLACK_KING] = invertSign(kingMiddleGameBonus)
	pq[BLACK_QUEEN] = invertSign(queenBonus)
	pq[BLACK_ROOK] = invertSign(rookBonus)
	pq[BLACK_BISHOP] = invertSign(bishopBonus)
	pq[BLACK_KNIGHT] = invertSign(knightBonus)
	pq[BLACK_PAWN] = invertSign(pawnBonus)

	// King endgame bonus
	pq[WHITE_KING_ENDGAME] = reverse(kingEndGameBonus)
	pq[BLACK_KING_ENDGAME] = invertSign(kingEndGameBonus)

	return
}

func isEndgame(pos board.Position) bool {
	// Both sides have no queens or
	// TODO: Every side which has a queen has additionally no other pieces or one minorpiece maximum.
	if pos.Bitboards[WHITE_QUEEN] == 0 && pos.Bitboards[BLACK_QUEEN] == 0 {
		return true
	}

	return false
}
