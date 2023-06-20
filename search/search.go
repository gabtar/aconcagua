package search

import (
	"math"

	"github.com/gabtar/aconcagua/board"
	"github.com/gabtar/aconcagua/evaluation"
)

// Pseudocode from wikipedia
// function negamax(node, depth, α, β, color) is
//     if depth = 0 or node is a terminal node then
//         return color × the heuristic value of node
//
//     childNodes := generateMoves(node)
//     childNodes := orderMoves(childNodes)
//     value := −∞
//     foreach child in childNodes do
//         value := max(value, −negamax(child, depth − 1, −β, −α, −color))
//         α := max(α, value)
//         if α ≥ β then
//             break (* cut-off *)
//     return value

// (* Initial call for Player A's root node *)
// negamax(rootNode, depth, −∞, +∞, 1)

// negamax returns the score of the best posible move by the evaluation function
// for a fixed depth
func negamax(pos board.Position, depth int, maxDepth int, alpha int, beta int, bestMove *board.Move) (score int) {
	// FIX: not working properly -> check with queen sac test!
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
		newScore := -negamax(newPos, depth-1, maxDepth, -beta, -alpha, bestMove)

		if newScore > score {
			score = newScore
			// Set best move (only for first level)
			if depth == maxDepth {
				// TODO: / IDEA I should use a pointer to an engine struct eg.
				// type engine struct {
				//      bestMove Move
				//      currentDepth int
				//      *other useful fields*
				// }
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

	return
}

// minmax returns the bestMove score and the move sequence
func minmax(pos board.Position, depth int, alpha int, beta int, moveTrace []board.Move) (score int, moves []board.Move) {
	if depth == 0 || pos.Checkmate(board.WHITE) || pos.Checkmate(board.BLACK) {
		score = evaluation.Evaluate(pos)
		moves = moveTrace
		return
	}

	if pos.ToMove() == board.WHITE {
		score = math.MinInt

		for _, move := range pos.LegalMoves(pos.ToMove()) {
			newPos := pos.MakeMove(&move)
			newMoves := append(moveTrace, move)
			newScore, newMoveTrace := minmax(newPos, depth-1, alpha, beta, newMoves)
			if newScore > score {
				score = newScore
				moves = newMoveTrace
			}
			if score > alpha {
				alpha = score
			}
			if beta <= alpha {
				break
			}
		}
	} else {
		score = math.MaxInt

		for _, move := range pos.LegalMoves(pos.ToMove()) {
			newPos := pos.MakeMove(&move)
			newMoves := append(moveTrace, move)
			newScore, newMoveTrace := minmax(newPos, depth-1, alpha, beta, newMoves)
			if newScore < score {
				score = newScore
				moves = newMoveTrace
			}
			if score < beta {
				beta = score
			}
			if beta <= alpha {
				break
			}
		}
	}
	return
}

// BestMove returns the best move sequence in the position (for the current side)
// with its score evaluation.
func BestMove(pos *board.Position, depth int) (bestMoveScore int, bestMove board.Move) {

	// bestMoveScore = negamax(*pos, depth, depth, math.MinInt, math.MaxInt, &bestMove)

	var mvs []board.Move
	bestMoveScore, bm := minmax(*pos, depth, math.MinInt, math.MaxInt, mvs)
	bestMove = bm[0]

	return
}
