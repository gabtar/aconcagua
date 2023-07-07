package search

import (
	"math"

	"github.com/gabtar/aconcagua/board"
	"github.com/gabtar/aconcagua/evaluation"
)

// Pseudocode from wikipedia
// function negamax(node, depth, α, β, color) is
//     alphaOrig := α
//
//     (* Transposition Table Lookup; node is the lookup key for ttEntry *)
//     ttEntry := transpositionTableLookup(node)
//     if ttEntry is valid and ttEntry.depth ≥ depth then
//         if ttEntry.flag = EXACT then
//             return ttEntry.value
//         else if ttEntry.flag = LOWERBOUND then
//             α := max(α, ttEntry.value)
//         else if ttEntry.flag = UPPERBOUND then
//             β := min(β, ttEntry.value)
//
//         if α ≥ β then
//             return ttEntry.value
//
//     if depth = 0 or node is a terminal node then
//         return color × the heuristic value of node
//
//     childNodes := generateMoves(node)
//     childNodes := orderMoves(childNodes)
//     value := −∞
//     for each child in childNodes do
//         value := max(value, −negamax(child, depth − 1, −β, −α, −color))
//         α := max(α, value)
//         if α ≥ β then
//             break
//
//     (* Transposition Table Store; node is the lookup key for ttEntry *)
//     ttEntry.value := value
//     if value ≤ alphaOrig then
//         ttEntry.flag := UPPERBOUND
//     else if value ≥ β then
//         ttEntry.flag := LOWERBOUND
//     else
//         ttEntry.flag := EXACT
//     ttEntry.depth := depth
//     ttEntry.is_valid := true
//     transpositionTableStore(node, ttEntry)
//
//     return value

// (* Initial call for Player A's root node *)
// negamax(rootNode, depth, −∞, +∞, 1)

// negamax returns the score of the best posible move by the evaluation function
// for a fixed depth
func negamax(pos board.Position, depth int, initialDepth int, alpha int, beta int, bestMove *board.Move) (score int) {

	// Check if it's present on the transposition table and return the score
	// TODO: Implement flags on transpositionTable
	if ttScore, exists := tt.find(pos.Zobrist); exists {
		return ttScore
	}

	if depth == 0 || pos.Checkmate(board.WHITE) || pos.Checkmate(board.BLACK) {
		// Color modifier for evaluation
		color := 1
		if pos.ToMove() == board.BLACK {
			color = -1
		}

		score = color * evaluation.Evaluate(pos)
		return
	}

	score = math.MinInt

	for _, move := range pos.LegalMoves(pos.ToMove()) {
		newPos := pos.MakeMove(&move)
		newScore := -negamax(newPos, depth-1, initialDepth, -beta, -alpha, bestMove)

		if newScore > score {
			score = newScore
			// Set best move (only for first level)
			if depth == initialDepth {
				// TODO: / IDEA I should use a pointer to an engine struct eg.
				*bestMove = move
			}
		}
		if newScore > alpha {
			alpha = newScore
		}
		if alpha >= beta {
			return
		}
	}

	// Store the score in the transposition table
	if score <= alpha {
		tt.save(pos.Zobrist, depth, score)
	} else if score >= beta {
		tt.save(pos.Zobrist, depth, score)
	}

	return
}

// BestMove returns the best move sequence in the position (for the current side)
// with its score evaluation.
func BestMove(pos *board.Position, depth int) (bestMoveScore int, bestMove board.Move) {

	min := math.MinInt + 1 // NOTE: Due to overflow issues i need to use this way, now negamax works well
	max := math.MaxInt

	tt = newTranspositionTable()
	bestMoveScore = negamax(*pos, depth, depth, min, max, &bestMove)

	return
}
