package search

import (
	"math"
	"sort"

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
	alphaOrig := alpha

	// Check if it's present on the transposition table and return the score
	if ttEntry, exists := tt.table[pos.Zobrist]; exists && tt.table[pos.Zobrist].depth >= depth {
		if ttEntry.flag == EXACT {
			return ttEntry.score
		} else if ttEntry.flag == LOWERBOUND {
			alpha = max(alpha, ttEntry.score)
		} else if ttEntry.flag == UPPERBOUND {
			beta = min(beta, ttEntry.score)
		}

		if alpha >= beta {
			return ttEntry.score
		}
	}

	if depth == 0 || pos.Checkmate(board.White) || pos.Checkmate(board.Black) {
		// Color modifier for evaluation
		color := 1
		if pos.ToMove() == board.Black {
			color = -1
		}

		score = color * evaluation.Evaluate(pos)
		return
	}

	score = math.MinInt

	for _, move := range sortMoves(pos.LegalMoves(pos.ToMove()), bestMove) {
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

	if score <= alphaOrig {
		tt.save(pos.Zobrist, depth, score, UPPERBOUND)
	} else if score >= beta {
		tt.save(pos.Zobrist, depth, score, LOWERBOUND)
	} else {
		tt.save(pos.Zobrist, depth, score, EXACT)
	}

	return
}

// max returns the maximum value of the 2 integers passed
func max(a int, b int) int {
	if a >= b {
		return a
	}
	return b
}

// min returns the minimum value of the 2 integers passed
func min(a int, b int) int {
	if b <= a {
		return b
	}
	return a
}

// BestMove returns the best move sequence in the position (for the current side)
// with its score evaluation.
func BestMove(pos *board.Position, maxDepth int) (bestMoveScore int, bestMove board.Move) {

	min := math.MinInt + 1 // NOTE: Due to overflow issues i need to use this way, now negamax works well
	max := math.MaxInt
	tt = newTranspositionTable()

	// TODO: iterative deepening
	for d := 1; d <= maxDepth; d++ {
		bestMoveScore = negamax(*pos, d, d, min, max, &bestMove)
	}

	return
}

func sortMoves(legalMoves []board.Move, bestMove *board.Move) (sortedMoves []board.Move) {
	sort.Slice(legalMoves, func(i, j int) bool {
		// TODO: improve move ordering by using a principal variation
		// this only works on 1st iteration
		if legalMoves[i] == *bestMove {
			return true
		}

		return legalMoves[i].MoveType() > legalMoves[j].MoveType()
	})
	return legalMoves
}
